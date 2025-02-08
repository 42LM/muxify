package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// XXSMux is a simple builder for the http.DefaultServeMux.
type XXSMux struct {
	patterns      map[string]http.Handler
	patternPrefix string
	middlewares   []Middleware
	root          *XXSMux
	parent        *XXSMux

	subXXSMux []*XXSMux
}

// Middleware represents an http.Handler wrapper to inject addional functionality.
type Middleware func(http.Handler) http.Handler

// NewXXSMux returns a new XXSMux.
func NewXXSMux() *XXSMux {
	mux := &XXSMux{patterns: map[string]http.Handler{}}
	mux.root = mux
	mux.parent = mux
	return mux
}

// NewHandler returns an http.Handler wrapped with given middlewares.
func NewHandler(mw ...Middleware) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		next := h
		for _, m := range mw {
			next = m(next)
		}
		return next
	}
}

// Pattern registers hanglers for given patterns.
func (mux *XXSMux) Pattern(patterns map[string]http.Handler) {
	patternPrefix := mux.patternPrefix
	mux.patternPrefix = ""
	mux.patternPrefix = mux.root.patternPrefix + "/"

	for _, subMux := range mux.parent.subXXSMux {
		if mux.parent == mux.root {
			if subMux == mux {
				for _, subSubMux := range subMux.subXXSMux {
					mux.patternPrefix = subSubMux.patternPrefix + "/"
				}
			}
		} else {
			for _, subSubMux := range subMux.subXXSMux {
				mux.patternPrefix = subSubMux.patternPrefix + "/"
			}
		}
	}

	mux.patternPrefix += patternPrefix

	for pattern, handler := range patterns {
		// TODO: strings.Split could fail and not have 2 elements
		mux.patterns[removeDoubleSlash(mux.patternPrefix+strings.Split(pattern, " ")[1])] = handler
	}
	mux.subXXSMux = append(mux.subXXSMux, mux)
}

// removeDoubleSlash cleans up strings for double slashes `//`.
func removeDoubleSlash(text string) string {
	re := regexp.MustCompile(`//+`)
	return re.ReplaceAllString(text, "/")
}

// Use wraps a middleware to an XXSMux.
func (mux *XXSMux) Use(middleware ...Middleware) {
	mux.middlewares = append(mux.middlewares, middleware...)
}

// Prefix sets a prefix for the XXSMux.
func (mux *XXSMux) Prefix(prefix string) {
	// TODO: validate prefix (check if first char is `/`)
	mux.patternPrefix = prefix
}

// Subrouter returns an XXSMux child.
func (mux *XXSMux) Subrouter() *XXSMux {
	subMux := NewXXSMux()
	subMux.parent = mux
	subMux.root = mux.root

	if mux.root.middlewares != nil && subMux != mux.root {
		subMux.middlewares = append(subMux.middlewares, mux.root.middlewares...)
	}

	mux.subXXSMux = append(mux.subXXSMux, subMux)

	return subMux
}

// Build fills the given default serve mux with patterns and the connected handler.
//
// It simply calls http.Handle on the patterns and the connected handlers.
func (mux *XXSMux) Build(defaultServeMux *http.ServeMux) []string {
	queue := []*XXSMux{mux}
	visited := make(map[*XXSMux]bool)
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
				dataStream = append(dataStream, fmt.Sprintf("%s", pattern))
				defaultServeMux.Handle(pattern, NewHandler(current.middlewares...)(handler))
			}
		}

		queue = append(queue, current.subXXSMux...)
	}

	return dataStream
}

func main() {
	router := NewXXSMux()
	router.Use(Middleware1, Middleware4)

	// /v1/test
	// /v1/a
	// /v1/b
	router.Prefix("/v1")
	router.Pattern(map[string]http.Handler{
		"GET /test/{name}": http.HandlerFunc(greet),
		"GET /a":           http.HandlerFunc(helloWorld),
		"GET /b":           http.HandlerFunc(helloWorld),
	})

	// /v1/v2/{instance_id}/test
	v1Router := router.Subrouter()
	v1Router.Prefix("v2/{instance_id}")
	v1Router.Pattern(map[string]http.Handler{
		"GET /test": http.HandlerFunc(helloWorld),
	})

	// /v1/v2/{instance_id}/foo
	v12Router := v1Router.Subrouter()
	v12Router.Use(Middleware3)
	// v12Router.Prefix("")
	v12Router.Pattern(map[string]http.Handler{
		"GET /foo": http.HandlerFunc(helloWorld),
	})

	// /v1/v2/{instance_id}/foobar/foo
	v13Router := v12Router.Subrouter()
	v13Router.Use(Middleware3)
	v13Router.Prefix("foobar")
	v13Router.Pattern(map[string]http.Handler{
		"GET /bar": http.HandlerFunc(helloWorld),
	})

	// /v1/boo/test
	v2Router := router.Subrouter()
	v2Router.Prefix("boo")

	v2Router.Pattern(map[string]http.Handler{
		"GET /test": http.HandlerFunc(helloWorld),
	})
	v2Router.Use(Middleware2)

	// /v1/secret
	adminRouter := router.Subrouter()
	adminRouter.Use(AdminMiddleware)
	// adminRouter.Prefix("")
	adminRouter.Pattern(map[string]http.Handler{
		"GET /secret": http.HandlerFunc(secret),
	})

	defaultServeMux := http.DefaultServeMux

	// build the default serve mux aka
	// fill it with path patterns and the additional handlers
	dataStream := router.Build(defaultServeMux)

	fmt.Println(strings.Join(dataStream, "\n"))

	s := http.Server{
		Addr:    ":8080",
		Handler: defaultServeMux,
	}

	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("internal server error: %v", err)
	}
}

// dev test setup
// TODO: remove when separating main from package xxsmux

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("url.Path: %v\n", r.URL.Path)
	fmt.Printf("url.RawPath: %v\n", r.URL.RawPath)
	fmt.Printf("url.EscapedPath(): %v\n", r.URL.EscapedPath())
	name := r.PathValue("name")
	fmt.Fprintf(w, "Hello %s", name)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("url.Path: %v\n", r.URL.Path)
	fmt.Printf("url.RawPath: %v\n", r.URL.RawPath)
	fmt.Printf("url.EscapedPath(): %v\n", r.URL.EscapedPath())
	for range 7 {
		fmt.Fprint(w, "Hello world")
	}
}

func secret(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("url.Path: %v\n", r.URL.Path)
	fmt.Printf("url.RawPath: %v\n", r.URL.RawPath)
	fmt.Printf("url.EscapedPath(): %v\n", r.URL.EscapedPath())
	fmt.Fprintln(w, "Beep Boop Bob hello agent")
}

func Middleware1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "HELLO FROM MIDDLEWARE #1")

		next.ServeHTTP(w, r)
	})
}

func Middleware2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "HELLO FROM MIDDLEWARE #2")

		next.ServeHTTP(w, r)
	})
}

func Middleware3(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "HELLO FROM MIDDLEWARE #3")

		next.ServeHTTP(w, r)
	})
}

func Middleware4(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "HELLO FROM MIDDLEWARE #4")

		next.ServeHTTP(w, r)
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "HELLO FROM ADMIN MIDDLEWARE")

		usr, pw, ok := r.BasicAuth()
		if !ok {
			fmt.Fprintln(w, "⚠️ RESTRICTED AREA")
			return
		}
		if usr == "007" && pw == "martini" {
			next.ServeHTTP(w, r)
		} else {
			fmt.Fprintln(w, "AGENT WHO??? 🤣")
			return
		}
	})
}
