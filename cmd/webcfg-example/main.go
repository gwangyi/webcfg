package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gwangyi/webcfg/web"
)

type DatabaseConfig struct {
	Host     string `web:"host,Host Name,text,server,,"`
	Port     int    `web:"port,Port Number,number,hashtag,,,," `
	User     string `web:"user,Username,text,user,," `
	Password string `web:"password,Password,password,key,," `
}

type FeatureConfig struct {
	EnableFeatureA bool `web:"enable_a,Enable Feature A,,check-square,,," `
	EnableFeatureB bool `web:"enable_b,Enable Feature B,,check-square,,," `
}

type AdvancedConfig struct {
	MaxRetries uint          `web:"retries,Maximum Retries,number,redo,,," `
	Threshold  float64       `web:"threshold,Success Threshold,number,chart-line,,," `
	Duration   DurationValue `web:"duration,Refresh Interval,text,clock,,," `
}

type DurationValue time.Duration

func (d *DurationValue) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = DurationValue(dur)
	return nil
}

func (d DurationValue) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

type AppConfig struct {
	Database    DatabaseConfig
	Features    FeatureConfig
	Advanced    AdvancedConfig
	Theme       web.Theme
	Description struct {
		About string `web:"about,About this app,textarea,info,,,"`
	}
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	// Create temporary directory for assets
	assetsDir, err := os.MkdirTemp("", "webcfg-assets")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(assetsDir)

	log.Printf("Downloading assets to %s...", assetsDir)

	// Download random favicon (32x32)
	if err := downloadFile("https://picsum.photos/32", filepath.Join(assetsDir, "favicon.ico")); err != nil {
		log.Printf("Failed to download favicon: %v", err)
	}

	// Download random icon (128x128)
	if err := downloadFile("https://picsum.photos/128", filepath.Join(assetsDir, "icon.png")); err != nil {
		log.Printf("Failed to download icon: %v", err)
	}

	cfg := &AppConfig{
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
			User: "admin",
		},
		Features: FeatureConfig{
			EnableFeatureA: true,
		},
		Advanced: AdvancedConfig{
			MaxRetries: 3,
			Threshold:  0.95,
			Duration:   DurationValue(5 * time.Minute),
		},
		Theme: web.Theme{
			Primary: "#8e44ad", // Wisteria purple
		},
	}
	cfg.Description.About = "This is a simple application to demonstrate webcfg functionality.\nIt supports various field types including text, number, checkbox, and now textarea!\nTry changing some values and clicking 'Submit'."

	handler, err := web.New(cfg, web.WithAssets(os.DirFS(assetsDir)), web.WithTheme(&cfg.Theme))
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	// Simple logging wrapper to see updates
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			log.Printf("Update request for %s", r.URL.Path)
			defer func() {
				log.Printf("Current Config: %+v", cfg)
			}()
		}
		handler.ServeHTTP(w, r)
	})

	fmt.Println("WebCfg Example Server")
	fmt.Println("=====================")
	fmt.Println("Access the configuration at: http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop.")

	if err := http.ListenAndServe(":8080", wrappedHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
