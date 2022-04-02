package reconciler

import (
	. "github.com/justjake/react4c/react"
)

type HostKind interface {
	// comparable
}

type ParentInstance[K HostKind] interface {
	IsParent() K
	// AppendChild(child ChildInstance)
	// InsertChildBefore(child ChildInstance, beforeChild ChildInstance)
	// RemoveChild(child ChildInstance)
	// ClearChildren()
}

// Host that has children, but no props, tag, etc.
type Container[K HostKind] interface {
	IsContainer() K
	// No-ops in the test renderer
	// PrepareForCommit()
	// ResetAfterCommit()
}

type ChildInstance[K HostKind] interface {
	IsChild() K
}

// Leaf host that can never have children.
// Used for rendering text/strings
// For a DOM host, this would be a TextNode.
type TextInstance[K HostKind] interface {
	ChildInstance[K]
	IsText() K
}

// General host instance produced by rendering a host component.
// For a HTML DOM host, this would be an Element.
type Instance[K HostKind] interface {
	ChildInstance[K]
	ParentInstance[K]
	// Append child if missing, or make child the first child.
	// https://github.com/facebook/react/blob/848e802d203e531daf2b9b0edb281a1eb6c5415d/packages/react-test-renderer/src/ReactTestHostConfig.js#L163
	// AppendInitialChild(child ChildInstance[K])
}

type HostProps interface {
	ComparableProps
}

// Not sure if these should be type parameters or interfaces.

type HostContext interface{}
type InternalInstanceHandle interface{}
type HostUpdate interface{}

// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L395-L411
type HostConfigMicrotaskSupport interface {
	ScheduleMicrotask(task func())
}

// Support for mutation.
//
// Our MVP is RenderToString (no mutation) and rendering with the test DOM
// dingus (uses mutation), so we need to support mutation.
//
// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L413-L417
//
//     export const supportsMutation = false;
//     export const appendChild = shim;
//     export const appendChildToContainer = shim;
//     export const commitTextUpdate = shim;
//     export const commitMount = shim;
//     export const commitUpdate = shim;
//     export const insertBefore = shim;
//     export const insertInContainerBefore = shim;
//     export const removeChild = shim;
//     export const removeChildFromContainer = shim;
//     export const resetTextContent = shim;
//     export const hideInstance = shim;
//     export const hideTextInstance = shim;
//     export const unhideInstance = shim;
//     export const unhideTextInstance = shim;
//     export const clearContainer = shim;
type HostConfigMutationSupport[
	Props HostProps,
	Comp ComparableComponent[Props],
	K HostKind,
] interface {
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L480-L483
	AppendChild(parent Instance[K], child ChildInstance[K])

	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L487-L490
	AppendChildToContainer(parent Container[K], child ChildInstance[K])

	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L472-L476
	CommitTextUpdate(inst TextInstance[K], oldText string, newText string)

	CommitMount(inst Instance[K], comp Comp, newProps Props, internalInstanceHandle InternalInstanceHandle)

	// Apply update prepared with PrepareUpdate
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L453-L460
	CommitUpdate(inst Instance[K], updatePayload []HostUpdate, comp Comp, oldProps Props, newProps Props, internalInstanceHandle InternalInstanceHandle)

	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L517-L521
	InsertBefore(parent Instance[K], child ChildInstance[K], beforeChild ChildInstance[K])

	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L525-L529
	InsertInContainerBefore(parent Container[K], child ChildInstance[K], beforeChild ChildInstance[K])

	RemoveChild(parent Instance[K], child ChildInstance[K])
	RemoveChildFromContainer(parent Container[K], child ChildInstance[K])

	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L468-L470
	ResetTextContent(inst Instance[K])

	// set display=hidden
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L640-L642
	HideInstance(inst Instance[K])
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L652
	HideTextInstance(inst TextInstance[K])

	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L656
	UnhideInstance(inst Instance[K], props Props)
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L668
	UnhideTextInstance(inst TextInstance[K], text string)

	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L675
	ClearContainer(container Container[K])
}

type Unsupported interface {
	TODO()
}

// Hydration also needs a SuspenseInstance type. TODO.
// https://github.com/facebook/react/blob/2af4a79333108f59600970ef23e2614c3a5324e4/packages/react-reconciler/src/ReactFiberHostConfigWithNoHydration.js#L21-L23
type HostConfigHydrationSupport interface {
	Unsupported
}

