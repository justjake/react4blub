#include "macro_map.h"

#define CALLBACK_name_bound(callback_name) callback_name##_as_callback
#define CALLBACK_name_heap(callback_name) callback_name##_as_memo
#define CALLBACK_name_struct(callback_name) callback_name##_args

#define CALLBACK_struct_type(typename, var) typename var;
#define CALLBACK_fn_arg(typename, var) typename var
#define CALLBACK_struct_initialize(typename, var) var,
#define CALLBACK_struct_arg(typename, var) args->var

#define CALLBACK_def_fn(T, callback_name, ...) \
  T callback_name(MAP_TUPLES_LIST(CALLBACK_fn_arg, __VA_ARGS__))

#define CALLBACK_def_bound(T, callback_name, ...)                            \
  CALLBACK_def_fn(T, callback_name, __VA_ARGS__);                            \
  struct CALLBACK_name_struct(callback_name){                                \
      MAP_TUPLES(CALLBACK_struct_type, __VA_ARGS__)};                        \
  T CALLBACK_name_bound(callback_name)(struct callback_name##_args * args) { \
    return callback_name(                                                    \
        MAP_TUPLES_LIST(CALLBACK_struct_arg, __VA_ARGS__));                  \
  };

#define CALLBACK_def_heap(T, callback_name)                                                    \
  void *CALLBACK_name_heap(callback_name)(struct CALLBACK_name_struct(callback_name) * args) { \
    void *result = malloc(sizeof(T));                                                          \
    *result = CALLBACK_name_bound(callback_name)(args);                                        \
    return result;                                                                             \
  };

#define useFunctionHook(hook_impl, callback_impl, ...)    \
  hook_impl(                                              \
      callback_impl,                                      \
      sizeof(struct CALLBACK_name_struct(callback_name)), \
      &(                                                  \
          (struct CALLBACK_name_struct(callback_name)){   \
              __VA_ARGS__}))

#define defineCallback(callback_name, ...)              \
  CALLBACK_def_bound(void, callback_name, __VA_ARGS__); \
  CALLBACK_def_fn(void, callback_name, __VA_ARGS__)
#define useCallback(callback_name, ...) useFunctionHook(useCallbackRaw, &CALLBACK_name_bound(callback_name), __VA_ARGS__)

#define defineMemo(T, callback_name, ...)            \
  CALLBACK_def_bound(T, callback_name, __VA_ARGS__); \
  CALLBACK_def_heap(T, callback_name);               \
  CALLBACK_def_fn(T, callback_name, __VA_ARGS__)
#define useMemo(callback_name, ...) useFunctionHook(useMemoRaw, &CALLBACK_name_heap(callback_name), __VA_ARGS__)

#define useEffect(callback_name, ...) useFunctionHook(useEffectRaw, callback_name, __VA_ARGS__)

typedef void *(*RHookResultAlloc)();

// typedef struct RHookFunction {
//   void
//   void *(alloc)();
// } RHookFunction;