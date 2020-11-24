package mutation

import (
	"fmt"

	"github.com/open-policy-agent/gatekeeper/pkg/mutation/path/parser"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ MutatorWithSchema = &BaseMutator{}

func NewMutator(path parser.Path, value string) BaseMutator {
	return BaseMutator{
		path:  path,
		Value: value,
	}
}

// Base is a base mutator wrapping the Assign and AssignMetadata type.
type BaseMutator struct {
	path  parser.Path
	Value string
}

func (m BaseMutator) Mutate(obj *unstructured.Unstructured) error {
	return mutate(m, obj.Object, nil, 0)
}

func mutate(m BaseMutator, current interface{}, previous interface{}, depth int) error {
	if len(m.Path().Nodes)-1 == depth {
		return addValue(m, current, previous, depth)
	}

	pathEntry := m.Path().Nodes[depth]
	switch t := pathEntry.Type(); t {
	case parser.ObjectNode:
		next, ok := current.(map[string]interface{})[pathEntry.(parser.Object).Reference]
		if !ok {
			next = createMissingElement(m, current, previous, depth)
		}
		if err := mutate(m, next, current, depth+1); err != nil {
			return err
		}
		return nil
	case parser.ListNode:
		elementFound := false
		glob := pathEntry.(parser.List).Glob
		key := pathEntry.(parser.List).KeyField
		for _, listElement := range current.([]interface{}) {
			if glob {
				if err := mutate(m, listElement, current, depth+1); err != nil {
					return err
				}
				elementFound = true
			} else if elementValue, ok := listElement.(map[string]interface{})[key]; ok {
				if *pathEntry.(parser.List).KeyValue == elementValue {
					if err := mutate(m, listElement, current, depth+1); err != nil {
						return err
					}
					elementFound = true
				}
			}
		}
		// If no matching element in the array was found in non Globbed list, create a new element
		if !pathEntry.(parser.List).Glob && !elementFound {
			next := createMissingElement(m, current, previous, depth)
			if err := mutate(m, next, current, depth+1); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Unrecognized type: %v", t)
	}
	return nil
}

func addValue(m BaseMutator, current interface{}, previous interface{}, depth int) error {
	// TODO: it should be considered if the value set can be not just a simple string, but json which could be unmarshalled into an object
	pathEntry := m.Path().Nodes[depth]
	switch t := pathEntry.Type(); t {
	case parser.ObjectNode:
		if elementValue, ok := current.(map[string]interface{})[pathEntry.(parser.Object).Reference]; ok {
			if elementValue != m.Value {
				return fmt.Errorf("A value can not me modified by mutation")
			}
		} else {
			current.(map[string]interface{})[pathEntry.(parser.Object).Reference] = m.Value
		}
	case parser.ListNode:
		return addListElementWithValue(m, current, previous, depth)
	}
	return nil
}

func addListElementWithValue(m BaseMutator, current interface{}, previous interface{}, depth int) error {
	pathEntry := m.Path().Nodes[depth]

	if pathEntry.(parser.List).Glob {
		return fmt.Errorf("Last path entry can not be globbed")
	}
	key := pathEntry.(parser.List).KeyField
	keyValue := *pathEntry.(parser.List).KeyValue

	for _, listElement := range current.([]interface{}) {
		if elementValue, ok := listElement.(map[string]interface{})[key]; ok && keyValue == elementValue {
			return nil // Element is already present, skip the update
		}
	}
	current = append(current.([]interface{}), map[string]interface{}{key: keyValue})
	previous.(map[string]interface{})[m.Path().Nodes[depth-1].(parser.Object).Reference] = current
	return nil
}

func createMissingElement(m BaseMutator, current interface{}, previous interface{}, depth int) interface{} {
	var next interface{}
	pathEntry := m.Path().Nodes[depth]

	// Create new element of type
	switch m.Path().Nodes[depth+1].Type() {
	case parser.ObjectNode:
		next = make(map[string]interface{})
	case parser.ListNode:
		next = make([]interface{}, 0)
	}

	// Append to element of type
	switch pathEntry.Type() {
	case parser.ObjectNode:
		current.(map[string]interface{})[pathEntry.(parser.Object).Reference] = next
	case parser.ListNode:
		current = append(current.([]interface{}), next)
		next.(map[string]interface{})[pathEntry.(parser.List).KeyField] = *pathEntry.(parser.List).KeyValue
		previous.(map[string]interface{})[m.Path().Nodes[depth-1].(parser.Object).Reference] = current
	}
	return next
}

func (m BaseMutator) ID() ID {
	return ID{}
}

func (m BaseMutator) HasDiff(mutator Mutator) bool {
	return false
}

func (m BaseMutator) Path() *parser.Path {
	return &m.path
}

func (m BaseMutator) SchemaBindings() []SchemaBinding {
	return nil
}

func (m BaseMutator) DeepCopy() Mutator {
	return nil
}

func (m BaseMutator) Matches(obj metav1.Object, ns *corev1.Namespace) bool {
	return false
}
