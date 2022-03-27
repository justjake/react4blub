#define ListMember(T, name) \
  T *next_##name;           \
  T *prev_##name;

#define ListOwner(T, name) \
  T *first_##name;         \
  T *last_##name;

struct list_owner {
  void **first;
  void **last;
};

struct list_member {
  void *self;
  void **prev;
  void **next;
};

void list_push(struct list_owner owner, struct list_member pushed) {
  if (*owner.first == 0 || *owner.last == 0) {
    *owner.first = pushed.self;
    *owner.last = pushed.self;
    return;
  }
}