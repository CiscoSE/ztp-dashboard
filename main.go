package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/CiscoSE/ztp-dashboard/controller"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	templates := populateTemplates()

	controller.Startup(templates, r)
	log.Println("Listening in http://0.0.0.0:8080/web/")
	http.ListenAndServe(":8080", r)
}

func populateTemplates() map[string]*template.Template {
	result := make(map[string]*template.Template)
	basePath := os.Getenv("GOPATH") + "/src/github.com/CiscoSE/ztp-dashboard/htmlTemplates"
	layout := template.Must(template.ParseFiles(basePath + "/_layout.html"))
	template.Must(
		layout.ParseFiles(basePath + "/_default_menu.html"))
	dir, err := os.Open(basePath + "/content")
	if err != nil {
		panic("Failed to open template blocks directory: " + err.Error())
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		panic("Failed to read contents of content directory: " + err.Error())
	}
	for _, fi := range fis {
		f, err := os.Open(basePath + "/content/" + fi.Name())
		if err != nil {
			panic("Failed to open template '" + fi.Name() + "'")
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			panic("Failed to read content from file '" + fi.Name() + "'")
		}
		f.Close()
		tmpl := template.Must(layout.Clone())
		_, err = tmpl.Parse(string(content))
		if err != nil {
			panic("Failed to parse contents of '" + fi.Name() + "' as template")
		}
		result[fi.Name()] = tmpl
	}
	return result
}
