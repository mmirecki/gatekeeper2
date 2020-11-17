package schema

import (
	"fmt"
	"strings"
	"sync"

	"github.com/open-policy-agent/gatekeeper/pkg/mutation/path/parser"
	"github.com/open-policy-agent/gatekeeper/pkg/mutation/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// TODO we are assuming that remove will never get corrupted.
// this should be the case, but is that wise?

// Binding represent the specific GVKs that a
// mutation's implicit schema applies to
type Binding struct {
	Groups   []string
	Kinds    []string
	Versions []string
}

// MutatorWithSchema is a mutator exposing the implied
// schema of the target object.
type MutatorWithSchema interface {
	types.Mutator
	SchemaBindings() []Binding
	Path() *parser.Path
}

var (
	log = logf.Log.WithName("mutation_schema")
)

// New returns a new schema database
func New() *DB {
	return &DB{
		mutators: map[types.ID]MutatorWithSchema{},
		schemas:  map[schema.GroupVersionKind]*scheme{},
	}
}

// DB is a database that caches all the implied schemas.
// Will return an error when adding a mutator conflicting with the existing ones.
type DB struct {
	mutex    sync.Mutex
	mutators map[types.ID]MutatorWithSchema
	schemas  map[schema.GroupVersionKind]*scheme
}

// Upsert tries to insert or update the given mutator.
// If a conflict is detected, Upsert will return an error
func (db *DB) Upsert(mutator MutatorWithSchema) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	return db.upsert(mutator, true)
}

func (db *DB) upsert(mutator MutatorWithSchema, unwind bool) error {
	oldMutator, ok := db.mutators[mutator.ID()]
	if ok && !oldMutator.HasDiff(mutator) {
		return nil
	}
	if ok {
		db.remove(oldMutator.ID())
	}
	bindings := mutator.SchemaBindings()
	gvks := map[schema.GroupVersionKind]struct{}{}
	// deduplicate GVKs
	for _, binding := range bindings {
		for _, group := range binding.Groups {
			for _, version := range binding.Versions {
				for _, kind := range binding.Kinds {
					gvk := schema.GroupVersionKind{
						Group:   group,
						Version: version,
						Kind:    kind,
					}
					gvks[gvk] = struct{}{}
				}
			}
		}
	}

	modified := []*scheme{}
	for gvk := range gvks {
		s, ok := db.schemas[gvk]
		if !ok {
			s = &scheme{gvk: gvk}
			db.schemas[gvk] = s
		}
		if err := s.add(mutator.Path().Nodes); err != nil {
			// avoid infinite recursion
			if unwind {
				db.unwind(mutator, oldMutator, modified)
			}
			return err
		}
		modified = append(modified, s)
	}
	m := mutator.DeepCopy()
	db.mutators[mutator.ID()] = m.(MutatorWithSchema)
	return nil
}

// unwind a bad commit
func (db *DB) unwind(new, old MutatorWithSchema, schemes []*scheme) {
	for _, s := range schemes {
		s.remove(new.Path().Nodes)
		if s.root == nil {
			delete(db.schemas, s.gvk)
		}
	}
	if old == nil {
		return
	}
	if err := db.upsert(old, false); err != nil {
		// We removed all changes made by the previous mutator and
		// are re-adding a mutator that was already present. Because
		// this mutator was already present and we have a lock on the
		// db, this should never fail. If it does we are in an unknown
		// state and should panic so we can recover by bootstrapping
		// and raise the visibility of the issue.
		log.Error(err, "could not upsert previously existing mutator into schema, this is not recoverable")
		panic(err)
	}
}

// Remove removes the mutator with the given id from the
// db.
func (db *DB) Remove(id types.ID) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.remove(id)
	delete(db.mutators, id)
}

func (db *DB) remove(id types.ID) {
	mutator, ok := db.mutators[id]
	if !ok {
		// no mutator found, nothing to do
		return
	}
	bindings := mutator.SchemaBindings()
	gvks := map[schema.GroupVersionKind]struct{}{}
	// deduplicate GVKs
	for _, binding := range bindings {
		for _, group := range binding.Groups {
			for _, version := range binding.Versions {
				for _, kind := range binding.Kinds {
					gvk := schema.GroupVersionKind{
						Group:   group,
						Version: version,
						Kind:    kind,
					}
					gvks[gvk] = struct{}{}
				}
			}
		}
	}

	for gvk := range gvks {
		s, ok := db.schemas[gvk]
		if !ok {
			log.Error(nil, "mutator associated with missing schema", "mutator", id, "schema", gvk)
			panic(fmt.Sprintf("mutator %v associated with missing schema %v", id, gvk))
		}
		s.remove(mutator.Path().Nodes)
		if s.root == nil {
			delete(db.schemas, gvk)
		}
	}
}

type scheme struct {
	gvk  schema.GroupVersionKind
	root *node
}

