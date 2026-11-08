package compiler

import "github.com/arrangio/core/tags"

type tagRegistry struct {
	tags map[string]int
	next int
}

func newTagRegistry() *tagRegistry {
	return &tagRegistry{tags: make(map[string]int)}
}

func (r *tagRegistry) getOrRegister(tag string) int {
	if id, ok := r.tags[tag]; ok {
		return id
	}
	id := r.next
	r.tags[tag] = id
	r.next++
	return id
}

func (r *tagRegistry) mask(names []string) tags.Mask {
	m := tags.NewMask()
	for _, name := range names {
		m = m.With(r.getOrRegister(name))
	}
	return m
}
