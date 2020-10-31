package mutation

import (
	"fmt"

	"github.com/open-policy-agent/gatekeeper/pkg/mutation/path/parser"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MetadataMutator is a mutator wrapping the AssignMetadata type.
type AssignMetadataMutator struct {
	//assignMetadata *mutationsv1.AssignMetadata
	Path  parser.Path
	Value string
}

func (m AssignMetadataMutator) Matches(obj metav1.Object, ns *corev1.Namespace) bool {
	// TODO: add match logic
	return false
}

func (m AssignMetadataMutator) Mutate(obj *unstructured.Unstructured) error {
	return mutate(obj.Object, m.Path.Nodes, m.Value)
}

func mutate(current interface{}, remainingPaths []parser.Node, value string) error {
	pathEntry := remainingPaths[0]
	if len(remainingPaths) == 1 {
		// TODO: handle types other than object
		current.(map[string]interface{})[pathEntry.(parser.Object).Reference] = value
		return nil
	}
	switch t := pathEntry.Type(); t {
	case parser.ObjectNode:
		next, ok := current.(map[string]interface{})[pathEntry.(parser.Object).Reference]
		if !ok {
			// TODO: Add pathTests and add path if required.
			// next := NewEmptyOfCorrectType(remainingPaths[1])
			return fmt.Errorf("Path specified by PathEntry does not exist for resource")
		}
		if err := mutate(next, remainingPaths[1:], value); err != nil {
			return err
		}
		return nil
	case parser.ListNode:
		for _, listElement := range current.([]interface{}) {
			if pathEntry.(parser.List).Glob {
				if err := mutate(listElement, remainingPaths[1:], value); err != nil {
					return err
				}
			} else if elementValue, ok := listElement.(map[string]interface{})[pathEntry.(parser.List).KeyField]; ok {
				if *pathEntry.(parser.List).KeyValue == elementValue {
					if err := mutate(listElement, remainingPaths[1:], value); err != nil {
						return err
					}
				}
			}
		}
	default:
		return fmt.Errorf("Unrecognized type: %v", t)
	}
	return nil
}

func (m AssignMetadataMutator) ID() {
}

func (m AssignMetadataMutator) HasDiff(mutator Mutator) bool {
	return false
}
