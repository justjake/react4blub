#include <stdlib.h>
#include <string.h>

#include "array_list.h"
#include "sqlite/sqlite3.h"
#include "useFunctionHook.h"

// ---------------------------------------------------------------------------------------
// Types: Node, Props, Children

typedef struct RChildren RChildren;
typedef struct RComponent RComponent;
typedef struct RNode RNode;
struct RChildren {
  int length;
  RNode *children;
};
typedef struct RProps {
  RChildren children;
} RProps;

struct RNode {
  RComponent *comp;
  RProps *props;
  char *key;
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

// TODO: this sucks, delete it
typedef union {
  int as_int;
  // TODO: string?
  // TODO: tag needed?
  void *as_ptr;
} RDep;

typedef struct RRef {
  void *current;
} RRef;

typedef struct RState {
  const void *state;
  const void (*set)(void *nextState);
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
  int init;
  union {
    RRef ref;
    RCallback callback;
    RMemo memo;
  } state;
} RHookInstance;

typedef struct RReactFiber {
  // Relationships.
  struct RReactFiber *parent;
  // Note: ReactJS uses linked lists for this
  ArrayList(struct RReactFiber *) children;

  // State.
  void *mounted;  // !=0 -> mounted
  int dirty;      // !=0 -> needs re-render
  char *error;    // !=0 -> errored
  RHookInstance *first_hook;
  RHookInstance *next_hook;
  RHookInstance *prev_hook;
} RReactFiber;

static void fiber_start(RReactFiber *fiber) {
  fiber->dirty = 0;
  fiber->error = 0;
  fiber->next_hook = fiber->first_hook;
  fiber->prev_hook = fiber->first_hook;
}

typedef struct RReactRoot {
  sqlite3 *sqlite3;
  RReactFiber *root_fiber;
} RReactRoot;

static char *ERROR_HOOK_CREATE_AFTER_MOUNT = "Error: cannot create a hook after component mount";
static char *ERROR_HOOK_TYPE_MISMATCH = "Error: hook order didn't match";

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

static RReactFiber *currentFiber = 0;

RRef *useRef() {
  RHookInstance *hook = fiber_get_or_create_hook(currentFiber, RHookUseRef);
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
  RHookInstance *hook = fiber_get_or_create_hook(currentFiber, RHookUseCallback);
  if (deps_update(&(hook->state.callback.deps), hook->init, size_of_struct, args_struct)) {
    hook->init = 1;
    hook->state.callback.fp = bound_fp;
  }
  return hook->state.callback;
}

typedef void *(RMemoFunction)(void *);

void *useMemoRaw(RMemoFunction *fp, size_t args_size, void *args) {
  RHookInstance *hook = fiber_get_or_create_hook(currentFiber, RHookUseMemo);
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

// ---------------------------------------------------------------------------------------
// Sketch

typedef struct {
  RProps base;
  int initialCount;
  int incrBy;
} CounterProps;

// RNode *Counter(CounterProps props) {
//   RState count = useState(int, props.initialCount);
//   int incr_by = props.incrBy != 0 ? props.incrBy : 1;
//   incr = useCallback(&Counter_incr, RState, count, int, incr_by);

//   return (
//       h(div,
//         text("Current count: %d, ", count),
//         h(div,
//           (DivProps){
//               .onClick = incr},
//           text("Incr by %d", incr_by)));
// }

defineCallback(Counter_incr, (RState, count_state), (int, increment)) {
  count_state.set(count_state.state + increment);
}

// RNode *App(_props RProps) {
//   return (
//     h(&Counter, (CounterProps){{}, 0, 1}
//   )
// }
