package controller

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

type settingsController struct {
	template *template.Template
}

func (n settingsController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/ng/settings", n.handleSettings)

}

func (n settingsController) handleSettings(w http.ResponseWriter, r *http.Request) {
	n.template.Execute(w, nil)
}
