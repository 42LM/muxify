package muxify

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_newHandler(t *testing.T) {
	testCases := map[string]struct {
		middleware []Middleware
		expBody    string
	}{
		"no middleware": {
			middleware: []Middleware{},
			expBody:    "test response",
		},
		"wrap multiple middlewares": {
			middleware: []Middleware{testMiddleware1, testMiddleware2},
			expBody:    "MW1:MW2:test response",
		},
		"wrap middleware": {
			middleware: []Middleware{testMiddleware3},
			expBody:    "MW3:test response",
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			mux := http.ServeMux{}

			h := newHandler(tc.middleware...)(http.HandlerFunc(testHandler))
			mux.Handle("/test", h)

			server := httptest.NewServer(h)
			defer server.Close()

			resp, err := http.Get(server.URL + "/test")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("\nexpected: %v\ngot: %v\n", http.StatusOK, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			got := string(body)
			if got != tc.expBody {
				t.Errorf("\nexpected: %v\ngot: %v\n", tc.expBody, got)
			}
		})
	}
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("test response"))
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

func Test_Prefix(t *testing.T) {
	testCases := map[string]struct {
		prefix    string
		expPrefix string
	}{
		"ok - no slash prefix": {
			prefix:    "a",
			expPrefix: "/v1/a",
		},
		"ok - slash prefix": {
			prefix:    "/a",
			expPrefix: "/v1/a",
		},
		"ok - let go compiler handle": {
			prefix:    "//a",
			expPrefix: "/v1//a",
		},
		"ok - empty": {
			prefix:    "",
			expPrefix: "/v1",
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			mux := Mux{
				patternPrefix: "/v1",
			}

			mux.Prefix(tc.prefix)

			got := mux.patternPrefix

			if len(tc.prefix) > 0 {
				if tc.prefix[0] != '/' {
					tc.prefix = "/" + tc.prefix
				}
			}

			if got != tc.expPrefix {
				t.Errorf("\nwant: %v\ngot: %v\n", tc.expPrefix, got)
			}
		})
	}
}
