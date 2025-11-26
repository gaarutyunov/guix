//go:build js && wasm
// +build js,wasm

package runtime

import (
	"syscall/js"
)

// PatchType represents the type of patch to apply
type PatchType uint8

const (
	// PatchCreate creates a new node
	PatchCreate PatchType = iota
	// PatchDelete removes a node
	PatchDelete
	// PatchUpdateAttrs updates attributes
	PatchUpdateAttrs
	// PatchUpdateText updates text content
	PatchUpdateText
	// PatchMove moves a node to a new position
	PatchMove
	// PatchReplace completely replaces a node
	PatchReplace
	// PatchUpdateChildren recursively updates children
	PatchUpdateChildren
)

// Patch represents a change to apply to the DOM
type Patch struct {
	Type    PatchType
	OldNode *VNode
	NewNode *VNode
	Index   int
	Parent  js.Value
}

// Diff compares two VNode trees and returns patches
func Diff(oldNode, newNode *VNode) []Patch {
	if oldNode == nil && newNode == nil {
		return nil
	}

	if oldNode == nil {
		return []Patch{{
			Type:    PatchCreate,
			NewNode: newNode,
		}}
	}

	if newNode == nil {
		return []Patch{{
			Type:    PatchDelete,
			OldNode: oldNode,
		}}
	}

	// Different types - replace entirely
	if oldNode.Type != newNode.Type {
		return []Patch{{
			Type:    PatchReplace,
			OldNode: oldNode,
			NewNode: newNode,
		}}
	}

	var patches []Patch

	switch oldNode.Type {
	case TextNode:
		if oldNode.Text != newNode.Text {
			patches = append(patches, Patch{
				Type:    PatchUpdateText,
				OldNode: oldNode,
				NewNode: newNode,
			})
		}

	case ElementNode:
		// Different tags - replace
		if oldNode.Tag != newNode.Tag {
			return []Patch{{
				Type:    PatchReplace,
				OldNode: oldNode,
				NewNode: newNode,
			}}
		}

		// Check attributes
		if attrsChanged(oldNode.Attributes, newNode.Attributes) ||
			propsChanged(oldNode.Properties, newNode.Properties) {
			patches = append(patches, Patch{
				Type:    PatchUpdateAttrs,
				OldNode: oldNode,
				NewNode: newNode,
			})
		}

		// Diff children
		childPatches := DiffChildren(oldNode, newNode)
		patches = append(patches, childPatches...)

	case FragmentNode:
		childPatches := DiffChildren(oldNode, newNode)
		patches = append(patches, childPatches...)
	}

	return patches
}

// DiffChildren performs keyed reconciliation on children
func DiffChildren(oldParent, newParent *VNode) []Patch {
	oldChildren := oldParent.Children
	newChildren := newParent.Children

	if len(oldChildren) == 0 && len(newChildren) == 0 {
		return nil
	}

	var patches []Patch

	// Build key maps for keyed reconciliation
	oldKeyMap := make(map[interface{}]int)
	newKeyMap := make(map[interface{}]int)

	for i, child := range oldChildren {
		if child.Key != nil {
			oldKeyMap[child.Key] = i
		}
	}

	for i, child := range newChildren {
		if child.Key != nil {
			newKeyMap[child.Key] = i
		}
	}

	// Track which old children have been matched
	matched := make([]bool, len(oldChildren))

	// Process new children
	for newIdx, newChild := range newChildren {
		var oldChild *VNode
		var oldIdx int
		found := false

		// Try keyed lookup first
		if newChild.Key != nil {
			if idx, exists := oldKeyMap[newChild.Key]; exists {
				oldChild = oldChildren[idx]
				oldIdx = idx
				matched[idx] = true
				found = true
			}
		} else if newIdx < len(oldChildren) && oldChildren[newIdx].Key == nil {
			// Non-keyed positional matching
			oldChild = oldChildren[newIdx]
			oldIdx = newIdx
			matched[newIdx] = true
			found = true
		}

		if found {
			// Node exists, diff it
			if oldIdx != newIdx {
				patches = append(patches, Patch{
					Type:    PatchMove,
					OldNode: oldChild,
					NewNode: newChild,
					Index:   newIdx,
				})
			}

			// Recursively diff the matched nodes
			nodePatches := Diff(oldChild, newChild)
			patches = append(patches, nodePatches...)
		} else {
			// New node - create it
			patches = append(patches, Patch{
				Type:    PatchCreate,
				NewNode: newChild,
				Index:   newIdx,
				Parent:  oldParent.DOMNode,
			})
		}
	}

	// Remove unmatched old children
	for i, wasMatched := range matched {
		if !wasMatched {
			patches = append(patches, Patch{
				Type:    PatchDelete,
				OldNode: oldChildren[i],
			})
		}
	}

	return patches
}

