#include "macro_map.h"

#define CALLBACK_struct_type(typename, var) typename var;
#define CALLBACK_fn_arg(typename, var) typename var
#define CALLBACK_struct_initialize(typename, var) var,
#define CALLBACK_struct_arg(typename, var) args.var

#define defineCallback(callback_name, ...)                           \
  void callback_name(MAP_TUPLES_LIST(CALLBACK_fn_arg, __VA_ARGS__)); \
  struct callback_name##_args {                                      \
    MAP_TUPLES(CALLBACK_struct_type, __VA_ARGS__)                    \
  };                                                                 \
  void callback_name##_bound(struct callback_name##_args args) {     \
    return callback_name(                                            \
        MAP_TUPLES_LIST(CALLBACK_struct_arg, __VA_ARGS__));          \
  };                                                                 \
  void callback_name(MAP_TUPLES_LIST(CALLBACK_fn_arg, __VA_ARGS__))