<p align="center">
  <img width="300" alt="XXSMuX2" src="https://github.com/user-attachments/assets/eaf2317b-8edf-4331-8dec-3c54976a865e" />
</p>

# xxsmux
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
	defaultServeMux := http.DefaultServeMux
	router.Build(defaultServeMux)

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
