package render

import (
	"context"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/lmika/gopkgs/fp/slices"
)

var (
	templateSuffix = []string{".html", ".gohtml"}
)

type Config struct {
	templateFS fs.FS

	cacheMutex     *sync.RWMutex
	templateSet    *template.Template
	funcMaps       template.FuncMap
	frameTemplates []string
}

func New(tmplFS fs.FS, opts ...ConfigOption) *Config {
	cfg := &Config{
		templateFS:     tmplFS,
		cacheMutex:     new(sync.RWMutex),
		frameTemplates: nil,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	cfg.templateSet = cfg.buildTemplates()
	return cfg
}

func (tc *Config) rebuildTemplates() {
	newTemplates := tc.buildTemplates()

	tc.cacheMutex.Lock()
	defer tc.cacheMutex.Unlock()
	tc.templateSet = newTemplates
}

func (tc *Config) buildTemplates() *template.Template {
	mainTmpl := template.New("/")

	if tc.funcMaps != nil {
		mainTmpl = mainTmpl.Funcs(tc.funcMaps)
	}

	_ = fs.WalkDir(tc.templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if !slices.Contains(templateSuffix, filepath.Ext(path)) {
			return nil
		}

		tmpl, err := tc.parseTemplate(path)
		if err != nil {
			log.Printf("template %v: %v", path, err)
			return nil
		}

		if _, err := mainTmpl.AddParseTree(path, tmpl.Tree); err != nil {
			log.Printf("template %v: %v", path, err)
		}

		return nil
	})
	return mainTmpl
}

func (tc *Config) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := tc.NewInv()
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), renderContextKey, rc)))
	})
}

func (tc *Config) NewInv() *Inv {
	return &Inv{
		config: tc,
		values: make(map[string]interface{}),
	}
}

func (tc *Config) template(name string) (*template.Template, error) {
	return tc.templateSet.Lookup(name), nil
}

func (tc *Config) parseTemplate(name string) (*template.Template, error) {
	f, err := tc.templateFS.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tmplBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	tmpl := template.New(name)
	if tc.funcMaps != nil {
		tmpl = tmpl.Funcs(tc.funcMaps)
	}

	tmpl, err = tmpl.Parse(string(tmplBytes))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

type renderContextKeyType struct{}

var renderContextKey = renderContextKeyType{}
