package web

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"reflect"
	"strings"

	"github.com/crazy3lf/colorconv"
)

type Field struct {
	Name     string
	Label    string
	Value    string
	Type     string
	Icon     string
	Status   string
	Help     string
	Readonly bool
}

type Section struct {
	Title    string
	Subtitle string
	Action   string
	Fields   []Field
}

type Notification struct {
	Message string
	Status  string
}

type Page struct {
	Title         string
	Subtitle      string
	Notifications []Notification
	Sections      []Section
	HasAssets     bool
}

type Theme struct {
	Primary string
	Link    string
	Info    string
	Success string
	Warning string
	Danger  string
	Dark    string
	Text    string
}

type Option func(*configPageOptions)

type configPageOptions struct {
	assets fs.FS
	theme  *Theme
}

func WithAssets(assets fs.FS) Option {
	return func(o *configPageOptions) {
		o.assets = assets
	}
}

func WithTheme(theme *Theme) Option {
	return func(o *configPageOptions) {
		o.theme = theme
	}
}

type configPage[T any] struct {
	Page
	config        *T
	assetsHandler http.Handler
	theme         *Theme
}

type Notifier interface {
	Notify(Notification)
}

func (p *Page) Notify(n Notification) {
	p.Notifications = append(p.Notifications, n)
}

type Initializable interface {
	Initialize(parent any, n Notifier) error
}

type UpdateReceiver interface {
	Updated(parent any, n Notifier) error
}

func writeBulmaColorVar(w io.Writer, name, val string) error {
	if val == "" {
		return nil
	}

	// Try parsing as hex
	if !strings.HasPrefix(val, "#") {
		return nil
	}

	c, err := colorconv.HexToColor(val)
	if err != nil {
		return err
	}
	h, s, l := colorconv.ColorToHSL(c)
	// Bulma 1.0 vars use units: deg for Hue, % for Saturation and Lightness.
	fmt.Fprintf(w, "\t--bulma-%s-h: %.0fdeg;\n", name, h)
	fmt.Fprintf(w, "\t--bulma-%s-s: %.0f%%;\n", name, s*100)
	fmt.Fprintf(w, "\t--bulma-%s-l: %.0f%%;\n", name, l*100)
	return nil
}

func (p *configPage[T]) serveCustomCSS(w http.ResponseWriter) {
	if p.theme == nil {
		http.NotFound(w, nil)
		return
	}

	w.Header().Set("Content-Type", "text/css")
	w.Write([]byte(":root {\n"))

	v := reflect.ValueOf(*p.theme)
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sf := typ.Field(i)
		fv := v.Field(i)

		if fv.Kind() == reflect.String {
			val := fv.String()
			name := strings.ToLower(sf.Name)
			writeBulmaColorVar(w, name, val)
		}
	}

	w.Write([]byte("}\n"))
}

func parseTag(v reflect.Value, sf reflect.StructField) Field {
	f := Field{
		Name:  sf.Name,
		Label: sf.Name,
		Type:  "text",
	}
	if v.Kind() == reflect.Bool {
		f.Type = "checkbox"
	}

	tag := sf.Tag.Get("web")
	if tag != "" {
		parts := strings.Split(tag, ",")
		if len(parts) > 0 && parts[0] != "" {
			f.Name = parts[0]
			f.Label = parts[0]
		}
		if len(parts) > 1 && parts[1] != "" {
			f.Label = parts[1]
		}
		if len(parts) > 2 && parts[2] != "" {
			f.Type = parts[2]
		}
		if len(parts) > 3 && parts[3] != "" {
			f.Icon = parts[3]
		}
		if len(parts) > 4 && parts[4] != "" {
			f.Status = parts[4]
		}
		if len(parts) > 5 && parts[5] != "" {
			f.Help = parts[5]
		}
	}

	return f
}

func (p *configPage[T]) initialize() error {
	v := reflect.ValueOf(p.config).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)

		if !fieldVal.CanInterface() {
			continue
		}

		if fieldVal.CanAddr() {
			if init, ok := fieldVal.Addr().Interface().(Initializable); ok {
				if err := init.Initialize(p.config, p); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *configPage[T]) serveAssets(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/assets/css/custom.css" {
		p.serveCustomCSS(w)
		return
	}

	// If custom assets are provided, check for favicon.ico and icon.png
	if p.assetsHandler != nil {
		if path == "/assets/favicon.ico" || path == "/assets/icon.png" {
			p.assetsHandler.ServeHTTP(w, r)
			return
		}
	}

	embeddedAssetsHandler.ServeHTTP(w, r)
}

func (p *configPage[T]) servePost(w http.ResponseWriter, r *http.Request) {
	sectionName := strings.TrimPrefix(r.URL.Path, "/")
	if err := p.updateConfig(sectionName, r); err != nil {
		p.Notify(Notification{Message: "Update failed: " + err.Error(), Status: "danger"})
	} else {
		p.Notify(Notification{Message: "Section updated successfully", Status: "success"})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (p *configPage[T]) serveIndex(w http.ResponseWriter) {
	p.buildPage()
	if err := p.writeIndex(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Clear notifications after rendering
	p.Notifications = nil
}

func (p *configPage[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/assets/"):
		p.serveAssets(w, r)
	case r.Method == http.MethodPost:
		p.servePost(w, r)
	case r.URL.Path == "/" || r.URL.Path == "/index.html":
		p.serveIndex(w)
	default:
		http.NotFound(w, r)
	}
}

func New[T any](config *T, opts ...Option) (http.Handler, error) {
	options := &configPageOptions{}
	for _, o := range opts {
		o(options)
	}

	var assetsHandler http.Handler
	if options.assets != nil {
		assetsHandler = http.StripPrefix("/assets/", http.FileServer(http.FS(options.assets)))
	}
	cfg := &configPage[T]{config: config, assetsHandler: assetsHandler, theme: options.theme}
	err := cfg.initialize()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
