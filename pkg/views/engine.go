package views

import (
	"errors"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Engine struct {
	mu           sync.RWMutex
	tpls         map[string]*template.Template
	fs           fs.FS  // where templates live
	globPattern  string // e.g. "resources/views/**/*.tmpl"
	funcs        template.FuncMap
	devReload    bool   // if true, reparses on each Render
	entryPrefix  string // "views/"
	layoutPrefix string // "layouts/"
}

type Options struct {
	FS           fs.FS
	GlobPattern  string
	Funcs        template.FuncMap
	DevReload    bool
	EntryPrefix  string // default "views/"
	LayoutPrefix string // default "layouts/"
}

func New(opts Options) (*Engine, error) {
	if opts.FS == nil {
		// default to OS
		opts.FS = os.DirFS(".")
	}
	if opts.GlobPattern == "" {
		opts.GlobPattern = "resources/views/**/*.tmpl"
	}
	if opts.EntryPrefix == "" {
		opts.EntryPrefix = "views/"
	}
	if opts.LayoutPrefix == "" {
		opts.LayoutPrefix = "layouts/"
	}
	e := &Engine{
		tpls:         map[string]*template.Template{},
		fs:           opts.FS,
		globPattern:  opts.GlobPattern,
		funcs:        opts.Funcs,
		devReload:    opts.DevReload,
		entryPrefix:  opts.EntryPrefix,
		layoutPrefix: opts.LayoutPrefix,
	}
	if !e.devReload {
		if err := e.loadAll(); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (e *Engine) loadAll() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	matches, err := fs.Glob(e.fs, e.globPattern)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return errors.New("no templates found for pattern: " + e.globPattern)
	}

	// Parse all files into a single root, then clone per-entry on Render
	root := template.New("root").Funcs(e.funcs)
	var bld *template.Template = root
	for _, path := range matches {
		data, err := fs.ReadFile(e.fs, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(path)
		var t *template.Template
		if bld == root {
			t = bld.New(name)
		} else {
			t = bld.New(name)
		}
		if _, err := t.Parse(string(data)); err != nil {
			return err
		}
	}

	// Build per-entry templates (entries start with e.entryPrefix)
	e.tpls = make(map[string]*template.Template)
	for _, path := range matches {
		if !strings.Contains(path, "/") {
			continue
		}
		if !strings.Contains(path, e.entryPrefix) {
			continue
		}
		// entry name like "views/home"
		entry := strings.TrimSuffix(filepath.ToSlash(path), ".tmpl")
		// clone root and set the default to execute the entry template
		cl, err := root.Clone()
		if err != nil {
			return err
		}
		// Set the name we will Execute
		e.tpls[entry] = cl.Lookup(entry)
	}
	return nil
}

func (e *Engine) Render(wr Writer, status int, name string, data any) error {
	if e.devReload {
		// reparse on each render
		if err := e.loadAll(); err != nil {
			return err
		}
	}

	e.mu.RLock()
	tpl := e.tpls[name]
	e.mu.RUnlock()

	if tpl == nil {
		// maybe user passed "home" instead of "views/home"
		if !strings.HasPrefix(name, e.entryPrefix) {
			name = e.entryPrefix + name
		}
		e.mu.RLock()
		tpl = e.tpls[name]
		e.mu.RUnlock()
		if tpl == nil {
			return errors.New("view not found: " + name)
		}
	}

	wr.WriteHeader(status)
	return tpl.Execute(wr, data)
}

// ----------------- small interface to decouple http -----------------
type Writer interface {
	WriteHeader(statusCode int)
	Write([]byte) (int, error)
}
