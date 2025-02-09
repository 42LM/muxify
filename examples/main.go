// Example for execution and dev testing
//
// TODO: remove when releasing, write examples into readme file!
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/42LM/xxsmux"
)

func main() {
	router := xxsmux.New()
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
// TODO: remove when separating main from package defaultServeMuxBuilder

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
