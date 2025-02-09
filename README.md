<p align="center">
  <img width="300" alt="XXSMuX2" src="https://github.com/user-attachments/assets/eaf2317b-8edf-4331-8dec-3c54976a865e" />
</p>

# xxsmux
The `xxsmux.defaultServeMuxBuilder` acts as a builder for the `http.DefaultServeMux`.

The overall goal of this package is to build the `http.DefaultServeMux` with pattern/path prefixes and middleware wired in.

> [!CAUTION]
> 🚧 Work in progess 🚧
>
> Only works for go version `^1.22`.
> > For more info: Go 1.22 introduced [enhanced routing patterns](https://tip.golang.org/doc/go1.22#enhanced_routing_patterns)

## Dev
```sh
git clone git@github.com:42LM/xxsmux.git
```

```sh
go run examples/main.go
```

```sh
go test ./... -v
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
