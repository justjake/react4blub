package react

// Internal interface between a hook instance and internal reconciler machinery
// for handling state updates.
type HookCallbacks interface {
	ShouldRerender()
}

type HookHost interface {
	HookCallbacks() HookCallbacks
	GetOrCreateHook(makeHook func() HookInstance) (instance HookInstance, found bool)
}

type HookInstance interface {
	Unmount()
}

var currentHookHost HookHost

func getOrCreateHook[HookType HookInstance](hookHost HookHost, makeHook func() HookType) (instance HookType, found bool) {
	untyped, found := hookHost.GetOrCreateHook(func() HookInstance {
		return makeHook()
	})
	return untyped.(HookType), found
}

type refHook[T any] struct {
	ref RefStruct[T]
}

func (r *refHook[T]) Unmount() {}

// Create a ref in the current component.
func UseRef[T any]() *RefStruct[*T] {
	hook, _ := getOrCreateHook(currentHookHost, func() *refHook[*T] {
		return &refHook[*T]{}
	})
	return &hook.ref
}

// Create a ref with the given initial value.
func UseRefInitial[T any](initialValue T) *RefStruct[T] {
	hook, _ := getOrCreateHook(currentHookHost, func() *refHook[T] {
		return &refHook[T]{
			ref: RefStruct[T]{
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

func (*memoHook[T, Dep]) Unmount() {}

func UseMemo[T any, Dep comparable](compute func() T, dependencies Dep) T {
	hook, found := getOrCreateHook(currentHookHost, func() *memoHook[T, Dep] {
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
	handle  HookCallbacks
}

func (state *stateHook[T]) Unmount() {
	state.handle = nil
}

func (state *stateHook[T]) SetState(nextState T) {
	if state.handle == nil {
		return
	}

	if state.current == nextState {
		return
	}

	state.current = nextState
	state.handle.ShouldRerender()
}

func UseState[T comparable](initialState T) (state T, setState func(T)) {
	hook, _ := getOrCreateHook(currentHookHost, func() *stateHook[T] {
		return &stateHook[T]{
			current: initialState,
			handle:  currentHookHost.HookCallbacks(),
		}
	})
	return hook.current, hook.SetState
}

func UseStateLazy[T comparable](getInitialState func() T) (state T, setState func(T)) {
	hook, _ := getOrCreateHook(currentHookHost, func() *stateHook[T] {
		return &stateHook[T]{
			current: getInitialState(),
			handle:  currentHookHost.HookCallbacks(),
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

type EffectFunc interface {
	func() | func() func()
}

type effectHook[T EffectFunc, Deps comparable] struct {
	pending  *T
	cleanup  *func()
	prevDeps Deps
	// TODO: phase (effect/layoutEffect/...)
}

func (effect *effectHook[T, Deps]) Unmount() {
	if effect.cleanup != nil {
		(*effect.cleanup)()
		effect.cleanup = nil
	}
}

func (effect *effectHook[T, Deps]) mount() {
	if effect.pending != nil {
		pending := *effect.pending
		effect.Unmount()

		if withCleanup, ok := any(pending).(func() func()); ok {
			res := withCleanup()
			effect.cleanup = &res
		} else {
			any(pending).(func())()
		}

		effect.pending = nil
	}
}

func UseEffect[T EffectFunc, Deps comparable](fn T, dependencies Deps) {
	hook, found := getOrCreateHook(currentHookHost, func() *effectHook[T, Deps] {
		return &effectHook[T, Deps]{
			pending:  &fn,
			prevDeps: dependencies,
		}
	})

	if !found && dependencies != hook.prevDeps {
		hook.pending = &fn
	}
}
