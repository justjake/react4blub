package main

type hook interface {
	isHook()
}

func getOrCreateHook[HookType hook](fiber *fiber, makeHook func() HookType) (instance HookType, found bool) {
	var hook HookType
	found = true

	if len(fiber.hooks) > fiber.nextHook {
		hook = fiber.hooks[fiber.nextHook].(HookType)
	} else if fiber.mounted == nil {
		hook = makeHook()
		fiber.hooks = append(fiber.hooks, hook)
		found = false
	}

	fiber.nextHook++
	return hook, found
}

type Ref[T any] struct {
	Current T
}

type refHook[T any] struct {
	ref Ref[T]
}

func (r *refHook[T]) isHook() {}

// Create a ref in the current component.
func UseRef[T any]() *Ref[*T] {
	hook, _ := getOrCreateHook(globalState.currentFiber, func() *refHook[*T] {
		return &refHook[*T]{}
	})
	return &hook.ref
}

// Create a ref with the given initial value.
func UseRefInitial[T any](initialValue T) *Ref[T] {
	hook, _ := getOrCreateHook(globalState.currentFiber, func() *refHook[T] {
		return &refHook[T]{
			ref: Ref[T]{
				Current: initialValue,
			},
		}
	})
	return &hook.ref
}

type memoHook[T any, Dep comparable] struct {
	prev     T
	prevDeps Dep
}

func (_ *memoHook[T, Dep]) isHook() {}

func UseMemo[T any, Dep comparable](compute func() T, dependencies Dep) T {
	hook, found := getOrCreateHook(globalState.currentFiber, func() *memoHook[T, Dep] {
		return &memoHook[T, Dep]{
			prev:     compute(),
			prevDeps: dependencies,
		}
	})
	if found && dependencies != hook.prevDeps {
		hook.prev = compute()
	}
	return hook.prev
}

type stateHook[T comparable] struct {
	current T
	fiber   *fiber
}

func (_ *stateHook[T]) isHook() {}

func (state *stateHook[T]) SetState(nextState T) {
	if state.current == nextState {
		return
	}

	state.current = nextState
	// fiber.schedule()
}

func UseState[T comparable](initialState T) (state T, setState func(T)) {
	hook, _ := getOrCreateHook(globalState.currentFiber, func() *stateHook[T] {
		return &stateHook[T]{
			current: initialState,
			fiber:   globalState.currentFiber,
		}
	})
	return hook.current, hook.SetState
}

func Default[T any](pointer *T, defaultValue T) T {
	if pointer == nil {
		return defaultValue
	}
	return *pointer
}

func Some[T any](value T) *T {
	return &value
}

func UseCallback[T any, Deps comparable](fn T, dependencies Deps) T {
	return UseMemo(func() T { return fn }, dependencies)
}
