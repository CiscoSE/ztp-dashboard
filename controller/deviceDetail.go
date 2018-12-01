package controller

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

type deviceDetailController struct {
	template *template.Template
}

func (n deviceDetailController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/ng/devices/detail", n.handleDeviceDetail)
}

func (n deviceDetailController) handleDeviceDetail(w http.ResponseWriter, r *http.Request) {
	n.template.Execute(w, nil)
}
