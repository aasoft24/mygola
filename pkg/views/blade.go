// pkg/views/blade.go
package views

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

func RenderBlade(w http.ResponseWriter, view string, data map[string]interface{}) {
	content, _ := ioutil.ReadFile("resources/views/" + view)
	tplStr := string(content)

	// Extend layout
	if strings.Contains(tplStr, "@extends") {
		parts := strings.Split(tplStr, "\n")
		layout := strings.Trim(parts[0][len("@extends('"):len(parts[0])-2], "'")
		layoutContent, _ := ioutil.ReadFile("resources/views/" + layout + ".html")
		tplStr = string(layoutContent)
		// Replace @yield('content') with actual content
		start := strings.Index(string(content), "@section('content')")
		end := strings.Index(string(content), "@endsection")
		contentBlock := string(content)[start+18 : end]
		tplStr = strings.ReplaceAll(tplStr, "@yield('content')", contentBlock)

		// Replace @yield('title') if present
		titleStart := strings.Index(string(content), "@section('title'")
		if titleStart != -1 {
			titleEnd := strings.Index(string(content), ")")
			title := string(content)[titleStart+17 : titleEnd]
			tplStr = strings.ReplaceAll(tplStr, "@yield('title')", title)
		}
	}

	tmpl, _ := template.New("blade").Parse(tplStr)
	buf := new(bytes.Buffer)
	tmpl.Execute(buf, data)
	w.Write(buf.Bytes())
}
