#include <stdarg.h>
#include <stdlib.h>
#include <string.h>

#include "array_list.h"
#include "intrusive_list.h"
#include "macro_length.h"
#include "sqlite/sqlite3.h"
#include "useFunctionHook.h"

// ---------------------------------------------------------------------------------------
// Types: Node, Props, Children

typedef struct RChildren RChildren;
typedef struct RComponent RComponent;
typedef struct RNode RNode;

typedef struct RProps {
} RProps;

struct RNode {
  RComponent *comp;
  char *key;
  // TODO: ref?
  RProps *props;
  unsigned children_count;
  RNode *children[];
};

// ---------------------------------------------------------------------------------------
// Types: Component

typedef enum RTag {
  RTagText,
  RTagFragment,
  RTagDiv,
} RTag;

typedef enum RComponentType {
  RComponentTypeFunction,
  RComponentTypeTag,
} RComponentType;

typedef RNode (*RComponentFunction)(RProps);

struct RComponent {
  RComponentType type;
  union {
    RTag tag;
    RComponentFunction *fp;
  } comp;
};

// ---------------------------------------------------------------------------------------
// Hook Types

typedef struct RReactFiber RReactFiber;
typedef struct RReactRoot RReactRoot;

typedef struct RRef {
  void *current;
} RRef;

typedef struct RState {
  const void *state;
  const size_t size;
  const void *_internal;
} RState;

#define State(T) RState;

// ---------------------------------------------------------------------------------------
// Internal Types

typedef struct RRetainedDeps {
  size_t size;
  void *ptr;
} RRetainedDeps;

typedef struct RCallback {
  RRetainedDeps deps;
  void (*fp)(void *);
} RCallback;

typedef struct RMemo {
  RRetainedDeps deps;
  void *prev;
} RMemo;

typedef enum RHookType {
  RHookUseRef,
  RHookUseCallback,
  RHookUseMemo,
  RHookUseState,
} RHookType;

typedef struct RHookInstance {
  RHookType type;
  struct RHookInstance *next;
  RReactFiber *fiber;
  int init;
  union {
    RRef ref;
    RCallback callback;
    RMemo memo;
  } state;
} RHookInstance;

struct RReactFiber {
  // Relationships.
  RReactRoot *root;
  struct RReactFiber *parent;
  struct RReactFiber *next_dirty;
  // Note: ReactJS uses linked lists for this
  ArrayList(struct RReactFiber *) children;

  // State.
  void *mounted;    // !=0 -> mounted
  RNode *rendered;  // !=0 -> rendered
  int dirty;        // !=0 -> needs re-render
  char *error;      // !=0 -> errored
  RHookInstance *first_hook;
  RHookInstance *next_hook;
  RHookInstance *prev_hook;
};

static void fiber_start(RReactFiber *fiber) {
  fiber->dirty = 0;
  fiber->error = 0;
  fiber->next_hook = fiber->first_hook;
  fiber->prev_hook = fiber->first_hook;
}

struct RReactRoot {
  // Target
  sqlite3 *sqlite3;  // Render into DB

  RReactFiber *root_fiber;
  int dirty;
  RReactFiber *last_dirty;
  RReactFiber *first_dirty;
};

struct RRenderPath {
  ListMember(struct RRenderPath, node);
  RNode *node;
};

void root_render(RReactRoot *root) {
  if (!root->sqlite3) {
    sqlite3_open_v2(":memory:", &(root->sqlite3), SQLITE_OPEN_READWRITE | SQLITE_OPEN_MEMORY | SQLITE_OPEN_CREATE, 0);
  }
}

void fiber_schedule_render(RReactFiber *fiber) {
  if (fiber->dirty) {
    return;
  }

  fiber->dirty = 1;
  fiber->root->dirty = 1;
  if (fiber->root->last_dirty) {
    fiber->root->last_dirty->next_dirty = fiber;
    fiber->root->last_dirty = fiber;
  } else {
    fiber->root->first_dirty = fiber;
    fiber->root->last_dirty = fiber;
  }

  // if (!global_state_is_batching()) {
  root_render(fiber->root);
  // } else {
  // globalState.next_root = fiber->root;
  // }
}

static const char *ERROR_HOOK_CREATE_AFTER_MOUNT = "Error: cannot create a hook after component mount";
static const char *ERROR_HOOK_TYPE_MISMATCH = "Error: hook order didn't match";

static RHookInstance *fiber_get_or_create_hook(RReactFiber *fiber, RHookType tag) {
  RHookInstance *hook = fiber->next_hook;
  if (hook == 0) {
    if (!fiber->mounted) {
      // Init new hook
      hook = malloc(sizeof(RHookInstance));
      hook->type = tag;
      hook->next = 0;
      hook->init = 0;

      if (fiber->prev_hook) {
        fiber->prev_hook->next = hook;
      } else {
        fiber->first_hook = hook;
      }
      fiber->prev_hook = hook;
      return hook;
    }

    // Requesting a hook, but we're already mounted.
    // Can't create a new hook.
    fiber->error = ERROR_HOOK_CREATE_AFTER_MOUNT;
    fiber->prev_hook = hook;
    return 0;
  }

  if (hook->type != tag) {
    fiber->error = ERROR_HOOK_TYPE_MISMATCH;
    fiber->prev_hook = hook;
    return 0;
  }

  fiber->prev_hook = hook;
  return hook;
}

static struct {
  RReactRoot *currentRoot;
  RReactFiber *currentFiber;
  void *currentEvent;
  ListOwner(RReactRoot, pending_render);
} globalState;

int global_state_is_rendering() {
  if (globalState.currentFiber || globalState.currentRoot) {
    return 1;
  }

  return 0;
}

int global_state_is_batching() {
  return (global_state_is_batching() || globalState.currentEvent);
}

