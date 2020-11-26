package schema

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/open-policy-agent/gatekeeper/pkg/mutation/path/parser"
	"github.com/open-policy-agent/gatekeeper/pkg/mutation/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ MutatorWithSchema = &mockMutator{}

type mockMutator struct {
	id        types.ID
	ForceDiff bool
	Bindings  []Binding
	path      string
	pathCache *parser.Path
}

func (m *mockMutator) Matches(obj runtime.Object, ns *corev1.Namespace) bool { return false }

func (m *mockMutator) Mutate(obj *unstructured.Unstructured) error { return nil }

func (m *mockMutator) ID() types.ID { return m.id }

func (m *mockMutator) HasDiff(other types.Mutator) bool {
	if m.ForceDiff {
		return true
	}
	return !reflect.DeepEqual(m, other)
}

func deepCopyBindings(bindings []Binding) []Binding {
	cpy := []Binding{}
	for _, b := range bindings {
		cpy = append(cpy, Binding{
			Groups:   append([]string{}, b.Groups...),
			Kinds:    append([]string{}, b.Kinds...),
			Versions: append([]string{}, b.Versions...),
		})
	}
	return cpy
}

func (m *mockMutator) DeepCopy() types.Mutator {
	return &mockMutator{
		id:        m.id,
		ForceDiff: m.ForceDiff,
		Bindings:  deepCopyBindings(m.Bindings),
		path:      m.path,
	}
}

func (m *mockMutator) SchemaBindings() []Binding { return m.Bindings }

func (m *mockMutator) Path() *parser.Path {
	if m.pathCache != nil {
		return m.pathCache
	}
	out, err := parser.Parse(m.path)
	if err != nil {
		panic(err)
	}
	m.pathCache = out
	return out
}

func id(name string) types.ID {
	return types.ID{Name: name}
}

func bindings(kind string) []Binding {
	return []Binding{{Groups: []string{""}, Versions: []string{"v1"}, Kinds: []string{kind}}}
}

func gvk(kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    kind,
	}
}

func sp(s string) *string {
	return &s
}

func simpleMutator(mid, kind, path string) *mockMutator {
	return &mockMutator{
		id:       id(mid),
		Bindings: bindings(kind),
		path:     path,
	}
}

const (
	upsert = "upsert"
	remove = "remove"
)

type testOp struct {
	op            string
	errorExpected bool
	id            types.ID
	mutator       *mockMutator
}

// TODO write more tests
func TestParser(t *testing.T) {
	tests := []struct {
		name             string
		ops              []testOp
		expectedMutators map[types.ID]MutatorWithSchema
		expectedSchemas  map[schema.GroupVersionKind]*scheme
	}{
		{
			name: "Simple upsert",
			ops: []testOp{
				{
					op:      upsert,
					mutator: simpleMutator("simple", "FooKind", "spec.someValue"),
				},
			},
			expectedMutators: map[types.ID]MutatorWithSchema{
				id("simple"): simpleMutator("simple", "FooKind", "spec.someValue"),
			},
			expectedSchemas: map[schema.GroupVersionKind]*scheme{
				gvk("FooKind"): {
					gvk: gvk("FooKind"),
					root: &node{
						referenceCount: 1,
						nodeType:       parser.ObjectNode,
						children: map[string]*node{
							"spec": {
								referenceCount: 1,
								nodeType:       parser.ObjectNode,
								children: map[string]*node{
									"someValue": {
										referenceCount: 1,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Add and remove simple",
			ops: []testOp{
				{
					op:      upsert,
					mutator: simpleMutator("simple", "FooKind", "spec.someValue"),
				},
				{
					op: remove,
					id: id("simple"),
				},
			},
			expectedMutators: map[types.ID]MutatorWithSchema{},
			expectedSchemas:  map[schema.GroupVersionKind]*scheme{},
		},
		{
			name: "Simple upsert with list",
			ops: []testOp{
				{
					op:      upsert,
					mutator: simpleMutator("simple", "FooKind", "spec.someValue[hey: \"there\"]"),
				},
			},
			expectedMutators: map[types.ID]MutatorWithSchema{
				id("simple"): simpleMutator("simple", "FooKind", "spec.someValue[hey: \"there\"]"),
			},
			expectedSchemas: map[schema.GroupVersionKind]*scheme{
				gvk("FooKind"): {
					gvk: gvk("FooKind"),
					root: &node{
						referenceCount: 1,
						nodeType:       parser.ObjectNode,
						children: map[string]*node{
							"spec": {
								referenceCount: 1,
								nodeType:       parser.ObjectNode,
								children: map[string]*node{
									"someValue": {
										referenceCount: 1,
										nodeType:       parser.ListNode,
										keyField:       sp("hey"),
										child: &node{
											referenceCount: 1,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := New()
			for _, op := range test.ops {
				switch op.op {
				case upsert:
					err := db.Upsert(op.mutator)
					if op.errorExpected != (err != nil) {
						t.Errorf("error = %v, which is unexpected", err)
					}
				case remove:
					db.Remove(op.id)
				default:
					t.Error("malformed test: unrecognized op")
				}
			}
			if test.expectedSchemas != nil {
				if !reflect.DeepEqual(db.schemas, test.expectedSchemas) {
					t.Errorf("Difference in schemas: %v",
						cmp.Diff(db.schemas, test.expectedSchemas, cmp.AllowUnexported(
							scheme{},
							node{},
						)))
				}
			}
			if test.expectedMutators != nil {
				if !reflect.DeepEqual(db.mutators, test.expectedMutators) {
					t.Errorf("Difference in schemas: %v", cmp.Diff(db.mutators, test.expectedMutators, cmp.AllowUnexported(
						mockMutator{},
					)))
				}
			}
		})
	}
}
