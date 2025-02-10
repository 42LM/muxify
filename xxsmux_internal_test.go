package xxsmux

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
			expBody:    "MW2:MW1:test response",
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

func Test_removeDoubleSlash(t *testing.T) {
	testCases := map[string]struct {
		text              string
		expRemovedSlashes string
	}{
		"many slashes": {
			text:              "//a/////////b////c//",
			expRemovedSlashes: "/a/b/c/",
		},
		"slashes in between": {
			text:              "a///b///c",
			expRemovedSlashes: "a/b/c",
		},
	}
	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			removedSlashes := removeDoubleSlash(tc.text)
			if removedSlashes != tc.expRemovedSlashes {
				t.Errorf("\nexpected: %v\ngot: %v\n", tc.expRemovedSlashes, removedSlashes)
			}
		})
	}
}
