package resource

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"sync"
)

func NewList() *List {
	return &List{
		objs: make(map[Key]*Object, 0),
		mu:   sync.RWMutex{},
	}
}

type List struct {
	objs map[Key]*Object
	mu   sync.RWMutex
}

func (l *List) Visit(visitor Visitor, opts ...MatchOption) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, obj := range l.objs {
		if obj.Matches(opts...) {
			if err := visitor.Visit(obj); err != nil {
				// visitor will return a specific error here related
				// to the error associated with the resource kind, so just
				// propagate it up the chain
				return err
			}
		}
	}
	return nil
}

func (l *List) Insert(obj *Object) error {
	key := newResourceKey(obj)
	l.mu.Lock()
	if exists, ok := l.objs[key]; ok && obj.Equals(exists) {
		// return meaningful error
		return ErrResourceConflict
	}
	l.objs[key] = obj
	l.mu.Unlock()
	return nil
}

// Pop returns and removes the Object associated with the
// provided key
func (l *List) Pop(key Key) (*Object, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if obj, ok := l.objs[key]; ok {
		delete(l.objs, key)
		return obj, nil
	}
	return nil, ErrNotFound
}

func InsertFromReader(tree Tree, r io.Reader) error {
	d := yaml.NewDecoder(r)
	for {
		var m map[string]any
		if err := d.Decode(&m); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if len(m) == 0 {
			continue
		}
		if err := tree.Insert(Unstructured(m)); err != nil {
			return err
		}
	}
}
