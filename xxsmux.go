// Package xxsmux implements functionality for building http.DefaultServeMux.
//
// The xxsmux package is a default serve mux builder.
// Build patterns, handlers and wrap middlewares conveniently upfront.
// The xxsmux.defaultServeMuxBuilder acts as a builder for the http.DefaultServeMux.
// The overall goal of this package is to build the http.DefaultServeMux with pattern/path prefixes and middleware wired in.
package xxsmux

import (
	"net/http"
	"regexp"
	"strings"
)

// defaultServeMuxBuilder is a simple builder for the http.DefaultServeMux.
type defaultServeMuxBuilder struct {
	// patterns represent the given patterns to `http.Handle`/`http.HandleFunc`.
	patterns map[string]http.Handler
	// patternPrefix represent the prefix of the pattern of a subrouter.
	patternPrefix string
	// middlewares represent the middlewares that wrap the subrouter.
	middlewares []middleware
	// root always points to the root node of the default servce mux builder.
	root *defaultServeMuxBuilder
	// parent always points to the parent node.
	// For the `root` field the parent would also be `root`.
	parent *defaultServeMuxBuilder

	// subDefaultServeMuxBuilder stores the subrouters of the main router.
	subDefaultServeMuxBuilder []*defaultServeMuxBuilder
}

// middleware represents an http.Handler wrapper to inject addional functionality.
type middleware func(http.Handler) http.Handler

// New returns a new defaultServeMuxBuilder.
func New() *defaultServeMuxBuilder {
	b := &defaultServeMuxBuilder{patterns: map[string]http.Handler{}}
	b.root = b
	b.parent = b
	return b
}

// newHandler returns an http.Handler wrapped with given middlewares.
func newHandler(mw ...middleware) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		for _, m := range mw {
			h = m(h)
		}
		return h
	}
}

// Pattern registers hanglers for given patterns.
func (b *defaultServeMuxBuilder) Pattern(patterns map[string]http.Handler) {
	patternPrefix := b.patternPrefix
	b.patternPrefix = ""
	b.patternPrefix = b.root.patternPrefix + "/"

	for _, subBuilder := range b.parent.subDefaultServeMuxBuilder {
		if b.parent == b.root {
			if subBuilder == b {
				for _, subSubBuilder := range subBuilder.subDefaultServeMuxBuilder {
					b.patternPrefix = subSubBuilder.patternPrefix + "/"
				}
			}
		} else {
			for _, subSubBuilder := range subBuilder.subDefaultServeMuxBuilder {
				b.patternPrefix = subSubBuilder.patternPrefix + "/"
			}
		}
	}

	b.patternPrefix += patternPrefix

	for pattern, handler := range patterns {
		// TODO: strings.Split could fail and not have 2 elements
		b.patterns[removeDoubleSlash(b.patternPrefix+strings.Split(pattern, " ")[1])] = handler
	}
	b.subDefaultServeMuxBuilder = append(b.subDefaultServeMuxBuilder, b)
}

// removeDoubleSlash cleans up strings for double slashes `//`.
func removeDoubleSlash(text string) string {
	re := regexp.MustCompile(`//+`)
	return re.ReplaceAllString(text, "/")
}

// Use wraps a middleware to an defaultServeMuxBuilder.
func (b *defaultServeMuxBuilder) Use(middleware ...middleware) {
	b.middlewares = append(b.middlewares, middleware...)
}

// Prefix sets a prefix for the defaultServeMuxBuilder.
func (b *defaultServeMuxBuilder) Prefix(prefix string) {
	// TODO: validate prefix (check if first char is `/`)
	b.patternPrefix = prefix
}

// Subrouter returns an defaultServeMuxBuilder child.
func (b *defaultServeMuxBuilder) Subrouter() *defaultServeMuxBuilder {
	subBuilder := New()
	subBuilder.parent = b
	subBuilder.root = b.root

	if b.root.middlewares != nil && subBuilder != b.root {
		subBuilder.middlewares = append(subBuilder.middlewares, b.root.middlewares...)
	}

	b.subDefaultServeMuxBuilder = append(b.subDefaultServeMuxBuilder, subBuilder)

	return subBuilder
}

// Build fills the given default serve mux with patterns and the connected handler.
//
// It simply calls http.Handle on the patterns and the connected handlers.
func (b *defaultServeMuxBuilder) Build(defaultServeMux *http.ServeMux) []string {
	queue := []*defaultServeMuxBuilder{b}
	visited := make(map[*defaultServeMuxBuilder]bool)
	// TODO: remove when moving to actual pkg
	dataStream := make([]string, 0)
	dataStream = append(dataStream, "Registered Patterns:\n")

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		if current.patterns != nil {
			for pattern, handler := range current.patterns {
				dataStream = append(dataStream, pattern)
				defaultServeMux.Handle(pattern, newHandler(current.middlewares...)(handler))
			}
		}

		queue = append(queue, current.subDefaultServeMuxBuilder...)
	}

	return dataStream
}
