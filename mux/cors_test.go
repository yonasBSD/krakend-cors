package mux

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luraproject/lura/v2/logging"
)

func TestInvalidCfg(t *testing.T) {
	sampleCfg := map[string]interface{}{}
	corsMw := New(sampleCfg)
	if corsMw != nil {
		t.Error("The corsMw should be nil.\n")
	}
}

func TestNew(t *testing.T) {
	sampleCfg := map[string]interface{}{}
	serialized := []byte(`{ "github_com/devopsfaith/krakend-cors": {
			"allow_origins": [ "http://foobar.com" ],
			"allow_headers": [ "Origin" ],
			"allow_methods": [ "GET" ],
			"max_age": "2h"
			}
		}`)
	if err := json.Unmarshal(serialized, &sampleCfg); err != nil {
		t.Error(err)
		return
	}
	h := New(sampleCfg)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "https://example.com/foo", http.NoBody)
	req.Header.Add("Origin", "http://foobar.com")
	req.Header.Add("Access-Control-Request-Method", "GET")
	req.Header.Add("Access-Control-Request-Headers", "origin")
	handler := h.Handler(testHandler)
	handler.ServeHTTP(res, req)

	assertHeaders(t, res.Header(), map[string]string{
		"Vary":                         "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
		"Access-Control-Allow-Origin":  "http://foobar.com",
		"Access-Control-Allow-Methods": "GET",
		"Access-Control-Allow-Headers": "origin",
		"Access-Control-Max-Age":       "7200",
	})
}

func TestNewWithLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger, err := logging.NewLogger("DEBUG", buf, "")
	if err != nil {
		t.Error(err)
		return
	}
	sampleCfg := map[string]interface{}{}
	serialized := []byte(`{ "github_com/devopsfaith/krakend-cors": {
			"allow_origins": [ "http://foobar.com" ],
			"allow_methods": [ "GET" ],
			"max_age": "2h"
			}
		}`)
	if err := json.Unmarshal(serialized, &sampleCfg); err != nil {
		t.Error(err)
		return
	}
	h := NewWithLogger(sampleCfg, logger)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "https://example.com/foo", http.NoBody)
	req.Header.Add("Origin", "http://foobar.com")
	handler := h.Handler(testHandler)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Errorf("Invalid status code: %d should be 200", res.Code)
	}

	assertHeaders(t, res.Header(), map[string]string{
		"Vary":                         "Origin",
		"Access-Control-Allow-Origin":  "http://foobar.com",
		"Access-Control-Allow-Methods": "",
		"Access-Control-Allow-Headers": "",
		"Access-Control-Max-Age":       "",
	})

	loggedMsg := buf.String()
	if loggedMsg != "" {
		t.Error("unexpected logged msg:", loggedMsg)
	}
}

func TestAllowOriginEmpty(t *testing.T) {
	sampleCfg := map[string]interface{}{}
	serialized := []byte(`{ "github_com/devopsfaith/krakend-cors": {
			}
		}`)
	if err := json.Unmarshal(serialized, &sampleCfg); err != nil {
		t.Error(err)
		return
	}
	h := New(sampleCfg)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "https://example.com/foo", http.NoBody)
	req.Header.Add("Access-Control-Request-Method", "GET")
	req.Header.Add("Access-Control-Request-Headers", "origin")
	req.Header.Add("Origin", "http://foobar.com")
	handler := h.Handler(testHandler)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Errorf("Invalid status code: %d should be 204", res.Code)
	}

	assertHeaders(t, res.Header(), map[string]string{
		"Vary":                         "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
		"Access-Control-Allow-Headers": "origin",
	})
}

func TestOptionsSuccess(t *testing.T) {
	sampleCfg := map[string]interface{}{}
	serialized := []byte(`{ "github_com/devopsfaith/krakend-cors": {
				"options_success_status": 200
			}
		}`)
	if err := json.Unmarshal(serialized, &sampleCfg); err != nil {
		t.Error(err)
		return
	}
	h := New(sampleCfg)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "https://example.com/foo", http.NoBody)
	req.Header.Add("Access-Control-Request-Method", "GET")
	req.Header.Add("Access-Control-Request-Headers", "origin")
	req.Header.Add("Origin", "http://foobar.com")
	handler := h.Handler(testHandler)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Errorf("Invalid status code: %d should be 200", res.Code)
	}

	assertHeaders(t, res.Header(), map[string]string{
		"Vary":                         "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
		"Access-Control-Allow-Headers": "origin",
	})
}

func TestAllowPrivateNetwork(t *testing.T) {
	sampleCfg := map[string]interface{}{}
	serialized := []byte(`{ "github_com/devopsfaith/krakend-cors": {
				"allow_private_network": true
			}
		}`)
	if err := json.Unmarshal(serialized, &sampleCfg); err != nil {
		t.Error(err)
		return
	}
	h := New(sampleCfg)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "https://example.com/foo", http.NoBody)
	req.Header.Add("Access-Control-Request-Method", "GET")
	req.Header.Add("Access-Control-Request-Private-Network", "true")
	req.Header.Add("Origin", "http://foobar.com")
	handler := h.Handler(testHandler)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Errorf("Invalid status code: %d should be 204", res.Code)
	}

	assertHeaders(t, res.Header(), map[string]string{
		"Vary":                                 "Origin, Access-Control-Request-Method, Access-Control-Request-Headers, Access-Control-Request-Private-Network",
		"Access-Control-Allow-Origin":          "*",
		"Access-Control-Allow-Methods":         "GET",
		"Access-Control-Allow-Private-Network": "true",
	})
}

func TestOptionPasstrough(t *testing.T) {
	sampleCfg := map[string]interface{}{}
	serialized := []byte(`{ "github_com/devopsfaith/krakend-cors": {
				"options_passthrough": true
			}
		}`)

	if err := json.Unmarshal(serialized, &sampleCfg); err != nil {
		t.Error(err)
		return
	}

	h := New(sampleCfg)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "https://example.com/foo", http.NoBody)
	req.Header.Add("Access-Control-Request-Method", "GET")
	req.Header.Add("Origin", "http://foobar.com")
	handler := h.Handler(testHandler)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Errorf("Invalid status code: %d should be 200", res.Code)
	}

	assertHeaders(t, res.Header(), map[string]string{
		"Vary":                         "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
	})
}

var allHeaders = []string{
	"Vary",
	"Access-Control-Allow-Origin",
	"Access-Control-Allow-Methods",
	"Access-Control-Allow-Headers",
	"Access-Control-Allow-Credentials",
	"Access-Control-Max-Age",
	"Access-Control-Expose-Headers",
}

func assertHeaders(t *testing.T, resHeaders http.Header, expHeaders map[string]string) {
	for _, name := range allHeaders {
		got := strings.Join(resHeaders[name], ", ")
		want := expHeaders[name]
		if got != want {
			t.Errorf("Response header %q = %q, want %q", name, got, want)
		}
	}
}

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("bar"))
})
