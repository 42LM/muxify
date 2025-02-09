package xxsmux_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/42LM/xxsmux"
)

func Test_Bootstrap(t *testing.T) {
	testCases := map[string]struct {
		path       string
		expBody    string
		middleware [](func(http.Handler) http.Handler)
	}{
		"ok - no middleware": {
			path:    "/a/test/luke",
			expBody: "hello luke",
		},
		"ok - with middleware": {
			path:       "/a/test/luke",
			middleware: [](func(http.Handler) http.Handler){testMiddleware1},
			expBody:    "MW1:hello luke",
		},
		"ok - with multiple middleware": {
			path:       "/a/test/luke",
			middleware: [](func(http.Handler) http.Handler){testMiddleware1, testMiddleware2},
			expBody:    "MW2:MW1:hello luke",
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			// create the default service mux builder
			b := xxsmux.New()

			// apply some middleware
			if tc.middleware != nil {
				for _, mw := range tc.middleware {
					b.Use(mw)
				}
			}

			b.Prefix("/a")
			b.Pattern(map[string]http.Handler{
				"GET /test/{name}": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					name := r.PathValue("name")
					_, _ = w.Write([]byte("hello " + name))
				}),
			})

			// build http default serve mux
			mux := &http.ServeMux{}
			_ = b.Build(mux)

			server := httptest.NewServer(mux)
			defer server.Close()

			// perform some requests
			resp, err := http.Get(server.URL + tc.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("\nexpected: %q\ngot: %q\n", http.StatusOK, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			got := string(body)
			if got != tc.expBody {
				t.Errorf("\nexpected: %q\ngot: %q\n", tc.expBody, got)
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
			expBody:       "MW2:MW1:bar",
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
			b := xxsmux.New()
			if tc.middlewareB1 != nil {
				for _, mw := range tc.middlewareB1 {
					b.Use(mw)
				}
			}

			b.Pattern(map[string]http.Handler{
				"GET /": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
					_, _ = w.Write([]byte("foo" + r.URL.Path))
				}),
				"OPTIONS /": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
					_, _ = w.Write([]byte("foo"))
				}),
			})
			b.Prefix("/a")
			b.Pattern(map[string]http.Handler{
				"GET /foo": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("foo"))
				}),
			})

			b2 := b.Subrouter()
			if tc.middlewareB2 != nil {
				for _, mw := range tc.middlewareB2 {
					b2.Use(mw)
				}
			}

			b2.Prefix("/b")
			b2.Pattern(map[string]http.Handler{
				"GET /bar": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("bar"))
				}),
			})

			// build http default serve mux
			mux := &http.ServeMux{}
			_ = b.Build(mux)

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
				t.Errorf("\nexpected: %d\ngot: %d\n", tc.expStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			got := string(body)
			if got != tc.expBody {
				t.Errorf("\nexpected: %q\ngot: %q\n", tc.expBody, got)
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
			expBody:       "MW2:MW1:woo",
			expStatusCode: http.StatusOK,
		},
		"ok - middlewareB2 - /a/b/bar": {
			path:          "/a/b/bar",
			method:        http.MethodGet,
			middlewareB1:  [](func(http.Handler) http.Handler){testMiddleware1},
			middlewareB2:  [](func(http.Handler) http.Handler){testMiddleware2, testMiddleware3},
			expBody:       "MW3:MW2:MW1:bar",
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
			expBody:       "MW1:woo",
			expStatusCode: http.StatusOK,
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			b := xxsmux.New()
			if tc.middlewareB1 != nil {
				for _, mw := range tc.middlewareB1 {
					b.Use(mw)
				}
			}
			b.Prefix("/a")
			b.Pattern(map[string]http.Handler{
				"GET /foo": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("foo"))
				}),
			})

			b2 := b.Subrouter()
			if tc.middlewareB2 != nil {
				for _, mw := range tc.middlewareB2 {
					b2.Use(mw)
				}
			}
			b2.Prefix("/b")
			b2.Pattern(map[string]http.Handler{
				"GET /bar": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("bar"))
				}),
			})

			b3 := b.Subrouter()
			b3.Pattern(map[string]http.Handler{
				"GET /foobar": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("foobar"))
				}),
			})

			b4 := b2.Subrouter()
			if tc.middlewareB4 != nil {
				for _, mw := range tc.middlewareB4 {
					b4.Use(mw)
				}
			}
			b4.Prefix("/d")
			b4.Pattern(map[string]http.Handler{
				"GET /woo": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("woo"))
				}),
			})

			// build http default serve mux
			mux := &http.ServeMux{}
			_ = b.Build(mux)

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
				t.Errorf("\nexpected: %d\ngot: %d\n", tc.expStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			got := string(body)
			if got != tc.expBody {
				t.Errorf("\nexpected: %q\ngot: %q\n", tc.expBody, got)
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
