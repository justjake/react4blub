package reconciler

import "github.com/justjake/react4c/react"

type fiberHooks struct {
	allowMakeHook bool // false once mounted
	nextHook      int
	hooks         []react.HookInstance
	cb            react.HookCallbacks
}

func (h *fiberHooks) GetOrCreateHook(makeHook func() react.HookInstance) (instance react.HookInstance, found bool) {
	var hook react.HookInstance
	found = true

	if len(h.hooks) > h.nextHook {
		hook = h.hooks[h.nextHook]
	} else if h.allowMakeHook {
		hook = makeHook()
		h.hooks = append(h.hooks, hook)
		found = false
	}

	h.nextHook++
	return hook, found
}

func (h *fiberHooks) HookCallbacks() react.HookCallbacks {
	return h.cb
}
