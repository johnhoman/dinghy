package context

import (
	"context"
	"sync"
)

type Context struct {
	context.Context

	values map[string]any
	mu     sync.RWMutex
	debug  bool
}

func (ctx *Context) SetRoot(root string) {
	ctx.mu.Lock()
	ctx.values["root"] = root
	ctx.mu.Unlock()
}

func (ctx *Context) Root() string {
	ctx.mu.RLock()
	r, ok := ctx.values["root"]
	ctx.mu.RUnlock()
	if !ok {
		return ""
	}
	return r.(string)
}

func NewContext(debug bool) *Context {
	return &Context{
		Context: context.Background(),
		values:  make(map[string]any),
		mu:      sync.RWMutex{},
	}
}
