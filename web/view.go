package web

import (
	"embed"
	"encoding"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"reflect"
)

//go:embed templates/index.html.tmpl
var defaultIndexTmplFS embed.FS

var indexTmplFS fs.FS = defaultIndexTmplFS

//go:embed assets
var assetsFS embed.FS

var embeddedAssetsHandler = http.FileServer(http.FS(assetsFS))

func (p *Page) writeIndex(w io.Writer) error {
	tmpl, err := template.ParseFS(indexTmplFS, "templates/index.html.tmpl")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, p)
}

func (p *configPage[T]) buildPage() {
	v := reflect.ValueOf(p.config).Elem()
	t := v.Type()

	p.Title = t.Name()
	p.Sections = nil
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Only process exported fields that are structs (Sections)
		if field.PkgPath != "" || fieldVal.Kind() != reflect.Struct {
			continue
		}

		p.Sections = append(p.Sections, buildSection(fieldVal, field))
	}
	p.HasAssets = p.assetsHandler != nil
}

func buildSection(v reflect.Value, f reflect.StructField) Section {
	section := Section{
		Title:  f.Name,
		Action: f.Name,
	}

	st := v.Type()
	for i := 0; i < v.NumField(); i++ {
		subField := st.Field(i)
		subFieldVal := v.Field(i)

		if subField.PkgPath != "" {
			continue
		}

		section.Fields = append(section.Fields, buildField(subFieldVal, subField))
	}

	return section
}

func buildField(v reflect.Value, sf reflect.StructField) Field {
	f := parseTag(v, sf)

	// Get Value
	if tm, ok := v.Interface().(encoding.TextMarshaler); ok {
		if b, err := tm.MarshalText(); err == nil {
			f.Value = string(b)
		}
	} else {
		f.Value = fmt.Sprint(v.Interface())
	}

	return f
}
