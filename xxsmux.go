// Package xxsmux implements functionality for building http.DefaultServeMux.
//
// The xxsmux package is a default serve mux builder.
// Build patterns, handlers and wrap middlewares conveniently upfront.
// The xxsmux.DefaultServeMuxBuilder acts as a builder for the http.DefaultServeMux.
// The overall goal of this package is to build the http.DefaultServeMux with pattern/path prefixes and middleware wired in.
package xxsmux

import (
	"net/http"
	"regexp"
	"strings"
)

// DefaultServeMuxBuilder is a simple builder for the http.DefaultServeMux.
type DefaultServeMuxBuilder struct {
	// Patterns represent the given patterns to `http.Handle`/`http.HandleFunc`.
	Patterns map[string]http.Handler
	// PatternPrefix represent the prefix of the pattern of a subrouter.
	PatternPrefix string
	// Middlewares represent the middlewares that wrap the subrouter.
	Middlewares []Middleware
	// Root always points to the root node of the default servce mux builder.
	Root *DefaultServeMuxBuilder
	// Parent always points to the parent node.
	// For the `root` field the parent would also be `root`.
	Parent *DefaultServeMuxBuilder

	// SubDefaultServeMuxBuilder stores the subrouters of the main router.
	SubDefaultServeMuxBuilder []*DefaultServeMuxBuilder
}

// Middleware represents an http.Handler wrapper to inject addional functionality.
type Middleware func(http.Handler) http.Handler

// New returns a new DefaultServeMuxBuilder.
func New() *DefaultServeMuxBuilder {
	b := &DefaultServeMuxBuilder{Patterns: map[string]http.Handler{}}
	b.Root = b
	b.Parent = b
	return b
}

// newHandler returns an http.Handler wrapped with given middlewares.
func newHandler(mw ...Middleware) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		for _, m := range mw {
			h = m(h)
		}
		return h
	}
}

// Pattern registers hanglers for given patterns.
func (b *DefaultServeMuxBuilder) Pattern(patterns map[string]http.Handler) {
	patternPrefix := b.PatternPrefix
	b.PatternPrefix = ""
	b.PatternPrefix = b.Root.PatternPrefix + "/"

	for _, subBuilder := range b.Parent.SubDefaultServeMuxBuilder {
		if b.Parent == b.Root {
			if subBuilder == b {
				for _, subSubBuilder := range subBuilder.SubDefaultServeMuxBuilder {
					b.PatternPrefix = subSubBuilder.PatternPrefix + "/"
				}
			}
		} else {
			for _, subSubBuilder := range subBuilder.SubDefaultServeMuxBuilder {
				b.PatternPrefix = subSubBuilder.PatternPrefix + "/"
			}
		}
	}

	b.PatternPrefix += patternPrefix

	for pattern, handler := range patterns {
		tmpPattern := strings.Split(pattern, " ")

		var method string
		var patternPath string
		switch len(tmpPattern) {
		case 2:
			method = tmpPattern[0] + " "
			patternPath = tmpPattern[1]
		default:
			patternPath = tmpPattern[0]
		}

		b.Patterns[method+removeDoubleSlash(b.PatternPrefix+patternPath)] = handler
	}
	b.SubDefaultServeMuxBuilder = append(b.SubDefaultServeMuxBuilder, b)
}

// removeDoubleSlash cleans up strings for double slashes `//`.
func removeDoubleSlash(text string) string {
	re := regexp.MustCompile(`//+`)
	return re.ReplaceAllString(text, "/")
}

// Use wraps a middleware to an DefaultServeMuxBuilder.
func (b *DefaultServeMuxBuilder) Use(middleware ...Middleware) {
	b.Middlewares = append(b.Middlewares, middleware...)
}

// Prefix sets a prefix for the DefaultServeMuxBuilder.
func (b *DefaultServeMuxBuilder) Prefix(prefix string) {
	// TODO: validate prefix (check if first char is `/`)
	b.PatternPrefix = prefix
}

// Subrouter returns an DefaultServeMuxBuilder child.
func (b *DefaultServeMuxBuilder) Subrouter() *DefaultServeMuxBuilder {
	subBuilder := New()
	subBuilder.Parent = b
	subBuilder.Root = b.Root

	if b.Root.Middlewares != nil && subBuilder != b.Root {
		subBuilder.Middlewares = append(subBuilder.Middlewares, b.Root.Middlewares...)
	}

	b.SubDefaultServeMuxBuilder = append(b.SubDefaultServeMuxBuilder, subBuilder)

	return subBuilder
}

// Build constructs an http.ServeMux with the patterns, handlers and middlewares
// from the DefaultServeMuxBuilder.
func (b *DefaultServeMuxBuilder) Build() *http.ServeMux {
	defaultServeMux := http.ServeMux{}
	queue := []*DefaultServeMuxBuilder{b}
	visited := make(map[*DefaultServeMuxBuilder]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		if current.Patterns != nil {
			for pattern, handler := range current.Patterns {
				defaultServeMux.Handle(pattern, newHandler(current.Middlewares...)(handler))
			}
		}

		queue = append(queue, current.SubDefaultServeMuxBuilder...)
	}

	return &defaultServeMux
}
