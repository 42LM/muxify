# xxsmux
The `xxsmux.defaultServeMuxBuilder` acts as a builder for the `http.DefaultServeMux`.

The overall goal of this package is to build the `http.DefaultServeMux` with pattern/path prefixes and middleware wired in.

> [!CAUTION]
> 🚧 Work in progess 🚧

## Dev
```sh
go run examples/main.go
```

## Example curl requests
```sh
curl localhost:8080/v1/test/wolverine
```
```sh
curl localhost:8080/v1/v2/123/foo
```
```sh
curl localhost:8080/v1/b
```
```sh
curl localhost:8080/v1/secret -u 007
```
> [!TIP]
> When asked for the password enter `martini`:
> ```
> Enter host password for user '007': martini
> ```
