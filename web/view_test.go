package web_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gwangyi/webcfg/web"
)

type mockNotifier struct {
	notifications []web.Notification
}

func (m *mockNotifier) Notify(n web.Notification) {
	m.notifications = append(m.notifications, n)
}

type MyTextUnmarshaler struct {
	Value string
}

func (m *MyTextUnmarshaler) UnmarshalText(text []byte) error {
	if string(text) == "error" {
		return errors.New("unmarshal error")
	}
	m.Value = string(text)
	return nil
}

func (m MyTextUnmarshaler) MarshalText() ([]byte, error) {
	if m.Value == "marshal_error" {
		return nil, errors.New("marshal error")
	}
	return []byte(m.Value), nil
}

type TestConfig struct {
	Section1 struct {
		StringField  string            `web:"string_field,String Field,text,icon,status,help"`
		BoolField    bool              `web:",,,,"`
		IntField     int               `web:""`
		UintField    uint              `web:"uint_field"`
		FloatField   float64           `web:"float_field"`
		CustomField  MyTextUnmarshaler `web:"custom_field"`
		IgnoredField string            `web:"-"`
		unexported   string
	}
	Section2 struct {
		Int8Field    int8
		Int16Field   int16
		Int32Field   int32
		Int64Field   int64
		Uint8Field   uint8
		Uint16Field  uint16
		Uint32Field  uint32
		Uint64Field  uint64
		Float32Field float32
	}
	unexportedSection struct {
		Field string
	}
}

func TestBuildFieldMarshalError(t *testing.T) {
	cfg := &TestConfig{
		Section1: struct {
			StringField  string            `web:"string_field,String Field,text,icon,status,help"`
			BoolField    bool              `web:",,,,"`
			IntField     int               `web:""`
			UintField    uint              `web:"uint_field"`
			FloatField   float64           `web:"float_field"`
			CustomField  MyTextUnmarshaler `web:"custom_field"`
			IgnoredField string            `web:"-"`
			unexported   string
		}{
			CustomField: MyTextUnmarshaler{Value: "marshal_error"},
		},
	}
	handler, _ := web.New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestServeCustomCSS(t *testing.T) {
	cfg := &TestConfig{}

	handler, _ := web.New(cfg)
	req := httptest.NewRequest(http.MethodGet, "/assets/css/custom.css", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found when no theme is provided, got %d", rr.Code)
	}
}

func TestServeCustomCSSWithOption(t *testing.T) {
	cfg := &TestConfig{}

	theme := web.Theme{
		Primary: "#123456",
		Info:    "#111",
		Text:    "black",
	}
	handler, _ := web.New(cfg, web.WithTheme(&theme))
	req := httptest.NewRequest(http.MethodGet, "/assets/css/custom.css", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "text/css" {
		t.Errorf("expected Content-Type text/css")
	}

	body := rr.Body.String()
	if !strings.Contains(body, "--bulma-primary-h") {
		t.Errorf("expected custom primary color HSL in CSS")
	}
}
