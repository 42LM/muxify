<p align="center">
  <img width="225" alt="XXSMuX2" src="https://github.com/user-attachments/assets/5b1d6123-55c9-4e3f-81ee-51ffbea3f9d5" />
</p>

<br>

<div align="center">

  ![example3](https://github.com/42LM/xxsmux/actions/workflows/test.yaml/badge.svg)
  [![](https://godoc.org/github.com/42LM/xxsmux?status.svg)](http://godoc.org/github.com/42LM/xxsmux)

</div>

# xxsmux 🤏
The `xxsmux.defaultServeMuxBuilder` acts as a builder for the `http.DefaultServeMux`.

The overall goal of this package is to build the `http.DefaultServeMux` with pattern/path prefixes and middleware wired in.

The aim is to have a very small helper pkg that makes the use of the go [`http.DefaultServeMux`](https://pkg.go.dev/net/http#DefaultServeMux) easier to use.

> [!CAUTION]
> 🚧 Work in progess 🚧
>
> Only works for go version `^1.22`.
> > For more info: Go 1.22 introduced [enhanced routing patterns](https://tip.golang.org/doc/go1.22#enhanced_routing_patterns)

## Usage
```sh
go get github.com/42LM/xxsmux
```

### Example
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/42LM/xxsmux"
)

func main() {
	router := xxsmux.New()
	router.Use(Middleware1)

	// /hello/{name}
	// /a
	// /b
	router.Pattern(map[string]http.Handler{
		"GET /hello/{name}": http.HandlerFunc(greet),
		"GET /foo":          http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("foo")) }),
		"GET /bar":          http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("bar")) }),
	})

	// /v1/foobar
	v1Router := router.Subrouter()
	v1Router.Prefix("v1")
	v1Router.Pattern(map[string]http.Handler{
		"GET /foobar": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("foobar")) }),
	})

	// /barfoo
	v2RouterNoPrefix := router.Subrouter()
	v2RouterNoPrefix.Pattern(map[string]http.Handler{
		"GET /barfoo": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("barfoo")) }),
	})

	// /v1/admin/secret/{name}
	adminRouter := v1Router.Subrouter()
	adminRouter.Use(AdminMiddleware)
	adminRouter.Prefix("admin")
	adminRouter.Pattern(map[string]http.Handler{
		"GET /secret/{name}": http.HandlerFunc(greet),
	})

	// build the default serve mux aka
	// fill it with path patterns and the additional handlers
	defaultServeMux := router.Build()

	s := http.Server{
		Addr:    ":8080",
		Handler: defaultServeMux,
	}

	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("internal server error: %v", err)
	}
}

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("url.Path: %v\n", r.URL.Path)
	fmt.Printf("url.RawPath: %v\n", r.URL.RawPath)
	fmt.Printf("url.EscapedPath(): %v\n", r.URL.EscapedPath())
	name := r.PathValue("name")
	fmt.Fprintf(w, "Hello %s", name)
}

func Middleware1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "HELLO FROM MIDDLEWARE #1")

		next.ServeHTTP(w, r)
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "HELLO FROM ADMIN MIDDLEWARE")

		usr, _, ok := r.BasicAuth()
		if !ok {
			fmt.Fprintln(w, "⚠️ RESTRICTED AREA")
			return
		}
		if usr == "007" {
			next.ServeHTTP(w, r)
		} else {
			fmt.Fprintln(w, "AGENT WHO??? 🤣")
			return
		}
	})
}
```

### Example curl requests
Example curl requests for the above example.

```sh
curl localhost:8080/hello/max
```
```sh
curl localhost:8080/a
curl localhost:8080/b
```
```sh
curl localhost:8080/v1/foobar
curl localhost:8080/barfoo
```
```sh
curl localhost:8080/v1/admin/secret/max -u
```
> [!TIP]
> When asked for the password just enter. The password is not checked.

## Motivation
The motivation for this project derives from the following two problems with the enhanced routing patterns for the `http.DefaultServeMux`:

### 1. Every single handler needs to be wrapped with middleware. This leads to alot of repeating code and moreover to very unreadable code, too. IMHO it already starts to get out of hands when one handler needs to be wrapped with more than four middlewares.

> To give a little bit more context on this topic just take a look at the following code example:
> ```go
> mux.Handle("/foo", Middleware1(Middleware2(Middleware3(Middleware4(Middleware5(Middleware6(fooHandler)))))))
> ```
> So even for middlewares that maybe every handler should have (e.g. auth) this is pretty cumbersome to wrap every single handler in it.
>
> 💡 **XXSMuX** provides a convenient way of wrapping patterns/routes with middleware and subrouters take over these middlewares.

### 2. No subrouter functionality.

> It is not possible to use the `http.StripPrefix` without defining a pattern for the handler, but sometimes i want to just create a new subrouter from whatever router state.
>```go
> router.Handle("GET /ping/", makePingHandler(endpoints, options))
>
> subrouterV1 := http.NewServeMux()
> subrouterV1.Handle("/v1/", http.StripPrefix("/v1", router))
> ```
> Not being able to use a subrouter adds up to the other problem.
> A subrouter would help wrapping certain patterns/routes with middleware. A subrouter being created from another router/subrouter always inherits the middlewares.
>
> 💡 **XXSMuX** enables the possibility of defining subrouters.
