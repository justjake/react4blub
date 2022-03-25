#include "sqlite/sqlite3.h"

// TODO: JSX???
// #define <|(comp, props, children)
// #define |>

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
}

typedef enum RTag {
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
} RNode;

typedef struct RChildren {
  size_t length;
  children *RNode;
} RChildren;
