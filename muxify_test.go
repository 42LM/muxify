//go:build go1.22
// +build go1.22

// TODO: refactor tests
package muxify_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/42LM/muxify"
)

func Test_Bootstrap(t *testing.T) {
	testCases := map[string]struct {
		path          string
		middleware    [](func(http.Handler) http.Handler)
		method        string
		expBody       string
		expStatusCode int
	}{
		"ok - no middleware": {
			path:          "/a/test/luke",
			method:        http.MethodGet,
			expBody:       "hello luke",
			expStatusCode: http.StatusOK,
		},
		"ok - with middleware": {
			path:          "/a/test/luke",
			middleware:    [](func(http.Handler) http.Handler){testMiddleware1},
			method:        http.MethodGet,
			expBody:       "MW1:hello luke",
			expStatusCode: http.StatusOK,
		},
		"ok - with multiple middleware": {
			path:          "/a/test/luke",
			middleware:    [](func(http.Handler) http.Handler){testMiddleware1, testMiddleware2},
			method:        http.MethodGet,
			expBody:       "MW1:MW2:hello luke",
			expStatusCode: http.StatusOK,
		},
		"post with id": {
			path:          "/a/b/e/123",
			method:        http.MethodPost,
			expBody:       "POST id: 123",
			expStatusCode: http.StatusOK,
		},
		"delete with id": {
			path:          "/a/b/e/123",
			method:        http.MethodDelete,
			expBody:       "DELETE id: 123",
			expStatusCode: http.StatusOK,
		},
		"get with id (test remove double slashes)": {
			path:          "/a/b/e/d/f/123",
			method:        http.MethodGet,
			expBody:       "GET id: 123",
			expStatusCode: http.StatusOK,
		},
		"notfound /": {
			path:          "/",
			method:        http.MethodGet,
			expBody:       "not found",
			expStatusCode: http.StatusNotFound,
		},
		"notfound /random/path": {
			path:          "/random/path",
			method:        http.MethodGet,
			expBody:       "not found",
			expStatusCode: http.StatusNotFound,
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			// create the default service mux builder
			mux := muxify.NewMux()

			// apply some middleware
			if tc.middleware != nil {
				for _, mw := range tc.middleware {
					mux.Use(mw)
				}
			}

			mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
			})

			mux.Prefix("/a")
			mux.HandleFunc("GET /test/{name}", func(w http.ResponseWriter, r *http.Request) {
				name := r.PathValue("name")
				_, _ = w.Write([]byte("hello " + name))
			})

			subMux1 := mux.Subrouter()
			subMux1.Prefix("/b")
			subMux1.HandleFunc("POST /e/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := r.PathValue("id")
				_, _ = w.Write([]byte("POST id: " + id))
			})
			subMux1.HandleFunc("DELETE /e/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := r.PathValue("id")
				_, _ = w.Write([]byte("DELETE id: " + id))
			})
			subMux1.Handle("GET /e/////d///f//{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id := r.PathValue("id")
				_, _ = w.Write([]byte("GET id: " + id))
			}))

			server := httptest.NewServer(mux)
			defer server.Close()

			// perform some requests
			req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expStatusCode {
				t.Errorf("\nwant: %v\ngot: %v\n", tc.expStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			got := string(body)
			if got != tc.expBody {
				t.Errorf("\nwant: %v\ngot: %v\n", tc.expBody, got)
			}
		})
	}
}

func Test_MuxWithSubrouters_MiddlewareChaining(t *testing.T) {
	testCases := map[string]struct {
		path          string
		method        string
		middlewareB1  [](func(http.Handler) http.Handler)
		middlewareB2  [](func(http.Handler) http.Handler)
		middlewareB3  [](func(http.Handler) http.Handler)
		middlewareB4  [](func(http.Handler) http.Handler)
		expBody       string
		expStatusCode int
	}{
		"ok - no middleware": {
			path:          "/a/foo",
			method:        http.MethodGet,
			expBody:       "foo",
			expStatusCode: http.StatusOK,
		},
		"ok - with middlewareB1": {
			path:          "/a/foo",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			expBody:       "MW1:foo",
			expStatusCode: http.StatusOK,
		},
		"ok - with multiple middlewareB2": {
			path:          "/a/b/bar",
			method:        http.MethodGet,
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2},
			expBody:       "MW2:bar",
			expStatusCode: http.StatusOK,
		},
		"ok - with multiple middlewareB1 and middlewareB2 - foo": {
			path:          "/a/foo",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2},
			expBody:       "MW1:foo",
			expStatusCode: http.StatusOK,
		},
		"ok - with multiple middlewareB1 and middlewareB2 - bar": {
			path:          "/a/b/bar",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2},
			expBody:       "MW1:MW2:bar",
			expStatusCode: http.StatusOK,
		},
		"ok - with multiple middlewareB1, middlewareB2 and middlewareB3 - foobar": {
			path:          "/a/b/c/foobar",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2},
			middlewareB3:  [](func(http.Handler) http.Handler){testMiddleware3},
			expBody:       "MW1:MW2:MW3:foobar",
			expStatusCode: http.StatusOK,
		},
		"ok - with multiple middlewareB1, middlewareB2, middlewareB3 and middlewareB4 - barfoot": {
			path:          "/a/b/c/d/barfoot",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2},
			middlewareB3:  [](func(http.Handler) http.Handler){testMiddleware3},
			middlewareB4:  [](func(http.Handler) http.Handler){testMiddleware4},
			expBody:       "MW1:MW2:MW3:MW4:barfoot",
			expStatusCode: http.StatusOK,
		},
		"ok - no content": {
			path:          "/",
			method:        http.MethodGet,
			expStatusCode: http.StatusNoContent,
		},
		"ok - OPTIONS - no content": {
			path:          "/",
			method:        http.MethodOptions,
			expStatusCode: http.StatusNoContent,
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			mux := muxify.NewMux()
			if tc.middlewareB1 != nil {
				for _, mw := range tc.middlewareB1 {
					mux.Use(mw)
				}
			}

			mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				_, _ = w.Write([]byte("foo" + r.URL.Path))
			})
			mux.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				_, _ = w.Write([]byte("foo"))
			})
			mux.Prefix("/a")
			mux.HandleFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("foo"))
			})

			subMux1 := mux.Subrouter()
			if tc.middlewareB2 != nil {
				for _, mw := range tc.middlewareB2 {
					subMux1.Use(mw)
				}
			}

			subMux1.Prefix("/b").HandleFunc("GET /bar", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("bar"))
			})

			subMux2 := subMux1.Subrouter()
			if tc.middlewareB3 != nil {
				for _, mw := range tc.middlewareB3 {
					subMux2.Use(mw)
				}
			}
			subMux2.Prefix("/c")
			subMux2.HandleFunc("GET /foobar", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("foobar"))
			})

			subMux3 := subMux2.Subrouter()
			if tc.middlewareB4 != nil {
				for _, mw := range tc.middlewareB4 {
					subMux3.Use(mw)
				}
			}
			subMux3.Prefix("/d")
			subMux3.HandleFunc("GET /barfoot", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("barfoot"))
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			client := &http.Client{}

			// perform some requests
			req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expStatusCode {
				t.Errorf("\nwant: %v\ngot: %v\n", tc.expStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			got := string(body)
			if got != tc.expBody {
				t.Errorf("\nwant: %v\ngot: %v\n", tc.expBody, got)
			}
		})
	}
}

