# xxsmux
![example3](https://github.com/42LM/xxsmux/actions/workflows/test.yaml/badge.svg)
[![codecov](https://codecov.io/gh/42LM/xxsmux/graph/badge.svg?token=6CIY6SU7MJ)](https://codecov.io/gh/42LM/xxsmux)
[![](https://godoc.org/github.com/42LM/xxsmux?status.svg)](http://godoc.org/github.com/42LM/xxsmux)

<img width="150" alt="XXSMuX2" src="https://github.com/user-attachments/assets/5b1d6123-55c9-4e3f-81ee-51ffbea3f9d5" />
<br>
<br>

The _**xxsmux**_ package offers a convenient way to build a default [`http.ServeMux`](https://pkg.go.dev/net/http#ServeMux). Build patterns, connect handlers, wrap path prefixes and middlewares upfront. Create subrouters from routers that inherit the middleware and path prefix setup from the parent router.

The purpose of this package is to build a golang default serve mux and is not a standalone multiplexer. It rather helps to create the default serve mux without too much repition in code.

> [!CAUTION]
> 🚧 Work in progess 🚧
>
> Only works for go version above `^1.22`.
> > For more info: Go 1.22 introduced [enhanced routing patterns](https://tip.golang.org/doc/go1.22#enhanced_routing_patterns)

---

* [Install](#install)
* [Example](#example)
* [Motivation](#motivation)

> More examples can be found in the [wiki](https://github.com/42LM/xxsmux/wiki/Examples)

---

## Install
```sh
go get github.com/42LM/xxsmux
```

## Example
_**xxsmux**_ slightly adopts the syntax of [gorilla/mux](https://github.com/gorilla/mux).
It uses a common building block to create a router/subrouter.

It all starts with creating the `xxsmux.DefaultServeMuxBuilder`
```go
router := xxsmux.New()
```

Setup the router (setup prefix, middleware and pattern)
```go
router.Prefix("/v1") // optional
router.Use(AuthMiddleware) // optional
router.Pattern(map[string]http.Handler{
    "GET /hello/{name}": handler,
    "GET /foo": handler,
    "GET /bar": handler,
})
```

Create a subrouter from the root router
```go
subRouter := router.Subrouter()
subRouter.Use(AdminMiddleware, ChorsMiddleware) // optional
subRouter.Prefix("admin") // optional
subRouter.Pattern(map[string]http.Handler{
    "GET /secret": handler,
})
```

Build the default http serve mux
```go
defaultServeMux := router.Build(defaultServeMux)
```

Use it as usual
```go
s := http.Server{
    Addr:    ":8080",
    Handler: defaultServeMux,
}

s.ListenAndServe()
```

## Motivation
First of all this project exists for the sake of actually using the golang http default serve mux <3.

The motivation for this project derives from the following two problems with the enhanced routing patterns for the `http.DefaultServeMux`:

### 1. Every single handler needs to be wrapped with middleware. This leads to alot of repeating code and moreover to very unreadable code, too. IMHO it already starts to get out of hands when one handler needs to be wrapped with more than four middlewares.

> To give a little bit more context on this topic just take a look at the following code example:
> ```go
> mux.Handle("/foo", Middleware1(Middleware2(Middleware3(Middleware4(Middleware5(Middleware6(fooHandler)))))))
> ```
> So even for middlewares that maybe every handler should have (e.g. auth) this is pretty cumbersome to wrap every single handler in it.
>
> 💡 _**xxsmux**_ provides a convenient way of wrapping patterns/routes with middleware and subrouters take over these middlewares.

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
> 💡 _**xxsmux**_ enables the possibility of defining subrouters.
