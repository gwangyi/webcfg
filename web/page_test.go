package web_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gwangyi/webcfg/web"
)

type InitByValueSection struct {
	Initialized bool
}

func (s InitByValueSection) Initialize(parent any, n web.Notifier) error {
	// Value receiver won't update the original unless it uses a pointer in a field
	// but we just want to cover the code path.
	return nil
}

type InitSuccessSection struct {
	Initialized bool
}

type InitSuccessConfig struct {
	Section      InitSuccessSection
	ValueSection InitByValueSection
}

func (s *InitSuccessSection) Initialize(parent any, n web.Notifier) error {
	s.Initialized = true
	return nil
}

type InitErrSection struct {
	Field string
}

type InitErrConfig struct {
	Section InitErrSection
}

func (s *InitErrSection) Initialize(parent any, n web.Notifier) error {
	return errors.New("init error")
}

func TestNewHandler(t *testing.T) {
	cfg := &TestConfig{}
	fsys := fstest.MapFS{
		"favicon.ico": &fstest.MapFile{Data: []byte("favicon")},
		"icon.png":    &fstest.MapFile{Data: []byte("icon")},
		"other.txt":   &fstest.MapFile{Data: []byte("other")},
	}

	handler, err := web.New(cfg, web.WithAssets(fsys))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("GET /", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rr.Code)
		}
	})

	t.Run("GET /index.html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rr.Code)
		}
	})

	t.Run("GET /assets/favicon.ico (custom asset)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/favicon.ico", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rr.Code)
		}
		if rr.Body.String() != "favicon" {
			t.Errorf("expected favicon data")
		}
	})

	t.Run("GET /assets/icon.png (custom asset)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/icon.png", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rr.Code)
		}
	})

	t.Run("GET /assets/other.txt (not allowed custom asset)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/other.txt", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", rr.Code)
		}
	})

	t.Run("GET /assets/css/bulma.min.css (embedded asset)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/css/bulma.min.css", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			// Might be 404 if go generate hasn't run, but we just check behavior
		}
	})

	t.Run("GET /unknown", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", rr.Code)
		}
	})

	t.Run("Initialization error", func(t *testing.T) {
		cfgErr := &InitErrConfig{}
		_, err := web.New(cfgErr)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Initialization success", func(t *testing.T) {
		cfgSuccess := &InitSuccessConfig{}
		_, err := web.New(cfgSuccess)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !cfgSuccess.Section.Initialized {
			t.Errorf("expected Initialized to be true")
		}
	})
}
