package views

import (
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func Funcs(baseAssetURL string, assetVersion string) template.FuncMap {
	return template.FuncMap{
		// {{url "/users"}} => absolute path encode-safe
		"url": func(p string) string {
			u := &url.URL{Path: p}
			return u.String()
		},
		// {{asset "/css/app.css"}} -> /assets/css/app.css?v=...
		"asset": func(p string) string {
			u := &url.URL{Path: filepath.ToSlash(filepath.Join(baseAssetURL, p))}
			qs := ""
			if assetVersion != "" {
				qs = "?v=" + url.QueryEscape(assetVersion)
			}
			return u.String() + qs
		},
		// {{now}} -> time.Now formatted
		"now": func() string {
			return time.Now().Format(time.RFC1123)
		},
		// mark safe small html bits if needed: {{safe "<b>hi</b>"}}
		"safe": func(s string) template.HTML { return template.HTML(s) },
		// simple printf passthrough
		"printf": fmt.Sprintf,
	}
}

// Quick helper to compute file mtime hash-ish for cache-busting.
func ComputeAssetVersion(paths ...string) string {
	var latest int64
	for _, p := range paths {
		if fi, err := os.Stat(p); err == nil {
			if fi.ModTime().Unix() > latest {
				latest = fi.ModTime().Unix()
			}
		}
	}
	return fmt.Sprintf("%d", latest)
}
