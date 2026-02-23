package web_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gwangyi/webcfg/web"
)

type UpdateSuccessSection struct {
	UpdatedCalled bool
}

func (s *UpdateSuccessSection) Updated(parent any, n web.Notifier) error {
	s.UpdatedCalled = true
	return nil
}

type UpdateErrSection struct{}

func (s *UpdateErrSection) Updated(parent any, n web.Notifier) error {
	return errors.New("update error")
}

type UpdateByValueSection struct{}

func (s UpdateByValueSection) Updated(parent any, n web.Notifier) error {
	return nil
}

type UpdateTestConfig struct {
	Section      UpdateSuccessSection
	ErrSection   UpdateErrSection
	ValueSection UpdateByValueSection
}

func TestUpdateReceiver(t *testing.T) {
	cfg := &UpdateTestConfig{}
	handler, _ := web.New(cfg)

	t.Run("Updated called successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/Section", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if !cfg.Section.UpdatedCalled {
			t.Errorf("expected Updated to be called")
		}
	})

	t.Run("Updated returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/ErrSection", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("expected 303 See Other, got %d", rr.Code)
		}
	})

	t.Run("Updated with value receiver", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/ValueSection", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("expected 303 See Other, got %d", rr.Code)
		}
	})
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestPostUpdate(t *testing.T) {
	cfg := &TestConfig{}
	handler, _ := web.New(cfg)

	t.Run("Valid update", func(t *testing.T) {
		form := url.Values{}
		form.Add("string_field", "new value")
		form.Add("BoolField", "on")
		form.Add("IntField", "42")
		form.Add("uint_field", "42")
		form.Add("float_field", "3.14")
		form.Add("custom_field", "custom val")

		req := httptest.NewRequest(http.MethodPost, "/Section1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("expected 303 See Other, got %d", rr.Code)
		}

		if cfg.Section1.StringField != "new value" {
			t.Errorf("expected 'new value', got %s", cfg.Section1.StringField)
		}
		if !cfg.Section1.BoolField {
			t.Errorf("expected true")
		}
		if cfg.Section1.IntField != 42 {
			t.Errorf("expected 42")
		}
	})

	t.Run("ParseForm error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/Section1", errReader{})
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("expected 303 See Other, got %d", rr.Code)
		}
	})

	t.Run("Section not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/UnknownSection", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("expected 303 See Other, got %d", rr.Code)
		}
	})

	t.Run("Field error (int)", func(t *testing.T) {
		form := url.Values{}
		form.Add("IntField", "invalid")
		req := httptest.NewRequest(http.MethodPost, "/Section1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("expected 303 See Other, got %d", rr.Code)
		}
	})

	t.Run("Error custom unmarshal", func(t *testing.T) {
		form := url.Values{}
		form.Add("custom_field", "error")
		req := httptest.NewRequest(http.MethodPost, "/Section1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("expected 303 See Other, got %d", rr.Code)
		}
	})
}

func TestHandleFieldErrors(t *testing.T) {
	cfg := &TestConfig{}
	handler, _ := web.New(cfg)

	t.Run("Int error", func(t *testing.T) {
		form := url.Values{}
		form.Add("IntField", "abc")
		req := httptest.NewRequest(http.MethodPost, "/Section1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	})

	t.Run("Uint error", func(t *testing.T) {
		form := url.Values{}
		form.Add("uint_field", "abc")
		req := httptest.NewRequest(http.MethodPost, "/Section1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	})

	t.Run("Float error", func(t *testing.T) {
		form := url.Values{}
		form.Add("float_field", "abc")
		req := httptest.NewRequest(http.MethodPost, "/Section1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	})

	t.Run("Other ints/uints", func(t *testing.T) {
		form := url.Values{}
		form.Add("Int8Field", "12")
		form.Add("Int16Field", "12")
		form.Add("Int32Field", "12")
		form.Add("Int64Field", "12")
		form.Add("Uint8Field", "12")
		form.Add("Uint16Field", "12")
		form.Add("Uint32Field", "12")
		form.Add("Uint64Field", "12")
		form.Add("Float32Field", "12.5")
		req := httptest.NewRequest(http.MethodPost, "/Section2", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	})
}

type customError struct{}

func (customError) Error() string { return "custom" }

func TestParseError_Error(t *testing.T) {
	err := &web.ParseError{Message: "test msg", Field: "f1", Err: errors.New("inner")}
	if err.Error() != "test msg field f1: inner" {
		t.Errorf("unexpected Error() result: %s", err.Error())
	}
}

func TestParseError_Is(t *testing.T) {
	inner := customError{}
	err2 := &web.ParseError{Err: inner}
	if !err2.Is(inner) {
		t.Errorf("expected Is to return true")
	}
}

func TestParseError_As(t *testing.T) {
	inner := customError{}
	err2 := &web.ParseError{Err: inner}

	var target customError
	if !errors.As(err2, &target) {
		t.Errorf("expected As to return true")
	}

	var parseErrTarget *web.ParseError
	if !errors.As(err2, &parseErrTarget) {
		t.Errorf("expected As to return true for ParseError itself")
	}
}
