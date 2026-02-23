package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	osArgs = os.Args
	osExit = os.Exit
)

func run(args []string) error {
	// Download URL specified in first argument as a file specified in second argument
	if len(args) != 3 {
		return fmt.Errorf("Usage: downloader <URL> <filename>")
	}

	url := args[1]
	filename := args[2]

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Error downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error downloading %s: status code %d", url, resp.StatusCode)
	}

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error creating file %s: %w", filename, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("Error saving to file %s: %w", filename, err)
	}

	fmt.Printf("Downloaded %s to %s\n", url, filename)
	return nil
}

func main() {
	if err := run(osArgs); err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}
}