// https://github.com/facebook/react/blob/a724a3b578dce77d427bef313102a4d0e978d9b4/packages/react-reconciler/src/ReactFiberHostConfigWithNoPersistence.js#L21-L22
type HostConfigPersistenceSupport interface {
	Unsupported
}

// https://github.com/facebook/react/blob/a724a3b578dce77d427bef313102a4d0e978d9b4/packages/react-reconciler/src/ReactFiberHostConfigWithNoScopes.js#L22-L23
type HostConfigScopesSupport interface {
	Unsupported
}

// https://github.com/facebook/react/blob/a724a3b578dce77d427bef313102a4d0e978d9b4/packages/react-reconciler/src/ReactFiberHostConfigWithNoTestSelectors.js#L21-L22
//
// export const supportsTestSelectors = false;
// export const findFiberRoot = shim;
// export const getBoundingRect = shim;
// export const getTextContent = shim;
// export const isHiddenSubtree = shim;
// export const matchAccessibilityRole = shim;
// export const setFocusIfFocusable = shim;
// export const setupIntersectionObserver = shim;
type HostConfigTestSelectors interface {
	Unsupported
}

// Work-in-progress sketch of react-reconciler HostConfig.  The major difference
// in our design compared to react-reconciler is in expressiveness of the type
// system. react-reconciler falls firmly in the ADT-switch-statement approach to
// the expression problem. Eg, `Type` for them is usually a string enum, and the
// HostConfig implements everything as a `switch(type)`, since that is very
// natural in Flow/Typescript/Javascript, and class-instance-interface is less
// safe.
//
// In Golang, the compiler has much more support for reasoning about interfaces
// and type-casting as a first-level concept. So, instead of encouraging a a switch(...)
// ADT interface, we instead strive to be Interface Oriented.
//
// First we'll do a direct port, then we'll see about moving to be InterfaceOriented.
type HostConfig[
	// TODO: do we really need to take type parameters (go 1.18+ style), or can we
	// move all of this functionality behind interface types (go 1.0 style)?
	Props HostProps,
	Comp ComparableComponent[Props],
	K HostKind,
] interface {
	// https://github.com/facebook/react/blob/848e802d203e531daf2b9b0edb281a1eb6c5415d/packages/react-test-renderer/src/ReactTestHostConfig.js#L145-L151
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L247-L253
	CreateInstance(comp Comp, props Props, rootContainerInstance Container[K], hostContext HostContext, internalInstanceHandle InternalInstanceHandle) Instance[K]
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L357-L362
	CreateTextInstance(text string, rootContainerInstance Container[K], hostContext HostContext, internalInstanceHandle InternalInstanceHandle) TextInstance[K]

	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L285
	AppendInitialChild(parent Instance[K], child ChildInstance[K])
	// https://github.com/facebook/react/blob/848e802d203e531daf2b9b0edb281a1eb6c5415d/packages/react-test-renderer/src/ReactTestHostConfig.js#L174-L180
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L292-L298
	FinalizeInitialChildren(inst Instance[K], comp Comp, props Props, rootContainerInstance Container[K], hostContext HostContext) bool

	// Diff properties. Return nil on no update.
	// https://github.com/facebook/react/blob/1c44437355e21f2992344fdef9ab1c1c5a7f8c2b/packages/react-dom/src/client/ReactDOMHostConfig.js#L313-L320
	PrepareUpdate(inst Instance[K], comp Comp, oldProps Props, newProps Props, rootContainerInstance Container[K], hostContext HostContext) HostUpdate // | null

	// Return nil if unsupported
	// https://github.com/foacebook/react/blob/a724a3b578dce77d427bef313102a4d0e978d9b4/packages/react-reconciler/src/ReactFiberHostConfigWithNoMutation.js#L21-L22
	//
	// If mutation is unsupported, we can only append children during the initial render.
	// Example: https://github.com/facebook/react/blob/05c283c3c31184d68c6a54dfd6a044790b89a08a/packages/react-native-renderer/src/ReactFabricHostConfig.js#L320
	SupportMutation() HostConfigMutationSupport[Props, Comp, K]

	// Optional host features that our reconciler itself doesn't support yet.
	// Stubbed here to discover & follow up later/never.
	// Implementations should return nil for now.
	SupportHydration() HostConfigHydrationSupport
	SupportPersistence() HostConfigPersistenceSupport
	SupportScopes() HostConfigScopesSupport
	SupportTestSelectors() HostConfigTestSelectors
	SupportMicrotask() HostConfigMicrotaskSupport
}
