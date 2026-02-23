# webcfg

`webcfg` is a Go library that instantly generates a beautiful, responsive, and functional web-based configuration interface directly from your Go structs.

It eliminates the need to write HTML, CSS, or JavaScript boilerplate for simple configuration pages, making it perfect for admin panels, local tools, and prototype dashboards.

## Features

*   **Zero Boilerplate**: Just define a struct, and you get a UI.
*   **Modern UI**: Built with [Bulma 1.0](https://bulma.io/), featuring a clean and responsive design.
*   **Struct Tags**: Customize field labels, input types, icons, and helper text using `web` struct tags.
*   **Type Safe**: leverages Go's strong typing for form handling.
*   **Custom Themes**: Easily customize colors (Primary, Info, Danger, etc.) using CSS variables.
*   **Embedded Assets**: All necessary CSS and fonts are embedded, with support for custom assets (favicons, icons).
*   **Hooks**: `Initializable` and `UpdateReceiver` interfaces for custom logic on start and update.

## Installation

```bash
go get github.com/gwangyi/webcfg
```

## Quick Start

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gwangyi/webcfg/web"
)

// Define your configuration struct
type AppConfig struct {
	// Simple text field with customization via tags
	ServerName string `web:"server_name,Server Name,text,server,,"`
	
	// Number field with an icon
	Port int `web:"port,Port,number,hashtag,,,"`
	
	// Checkbox
	DebugMode bool `web:"debug,Enable Debug Mode,,bug,,,"`
	
	// Textarea for longer content
	Description string `web:"desc,Description,textarea,info,,,"`
}

func main() {
	// Initialize your config with default values
	cfg := &AppConfig{
		ServerName: "My Awesome App",
		Port:       8080,
		DebugMode:  true,
	}

	// Create the handler
	// web.New takes the config pointer and optional customizations
	handler, err := web.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Configuration server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
```

## Struct Tags Documentation

Customize how your struct fields are rendered using the `web` tag. The format is comma-separated values:

`web:"name,label,type,icon,status,help"`

| Position | Property | Description | Example |
| :--- | :--- | :--- | :--- |
| 1 | **Name** | The form field name (and ID). Defaults to struct field name. | `username` |
| 2 | **Label** | The human-readable label displayed above the input. | `User Name` |
| 3 | **Type** | The HTML input type. Supported: `text`, `number`, `password`, `email`, `checkbox`, `textarea`. | `password` |
| 4 | **Icon** | [FontAwesome](https://fontawesome.com/) icon name (without `fa-` prefix). | `user`, `lock`, `envelope` |
| 5 | **Status** | Bulma status color for the input (e.g., `primary`, `info`, `success`, `warning`, `danger`). | `danger` |
| 6 | **Help** | Help text displayed below the input field. | `Must be at least 8 chars` |

**Example:**
```go
Password string `web:"password,User Password,password,lock,danger,Required field"`
```

## Advanced Usage

### Customizing the Theme

You can override the default Bulma colors by passing a `web.Theme` struct to `web.New`. `webcfg` automatically converts hex codes to the necessary CSS variables.

```go
theme := web.Theme{
    Primary: "#8e44ad", // Wisteria Purple
    Link:    "#2980b9",
    Success: "#27ae60",
}

handler, _ := web.New(cfg, web.WithTheme(&theme))
```

### Handling Updates

To execute logic when a configuration section is updated (e.g., to reload a service or save to disk), implement the `UpdateReceiver[T]` interface on your struct fields.

```go
type DatabaseConfig struct {
    Host string `web:"host,Host,,,,"`
}

// Implement UpdateReceiver on the struct pointer
func (d *DatabaseConfig) Updated(parent *AppConfig, n web.Notifier) error {
    log.Printf("Database config updated! New host: %s", d.Host)
    
    // You can send notifications back to the UI
    n.Notify(web.Notification{
        Message: "Database connection restarted successfully",
        Status:  "success",
    })
    
    return nil
}
```

### Custom Assets

You can provide your own assets (like `favicon.ico` or `icon.png`) using `web.WithAssets`.

```go
// Using embed.FS or os.DirFS
handler, _ := web.New(cfg, web.WithAssets(os.DirFS("./static")))
```

## License

MIT License. See [LICENSE](LICENSE) file for details.
