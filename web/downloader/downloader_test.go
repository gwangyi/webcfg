package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloaderMain(t *testing.T) {
	// Mock os.Exit
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}
	defer func() { osExit = os.Exit }()
	defer func() { osArgs = os.Args }()

	// Start a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/404" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock data"))
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	outFile := filepath.Join(tempDir, "out.txt")

	t.Run("Successful run", func(t *testing.T) {
		osArgs = []string{"downloader", ts.URL, outFile}
		exitCode = 0
		main()
		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}
	})

	t.Run("Invalid args", func(t *testing.T) {
		osArgs = []string{"downloader"}
		exitCode = 0
		main()
		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}
	})

	t.Run("HTTP Error (Not Found)", func(t *testing.T) {
		osArgs = []string{"downloader", ts.URL + "/404", outFile}
		exitCode = 0
		main()
		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}
	})

	t.Run("Network Error (Invalid URL)", func(t *testing.T) {
		osArgs = []string{"downloader", "http://invalid-url-that-does-not-exist:1234", outFile}
		exitCode = 0
		main()
		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}
	})

	t.Run("File creation error", func(t *testing.T) {
		badFile := filepath.Join(tempDir, "does-not-exist", "out.txt")
		osArgs = []string{"downloader", ts.URL, badFile}
		exitCode = 0
		main()
		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}
	})
}

// Ensure the io.Copy error is covered.
// To do that, we can mock the server to return a response but write a huge file to a small disk, or mock os.Create/io.Copy.
// Actually, I can use an httptest server that sends chunked data and aborts.
func TestDownloaderCopyError(t *testing.T) {
	osExit = func(code int) {}
	defer func() { osExit = os.Exit }()
	defer func() { osArgs = os.Args }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock")) // Content length is 100 but only wrote 4 bytes, then close
		// To force io.Copy error, we can use a Hijacker to force close connection
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	outFile := filepath.Join(tempDir, "out.txt")
	osArgs = []string{"downloader", ts.URL, outFile}

	err := run(osArgs)
	if err == nil {
		t.Errorf("expected error during io.Copy, got nil")
	}
}