func (s *scheme) add(ref []parser.Node) error {
	if s.root == nil {
		s.root = &node{}
	}
	return s.root.add(ref)
}

func (s *scheme) remove(ref []parser.Node) {
	if s.root == nil {
		return
	}
	s.root.remove(ref)
	if s.root.referenceCount == 0 {
		s.root = nil
	}
}

type node struct {
	referenceCount uint
	nodeType       parser.NodeType

	// list-type nodes have a key field and only one child
	keyField *string
	child    *node

	// object-type nodes only have children
	children map[string]*node
}

// backup creates a shallow copy of the node we can restore in case of error
func (n *node) backup() *node {
	return &node{
		referenceCount: n.referenceCount,
		nodeType: n.nodeType,
		keyField: n.keyField,
	}
}

func (n *node) restore(backup *node) {
	// we assume reference count is always incremented, so we can avoid storing it
	// in the stack
	n.referenceCount = backup.referenceCount
	n.nodeType = backup.nodeType
	n.keyField = backup.keyField
}

func (n *node) add(ref []parser.Node) error {
	backup := n.backup()
	n.referenceCount++
	if len(ref) == 0 {
		return nil
	}

	current := ref[0]
	switch t := current.Type(); t {
	case parser.ObjectNode:
		if n.nodeType != "" && n.nodeType != parser.ObjectNode {
			return fmt.Errorf("node type conflict: %v vs %v", n.nodeType, parser.ObjectNode)
		}
		n.nodeType = parser.ObjectNode
		obj := current.(*parser.Object)
		if n.children == nil {
			n.children = make(map[string]*node)
		}
		child, ok := n.children[obj.Reference]
		if !ok {
			child = &node{}
		}
		if err := child.add(ref[1:]); err != nil {
			n.restore(backup)
			return wrapObjErr(obj, err)
		}
		n.children[obj.Reference] = child
	case parser.ListNode:
		if n.nodeType != "" && n.nodeType != parser.ListNode {
			return fmt.Errorf("node type conflict: %v vs %v", n.nodeType, parser.ObjectNode)
		}
		n.nodeType = parser.ListNode
		list := current.(*parser.List)
		if n.keyField != nil && *n.keyField != list.KeyField {
			return fmt.Errorf("key field conflict: %s vs %s", *n.keyField, list.KeyField)
		}
		if n.keyField == nil {
			n.keyField = stringPointer(list.KeyField)
		}
		child := n.child
		if child == nil {
			child = &node{}
		}
		if err := child.add(ref[1:]); err != nil {
			n.restore(backup)
			return wrapListErr(list, err)
		}
		n.child = child
	default:
		return fmt.Errorf("unknown node type: %v", t)
	}
	return nil
}

func (n *node) remove(ref []parser.Node) {
	n.referenceCount--
	if len(ref) == 0 {
		return
	}
	current := ref[0]
	switch t := current.Type(); t {
	case parser.ObjectNode:
		obj := current.(*parser.Object)
		if n.children == nil {
			// no children means nothing to clean
			return
		}
		child, ok := n.children[obj.Reference]
		if !ok {
			// child is missing, nothing to clean
			return
		}
		// decrementing the reference count would remove the
		// object, we can stop traversing the tree
		if child.referenceCount <= 1 {
			delete(n.children, obj.Reference)
			return
		}
		child.remove(ref[1:])
		return
	case parser.ListNode:
		if n.child == nil {
			// no child, nothing to clean
			return
		}
		// decrementing the reference count would remove the
		// object, we can stop traversing the tree
		if n.child.referenceCount <= 1 {
			n.child = nil
			return
		}
		n.child.remove(ref[1:])
		return
	default:
		log.Error(fmt.Errorf("unknown node type"), "unknown node type", "node_type", t)
		panic(fmt.Sprintf("unknown node type, schema db in unknown state: %s", t))
	}
}

// return a pointer to a copied string for safety
func stringPointer(s string) *string {
	return &s
}

var _ error = &Error{}

// Error holds errors processing the schema
type Error struct {
	nodeName   string
	childError error
}

func (e *Error) Error() string {
	builder := &strings.Builder{}
	current := e
	for {
		builder.WriteString(current.nodeName)
		child, ok := current.childError.(*Error)
		if !ok {
			break
		}
		current = child
	}
	builder.WriteString(": ")
	builder.WriteString(current.childError.Error())
	return strings.TrimPrefix(builder.String(), ".")
}

func wrapObjErr(obj *parser.Object, err error) *Error {
	return &Error{
		childError: err,
		nodeName:   fmt.Sprintf(".%s", obj.Reference),
	}
}

func wrapListErr(list *parser.List, err error) *Error {
	var value string
	if list.Glob {
		value = "*"
	} else {
		value = fmt.Sprintf("\"%s\"", *list.KeyValue)
	}
	return &Error{
		childError: err,
		nodeName:   fmt.Sprintf("[\"%s\": %s]", list.KeyField, value),
	}
}
