package controller

import (
	"html/template"
	"net/http"
	"github.com/gorilla/mux"
)

type index struct {
	template *template.Template
}

func (h index) registerRoutes(r *mux.Router) {

	r.NotFoundHandler = http.HandlerFunc(h.redirectHome)
	r.HandleFunc("/web/", h.handleIndex)
	r.HandleFunc("/web/index", h.handleIndex)
	r.HandleFunc("/web/home", h.handleIndex)
	r.HandleFunc("/web/ncsztp", h.handleIndex)
	r.HandleFunc("/web/smartphy", h.handleIndex)
	r.HandleFunc("/web/smartphy/cbr8", h.handleIndex)
	r.HandleFunc("/web/smartphy/rpd", h.handleIndex)
	r.HandleFunc("/web/smartphy/rpdassociation", h.handleIndex)
}

func (h index) handleIndex(w http.ResponseWriter, r *http.Request) {

	h.template.Execute(w, nil)
}

func (h index) redirectHome(w http.ResponseWriter, r *http.Request){
	http.Redirect(w, r, "/web/", http.StatusSeeOther)
}