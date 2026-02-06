package utils

import (
	//"goFrame/src/types"
	"html/template"
	"net/http"
	"path/filepath"
)

// AppVersion is set from main.go at startup
var AppVersion = "dev"

type PageData struct {
	Title      string
	Theme      string
	Version    string
	CustomData map[string]interface{}
}

// Define the base directories for views and templates
const (
	viewsDir     = "www/views/"
	templatesDir = "www/views/templates/"
)

// Define the common layout templates filenames
var templateFiles = []string{
	"layout.html",
	"header.html",
	"footer.html",
}

// Initialize the common templates with full paths
var layout = PrependDir(templatesDir, templateFiles)

var loginLayout = PrependDir(templatesDir, []string{"login-layout.html", "footer.html"})
var gameLayout = PrependDir(templatesDir, []string{"game-layout.html"})

func RenderTemplate(w http.ResponseWriter, data PageData, view string, useLoginLayout bool) {
	RenderTemplateWithLayout(w, data, view, useLoginLayout, false)
}

func RenderTemplateWithLayout(w http.ResponseWriter, data PageData, view string, useLoginLayout bool, useGameLayout bool) {
	if data.Version == "" {
		data.Version = AppVersion
	}
	viewTemplate := filepath.Join(viewsDir, view)
	componentPattern := filepath.Join(viewsDir, "components", "*.html")
	componentTemplates, err := filepath.Glob(componentPattern)
	if err != nil {
		http.Error(w, "Error loading component templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load game components
	gameComponentPattern := filepath.Join(viewsDir, "game", "*.html")
	gameComponentTemplates, err := filepath.Glob(gameComponentPattern)
	if err != nil {
		http.Error(w, "Error loading game component templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load game tab components
	gameTabPattern := filepath.Join(viewsDir, "game", "tabs", "*.html")
	gameTabTemplates, err := filepath.Glob(gameTabPattern)
	if err != nil {
		http.Error(w, "Error loading game tab templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load main page tab components (for homepage)
	tabPattern := filepath.Join(viewsDir, "tabs", "*.html")
	tabTemplates, err := filepath.Glob(tabPattern)
	if err != nil {
		http.Error(w, "Error loading tab templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var templates []string
	var layoutName string

	if useGameLayout {
		templates = append(gameLayout, viewTemplate)
		layoutName = "game-layout"
	} else if useLoginLayout {
		templates = append(loginLayout, viewTemplate)
		layoutName = "login-layout"
	} else {
		templates = append(layout, viewTemplate)
		layoutName = "layout"
	}
	templates = append(templates, componentTemplates...)
	templates = append(templates, gameComponentTemplates...)
	templates = append(templates, gameTabTemplates...)
	templates = append(templates, tabTemplates...)

	tmpl, err := template.New("").Funcs(template.FuncMap{}).ParseFiles(templates...)
	if err != nil {
		http.Error(w, "Error parsing templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, layoutName, data)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}

// Helper function to prepend a directory path to a list of filenames
func PrependDir(dir string, files []string) []string {
	var fullPaths []string
	for _, file := range files {
		fullPaths = append(fullPaths, dir+file)
	}
	return fullPaths
}