static struct {
  void *event
} currentEvent = {0};

RRef *useRef() {
  RHookInstance *hook = fiber_get_or_create_hook(globalState.currentFiber, RHookUseRef);
  if (!hook->init) {
    hook->init = 1;
    hook->state.ref.current = 0;
  }
  return &(hook->state.ref);
}

/**
 * @brief Update deps
 *
 * @param deps
 * @param init
 * @return int 1 if deps updated, 0 if same.
 */
int deps_update(RRetainedDeps *deps, int already_init, size_t next_size, void *next_data) {
  if (already_init && deps->size == next_size && memcmp(deps->ptr, next_data, next_size) == 0) {
    return 0;
  }

  if (!already_init) {
    deps->size = next_size;
    deps->ptr = malloc(next_size);
  } else if (deps->size != next_size) {
    free(deps->ptr);
    deps->size = next_size;
    deps->ptr = malloc(next_size);
  }

  memcpy(deps->ptr, next_data, next_size);
  return 1;
}

RCallback useCallbackRaw(void *bound_fp, int size_of_struct, void *args_struct, RHookResultAlloc *_alloc) {
  RHookInstance *hook = fiber_get_or_create_hook(globalState.currentFiber, RHookUseCallback);
  if (deps_update(&(hook->state.callback.deps), hook->init, size_of_struct, args_struct)) {
    hook->init = 1;
    hook->state.callback.fp = bound_fp;
  }
  return hook->state.callback;
}

void invokeCallback(RCallback cb) {
  return cb.fp(cb.deps.ptr);
}

typedef void *(RMemoFunction)(void *);

void *useMemoRaw(RMemoFunction *fp, size_t args_size, void *args) {
  RHookInstance *hook = fiber_get_or_create_hook(globalState.currentFiber, RHookUseMemo);
  if (deps_update(&(hook->state.callback.deps), hook->init, args_size, args)) {
    if (hook->init) {
      free(hook->state.memo.prev);
    } else {
      hook->init = 1;
    }
    hook->state.memo.prev = fp(args);
  }
  return hook->state.memo.prev;
}

#define useState(T, defaultValue) \
  useStateRaw(sizeof(T), &defaultValue);

RState useStateRaw(size_t size, void *default_value) {
  RHookInstance *hook = fiber_get_or_create_hook(globalState.currentFiber, RHookUseState);
  if (!hook->init) {
    hook->init = 1;
    hook->state.ref.current = malloc(size);
    // TODO: make state less mutable?
    // TODO: pass state by value?
    memcpy(hook->state.ref.current, default_value, size);
  }
  return (RState){
      hook->state.ref.current,
      hook,
  };
}

void setState(RState state, void *next_state) {
  RHookInstance *hook = state._internal;
  RReactFiber *fiber = hook->fiber;
  if (!fiber->mounted) {
    // Prepare to segfault
    fiber = 0;
    hook = 0;
  }
  memcpy(state.state, next_state, state.size);
  fiber_schedule_render(fiber);
}

// ---------------------------------------------------------------------------------------
// Sketch

typedef struct DivProps DivProps;
struct DivProps {
  char *id;
  char *className;
  RCallback onClick;
};

typedef struct CounterProps {
  RProps base;
  int initialCount;
  int incrBy;
} CounterProps;

RNode createNode(size_t sizeof_props, const void *given_props, unsigned children_length, va_list args) {
  void *props = malloc(sizeof_props);
  memcpy(props, given_props, sizeof_props);

  RChildren[]
}

#define COMPONENT_node_args(propType, propArg) \
  propType props

#define COMPONENT_fn_args(propType, propArg) \
  propType propArg

#define COMPONENT_prop_typedef(propType, propArg) \
  typedef propType __##ComponentName##_props

#define defineComponent(ComponentName, PropsTuple)                                       \
  COMPONENT_prop_typedef PropsTuple;                                                     \
  RNode ComponentName##_node(COMPONENT_node_args PropsTuple, int children_length, ...) { \
    va_list args;                                                                        \
    va_start(args, children_length);                                                     \
    RNode result = createNode(sizeof(props), &props, children_length, args);             \
    va_end(args);                                                                        \
    node_set_function(&result, &ComponentName);                                          \
    return result;                                                                       \
  };                                                                                     \
  RNode ComponentName(Component_fn_args(PropsTuple))

#define defineTagComponent(ComponentName, Tag, propType)                     \
  typedef propType __##ComponentName##_props;                                \
  RNode ComponentName##_node(propType props, int children_length, ...) {     \
    va_list args;                                                            \
    va_start(args, children_length);                                         \
    RNode result = createNode(sizeof(props), &props, children_length, args); \
    va_end(args);                                                            \
    node_set_tag(&result, Tag);                                              \
    return result;                                                           \
  };

#define h(ComponentName, props, ...) \
  ComponentName##_node((__##ComponentName##_props)props, NUMARGS(__VA_ARGS__) __VA_OPT__(, ) __VA_ARGS__)

defineTagComponent(div, RTagDiv, DivProps);

defineComponent(Counter, (CounterProps, props)) {
  RState count = useState(int, props.initialCount);
  int incr_by = props.incrBy != 0 ? props.incrBy : 1;
  RCallback incr = useCallback(Counter_incr, count, incr_by);

  return (
      h(div, {},
        text("Current count: %d, ", count),
        h(div,
          (DivProps){
              .onClick = incr},
          text("Incr by %d", incr_by))));
}

defineCallback(Counter_incr, (RState, count_state), (int, increment)) {
  int count = increment + *(int *)count_state.state;
  setState(count_state, &count);
}

// RNode *App(_props RProps) {
//   return (
//     h(&Counter, (CounterProps){{}, 0, 1}
//   )
// }
