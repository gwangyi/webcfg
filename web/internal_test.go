package web

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var MockFSError = errors.New("mock fs error")

type MockFS struct{}

func (m *MockFS) Open(name string) (fs.File, error) {
	return nil, MockFSError
}

func TestWriteIndexParseError(t *testing.T) {
	// Backup original FS
	originalFS := indexTmplFS
	defer func() { indexTmplFS = originalFS }()

	// Set bad FS
	indexTmplFS = &MockFS{}

	p := &Page{}
	// We can call unexported writeIndex since we are in package web
	err := p.writeIndex(nil) // Writer doesn't matter as ParseFS fails first

	if err == nil {
		t.Errorf("Expected error from writeIndex with bad FS")
	}
	if !errors.Is(err, MockFSError) {
		if !strings.Contains(err.Error(), MockFSError.Error()) && !strings.Contains(err.Error(), "pattern matches no files") {
			t.Errorf("Expected error wrapping MockFSError or 'pattern matches no files', got %v", err)
		}
	}
}

func TestServeHTTP_WriteError(t *testing.T) {
	originalFS := indexTmplFS
	defer func() { indexTmplFS = originalFS }()
	indexTmplFS = &MockFS{}

	cfg := &struct{}{}
	handler, _ := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", rr.Code)
	}
}