// ApplyPatch applies a single patch to the DOM
func ApplyPatch(patch Patch) error {
	switch patch.Type {
	case PatchCreate:
		if patch.Parent.IsUndefined() {
			return nil // Will be handled by parent
		}

		var beforeNode js.Value
		if patch.Index < patch.Parent.Get("childNodes").Length() {
			beforeNode = patch.Parent.Get("childNodes").Index(patch.Index)
		}

		if beforeNode.IsUndefined() || beforeNode.IsNull() {
			return Mount(patch.NewNode, patch.Parent)
		}
		return InsertBefore(patch.Parent, patch.NewNode, beforeNode)

	case PatchDelete:
		Unmount(patch.OldNode)

	case PatchUpdateAttrs:
		UpdateElement(
			patch.OldNode,
			patch.OldNode.Attributes,
			patch.NewNode.Attributes,
			patch.OldNode.Properties,
			patch.NewNode.Properties,
		)
		// Update the old node's attributes to match new
		patch.OldNode.Attributes = patch.NewNode.Attributes
		patch.OldNode.Properties = patch.NewNode.Properties

	case PatchUpdateText:
		SetTextContent(patch.OldNode, patch.NewNode.Text)

	case PatchMove:
		parent := patch.OldNode.DOMNode.Get("parentNode")
		if parent.IsUndefined() || parent.IsNull() {
			return nil
		}

		var beforeNode js.Value
		if patch.Index < parent.Get("childNodes").Length() {
			beforeNode = parent.Get("childNodes").Index(patch.Index)
		}

		MoveNode(patch.OldNode, parent, beforeNode)

	case PatchReplace:
		return ReplaceNode(patch.OldNode, patch.NewNode)
	}

	return nil
}

// ApplyPatches applies a list of patches to the DOM
func ApplyPatches(patches []Patch) error {
	for _, patch := range patches {
		if err := ApplyPatch(patch); err != nil {
			return err
		}
	}
	return nil
}

// CopyDOMRefs copies DOMNode references from old VNode tree to new VNode tree
// This preserves DOM references after updates so subsequent diffs can find nodes
func CopyDOMRefs(oldNode, newNode *VNode) {
	if oldNode == nil || newNode == nil {
		return
	}

	// Copy the DOMNode reference
	if !oldNode.DOMNode.IsUndefined() && !oldNode.DOMNode.IsNull() {
		newNode.DOMNode = oldNode.DOMNode
	}

	// Recursively copy for children (match by position)
	oldChildren := oldNode.Children
	newChildren := newNode.Children
	minLen := len(oldChildren)
	if len(newChildren) < minLen {
		minLen = len(newChildren)
	}

	for i := 0; i < minLen; i++ {
		CopyDOMRefs(oldChildren[i], newChildren[i])
	}
}

// Helper functions

func attrsChanged(old, new map[string]string) bool {
	if len(old) != len(new) {
		return true
	}

	for k, v := range old {
		if newV, exists := new[k]; !exists || newV != v {
			return true
		}
	}

	return false
}

func propsChanged(old, new map[string]interface{}) bool {
	if len(old) != len(new) {
		return true
	}

	for k, v := range old {
		if newV, exists := new[k]; !exists || newV != v {
			return true
		}
	}

	return false
}