func Test_MuxWithSubrouters(t *testing.T) {
	testCases := map[string]struct {
		path          string
		method        string
		middlewareB1  [](func(http.Handler) http.Handler)
		middlewareB2  [](func(http.Handler) http.Handler)
		middlewareB4  [](func(http.Handler) http.Handler)
		expBody       string
		expStatusCode int
	}{
		"ok - no middleware - /a/foo": {
			path:          "/a/foo",
			method:        http.MethodGet,
			expBody:       "foo",
			expStatusCode: http.StatusOK,
		},
		"ok - no middleware - /a/b/bar": {
			path:          "/a/b/bar",
			method:        http.MethodGet,
			expBody:       "bar",
			expStatusCode: http.StatusOK,
		},
		"ok - no middleware - /a/foobar": {
			path:          "/a/foobar",
			method:        http.MethodGet,
			expBody:       "foobar",
			expStatusCode: http.StatusOK,
		},
		"ok - no middleware - /a/b/d/woo": {
			path:          "/a/b/d/woo",
			method:        http.MethodGet,
			expBody:       "woo",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB1 - /a/foo": {
			path:          "/a/foo",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			expBody:       "MW1:foo",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB1 - /a/b/bar": {
			path:          "/a/b/bar",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			expBody:       "MW1:bar",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB1 - /a/foobar": {
			path:          "/a/foobar",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			expBody:       "MW1:foobar",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB1 - /a/b/d/woo": {
			path:          "/a/b/d/woo",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			expBody:       "MW1:woo",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB1 multiple middlewares - /a/foo": {
			path:          "/a/b/d/woo",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1, testMiddleware2},
			expBody:       "MW1:MW2:woo",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB2 - /a/b/bar": {
			path:          "/a/b/bar",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2, testMiddleware3},
			expBody:       "MW1:MW2:MW3:bar",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB2 - /a/foobar": {
			path:          "/a/foobar",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2, testMiddleware3},
			expBody:       "MW1:foobar",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB2 - /a/b/d/woo": {
			path:          "/a/b/d/woo",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2, testMiddleware3},
			expBody:       "MW1:MW2:MW3:woo",
			expStatusCode: http.StatusOK,
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			mux := muxify.NewMux()
			if tc.middlewareB1 != nil {
				for _, mw := range tc.middlewareB1 {
					mux.Use(mw)
				}
			}
			mux.Prefix("/a")
			mux.HandleFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("foo"))
			})

			subMux1 := mux.Subrouter()
			if tc.middlewareB2 != nil {
				for _, mw := range tc.middlewareB2 {
					subMux1.Use(mw)
				}
			}
			subMux1.Prefix("/b")
			subMux1.HandleFunc("GET /bar", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("bar"))
			})

			subMux2 := mux.Subrouter()
			subMux2.HandleFunc("GET /foobar", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("foobar"))
			})

			subMux3 := subMux1.Subrouter()
			if tc.middlewareB4 != nil {
				for _, mw := range tc.middlewareB4 {
					subMux3.Use(mw)
				}
			}
			subMux3.Prefix("/d")
			subMux3.HandleFunc("GET /woo", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("woo"))
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			client := &http.Client{}

			// perform some requests
			req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expStatusCode {
				t.Errorf("\nwant: %v\ngot: %v\n", tc.expStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			got := string(body)
			if got != tc.expBody {
				t.Errorf("\nwant: %v\ngot: %v\n", tc.expBody, got)
			}
		})
	}
}

func testMiddleware1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("MW1:"))

		next.ServeHTTP(w, r)
	})
}

func testMiddleware2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("MW2:"))

		next.ServeHTTP(w, r)
	})
}

func testMiddleware3(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("MW3:"))

		next.ServeHTTP(w, r)
	})
}

func testMiddleware4(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("MW4:"))

		next.ServeHTTP(w, r)
	})
}
