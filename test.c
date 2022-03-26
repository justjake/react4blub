#include "sqlite/sqlite3.h"
#include "useCallback.h"

// ---------------------------------------------------------------------------------------
// Types: Component

typedef enum RComponentType {
  RComponentTypeFunction,
  RComponentTypeTag,
} RComponentType;

typedef struct RComponent {
  RComponentType type;
  union {
    RTag tag;
    RComponentFunction *fp;
  } comp;
} RComponent;

typedef enum RTag {
  RTagText,
  RTagFragment,
  RTagDiv,
} RTag;

typedef RNode RComponentFunction(RProps props);

// ---------------------------------------------------------------------------------------
// Types: Node, Props, Children

typedef struct RProps {
  RChildren children;
} RProps;

typedef struct RNode {
  RComponent *comp;
  RProps *props;
  char *key;
} RNode;

typedef struct RChildren {
  int length;
  RNode *children;
} RChildren;

// ---------------------------------------------------------------------------------------
// Hook Types

typedef union {
  int as_int;
  // TODO: string?
  // TODO: tag needed?
  void *as_ptr;
} RDep;

typedef struct RHookDeps {
  int length;
  RDep **deps;
} RHookDeps;

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

typedef struct RReactRoot {
  sqlite3 *sqlite3;
} RReactRoot;

typedef struct RReactFiber {
  void *mounted;
  int hooks_c;
  RReactHookState *hooks;
  // Note: ReactJS uses linked lists for this
  int children_c;
  RReactFiber **children;
} RReactFiber;

typedef enum RHookType {
  RHookUseRef,
  RHookUseState,
  RHookUseMemo,
} RHookType;

typedef struct RReactHookState {
  RHookType type;
  union {
    RRef ref;
    RRef state;
    struct {
      void *prev;
      RHookDeps deps;
    } memo;
  } state;
} RReactHookState;

// Sketch

struct CounterProps {
  RProps base;
  int initialCount;
  int incrBy;
}

RNode *
Counter(CounterProps props) {
  RState count = useState(int, props.initialCount);
  int incr_by = props.incrBy != 0 ? props.incrBy : 1;
  incr = useCallback(&Counter_incr, RState, count, int, incr_by);
});

return (
    h(div,
      text("Current count: %d, ", count),
      h(div,
        (DivProps){
            .onClick = incr},
        text("Incr by %d", incr_by))));
}

defineCallback(Counter_incr, (RState, count_state), (int, increment)) {
  count.set(count.state + incr_by);
}

RNode *App(_props RProps) {
  return (
    h(&Counter, (CounterProps){{}, 0, 1}
  )
}
