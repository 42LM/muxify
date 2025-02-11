package muxify_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/42LM/muxify"
)

func Test_Bootstrap_below_122(t *testing.T) {
	testCases := map[string]struct {
		path          string
		middleware    [](func(http.Handler) http.Handler)
		method        string
		expBody       string
		expStatusCode int
	}{
		"ok - no middleware": {
			path:          "/a/test",
			method:        http.MethodGet,
			expBody:       "hello",
			expStatusCode: http.StatusOK,
		},
		"ok - with middleware": {
			path:          "/a/test",
			middleware:    [](func(http.Handler) http.Handler){testMiddleware1},
			method:        http.MethodGet,
			expBody:       "MW1:hello",
			expStatusCode: http.StatusOK,
		},
		"ok - with multiple middleware": {
			path:          "/a/test",
			middleware:    [](func(http.Handler) http.Handler){testMiddleware1, testMiddleware2},
			method:        http.MethodGet,
			expBody:       "MW2:MW1:hello",
			expStatusCode: http.StatusOK,
		},
		"post with id": {
			path:          "/a/b/e",
			method:        http.MethodPost,
			expBody:       "POST",
			expStatusCode: http.StatusOK,
		},
		"get with id (test remove double slashes)": {
			path:          "/a/b/e/d/f",
			method:        http.MethodDelete,
			expBody:       "DELETE",
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
		"no method set (default to GET)": {
			path:          "/oldschool",
			expBody:       "oldschool",
			expStatusCode: http.StatusOK,
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			// create the default service mux builder
			b := muxify.NewServeMuxBuilder()

			// apply some middleware
			if tc.middleware != nil {
				for _, mw := range tc.middleware {
					b.Use(mw)
				}
			}

			b.Pattern(map[string]http.Handler{
				"/": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte("not found"))
				}),
				"/oldschool": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("oldschool"))
				}),
			})

			b.Prefix("/a")
			b.Pattern(map[string]http.Handler{
				"/test": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("hello"))
				}),
			})

			b1 := b.Subrouter()
			b1.Prefix("/b")
			b1.Pattern(map[string]http.Handler{
				"/e": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("POST"))
				}),
				"/e/////d///f//": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("DELETE"))
				}),
			})

			// build http default serve mux
			mux := b.Build()

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
