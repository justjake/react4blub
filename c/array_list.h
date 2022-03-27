#define ArrayList(T) \
  struct {           \
    int cap;         \
    int length;      \
    T *vals;         \
  }
