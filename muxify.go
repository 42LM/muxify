// Package muxify implements functionality for building a http.ServeMux.
//
// The muxify package is a default serve mux builder.
// Build patterns, handlers and wrap middlewares conveniently upfront.
// The muxify.ServeMuxBuilder acts as a builder for the http.ServeMux.
// The overall goal of this package is to build the http.ServeMux with pattern/path prefixes and middleware wired in.
package muxify

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// ServeMuxBuilder is a simple builder for the http.ServeMux.
type ServeMuxBuilder struct {
	// Patterns represent the given patterns to http.Handle/http.HandleFunc.
	Patterns map[string]http.Handler
	// PatternPrefix represent the prefix of the pattern of a subrouter.
	PatternPrefix string
	// Middlewares represent the middlewares that wrap the subrouter.
	Middlewares []Middleware
	// Root always points to the root node of the default servce mux builder.
	Root *ServeMuxBuilder
	// Parent always points to the parent node.
	// For the root field the parent would also be root.
	Parent *ServeMuxBuilder

	// SubDefaultServeMuxBuilder stores the subrouters of the main router.
	SubDefaultServeMuxBuilder []*ServeMuxBuilder

	// executedBuild is used to track if the Build() function has been executed.
	executedBuild bool
	// registeredPatterns stores the patterns that have been registered to the default serve mux.
	registeredPatterns []string
}

// Middleware represents an http.Handler wrapper to inject addional functionality.
type Middleware func(http.Handler) http.Handler

// NewServeMuxBuilder returns a new ServeMuxBuilder.
func NewServeMuxBuilder() *ServeMuxBuilder {
	b := &ServeMuxBuilder{Patterns: map[string]http.Handler{}}
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
func (b *ServeMuxBuilder) Pattern(patterns map[string]http.Handler) {
	patternPrefix := b.PatternPrefix
	b.PatternPrefix = ""
	b.PatternPrefix = b.Root.PatternPrefix

	for _, subBuilder := range b.Parent.SubDefaultServeMuxBuilder {
		if b.Parent != b.Root {
			for _, subSubBuilder := range subBuilder.SubDefaultServeMuxBuilder {
				b.PatternPrefix = subSubBuilder.PatternPrefix
			}
		}
	}

	b.PatternPrefix += patternPrefix

	// Using oldschool pattern without method specified like "/a/b/c"
	// This will default to "GET /a/b/c"
	for pattern, handler := range patterns {
		tmpPattern := strings.Split(pattern, " ")

		var method string
		var patternPath string
		switch len(tmpPattern) {
		case 2:
			method = tmpPattern[0] + " "
			patternPath = tmpPattern[1]
		default:
			patternPath = http.MethodGet + " " + tmpPattern[0]
		}

		b.Patterns[method+removeDoubleSlash(b.PatternPrefix+patternPath)] = handler
	}
	b.SubDefaultServeMuxBuilder = append(b.SubDefaultServeMuxBuilder, b)
}

func removeDoubleSlash(text string) string {
	re := regexp.MustCompile(`//+`)
	return re.ReplaceAllString(text, "/")
}

// Use wraps a middleware to an ServeMuxBuilder.
func (b *ServeMuxBuilder) Use(middleware ...Middleware) {
	b.Middlewares = append(b.Middlewares, middleware...)
}

// Prefix sets a prefix for the ServeMuxBuilder.
func (b *ServeMuxBuilder) Prefix(prefix string) {
	if len(prefix) > 0 {
		if prefix[0] != '/' {
			prefix = "/" + prefix
		}
	}

	b.PatternPrefix = prefix
}

// Subrouter returns an ServeMuxBuilder child.
func (b *ServeMuxBuilder) Subrouter() *ServeMuxBuilder {
	subBuilder := NewServeMuxBuilder()
	subBuilder.Parent = b
	subBuilder.Root = b.Root

	if b.Root.Middlewares != nil && subBuilder != b.Root {
		subBuilder.Middlewares = append(subBuilder.Middlewares, b.Root.Middlewares...)
	}

	b.SubDefaultServeMuxBuilder = append(b.SubDefaultServeMuxBuilder, subBuilder)

	return subBuilder
}

// PrintRegisteredPatterns prints the registered patterns of the http.ServeMux.
// The Build() method needs to be called before!
func (b *ServeMuxBuilder) PrintRegisteredPatterns() {
	if b.Root.executedBuild {
		fmt.Println("* Registered patterns:", strings.Repeat("*", 47))
		fmt.Println(strings.Join(b.Root.registeredPatterns, "\n"))
		fmt.Printf("%s\n\n", strings.Repeat("*", 70))
	}
}

// Build constructs an http.ServeMux with the patterns, handlers and middlewares
// from the ServeMuxBuilder.
//
// Always builds from root ServeMuxBuilder node.
func (b *ServeMuxBuilder) Build() *http.ServeMux {
	defaultServeMux := http.ServeMux{}
	queue := []*ServeMuxBuilder{b.Root}
	visited := make(map[*ServeMuxBuilder]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		if current.Patterns != nil {
			for pattern, handler := range current.Patterns {
				b.Root.registeredPatterns = append(b.Root.registeredPatterns, pattern)
				defaultServeMux.Handle(pattern, newHandler(current.Middlewares...)(handler))
			}
		}

		queue = append(queue, current.SubDefaultServeMuxBuilder...)
	}

	b.Root.executedBuild = true
	return &defaultServeMux
}
