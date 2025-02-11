# muxify
[![test: windows](https://github.com/42LM/muxify/actions/workflows/test-windows.yaml/badge.svg)](https://github.com/42LM/muxify/actions/workflows/test-windows.yaml)
[![test: ubuntu/macos](https://github.com/42LM/muxify/actions/workflows/test-ubuntu-macos.yaml/badge.svg)](https://github.com/42LM/muxify/actions/workflows/test-ubuntu-macos.yaml)
[![codecov](https://codecov.io/gh/42LM/muxify/graph/badge.svg?token=6CIY6SU7MJ)](https://codecov.io/gh/42LM/muxify)
[![](https://godoc.org/github.com/42LM/muxify?status.svg)](http://godoc.org/github.com/42LM/muxify)

<img width="150" alt="muxify" src="https://github.com/user-attachments/assets/5b1d6123-55c9-4e3f-81ee-51ffbea3f9d5" />
<br>
<br>

🪄 _**muxify**_ your mux setup in Go! Tiny package for easily building and enhancing the [`http.ServeMux`](https://pkg.go.dev/net/http#ServeMux) with middlewares, subrouters and prefixes.

Make the default `http.ServeMux` setup a breeze in Go 😌!

> [!CAUTION]
> 🚧 Work in progess 🚧

---

* [Install](#install)
* [Example](#example)
* [Motivation](#motivation)

> More examples can be found in the [wiki](https://github.com/42LM/muxify/wiki/Examples)

---

## Install
```sh
go get github.com/42LM/muxify
```

## Example
_**muxify**_ slightly adopts the syntax of [gorilla/mux](https://github.com/gorilla/mux).
It uses a common building block to create a router/subrouter for the serve mux builder.

It all starts with creating the `muxify.ServeMuxBuilder`
```go
muxb := muxify.NewServeMuxBuilder()
```

Setup the router
```go
muxb.Pattern(map[string]http.Handler{
    "GET /": notFoundHandler,
})
```

Create a subrouter from the root router (prefix and middleware are optional)
```go
subMuxb := muxb.Subrouter()
subMuxb.Use(AdminMiddleware, ChorsMiddleware)
subMuxb.Prefix("admin")
subMuxb.Pattern(map[string]http.Handler{
    "POST /{id}": createAdminHandler,
    "DELETE /{id}": deleteAdminHandler,
})
```

Build the default http serve mux (always uses root `ServeMuxBuilder` node)
```go
mux := mux.Build()
```

Use it as usual
```go
s := http.Server{
    Addr:    ":8080",
    Handler: mux,
}

s.ListenAndServe()
```

> [!TIP]
> Check out the registered patterns?
> > ⚠️ Make sure the `Build()` function is called before printing.
> ```go
> muxb.PrintRegisteredPatterns()
> ```
>
> Hm 🤔 `map[string]http.Handler` daaamn ew... 🤢  
> Why don't define your own type?
> ```go
> type P map[string]http.Handler
>
> muxb.Pattern(P{
>     "GET /a": handlerA,
>     "GET /b": handlerB,
> })
> ```

## Motivation
First of all this project exists for the sake of actually using the golang http default serve mux <3.

The motivation for this project derives from the following two problems with the enhanced routing patterns for the `http.ServeMux`:

### 1. Every single handler needs to be wrapped with middleware. This leads to alot of repeating code and moreover to very unreadable code, too. IMHO it already starts to get out of hands when one handler needs to be wrapped with more than four middlewares.

> To give a little bit more context on this topic just take a look at the following code example:
> ```go
> mux.Handle("/foo", Middleware1(Middleware2(Middleware3(Middleware4(Middleware5(Middleware6(fooHandler)))))))
> ```
> So even for middlewares that maybe every handler should have (e.g. auth) this is pretty cumbersome to wrap every single handler in it.
>
> 💡 _**muxify**_ provides a convenient way of wrapping patterns/routes with middleware and subrouters take over these middlewares.

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
> 💡 _**muxify**_ enables the possibility of defining subrouters.
