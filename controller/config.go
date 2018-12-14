package controller

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/CiscoSE/ztp-dashboard/model"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

// configController defines the templates and databases to be used
type configController struct {
	configListTemplate   *template.Template
	configDetailTemplate *template.Template
	db                   dbController
}

// registerRoutes specifies what are the URL that this controller will respond to
func (c configController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/ng/configs", c.handleConfigs)
	r.HandleFunc("/api/configs", c.handleAPIConfigs)
	r.HandleFunc("/ng/configs/detail", c.handleConfigDetail)
}

// handleConfig will be executed when a request to /ng/configs is done
func (c configController) handleConfigs(w http.ResponseWriter, r *http.Request) {
	c.configListTemplate.Execute(w, nil)
}

// handleConfigDetail will be executed when a request to /ng/configs/detail is done
func (c configController) handleConfigDetail(w http.ResponseWriter, r *http.Request) {
	c.configDetailTemplate.Execute(w, nil)
}

// handleAPIConfigs will be executed when a request to /api/configs is done
func (c configController) handleAPIConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// If method is POST, create a new object
	case "POST":
		// Decode the request body into an Config model.
		dec := json.NewDecoder(r.Body)
		var config = &model.Config{}
		err := dec.Decode(config)

		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Couldn't decode json:" + err.Error() + "\n"))
			return
		}

		// Open database
		session, err := c.db.OpenSession()
		if err != nil {
			log.Print("Cannot open database:" + err.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer session.Close()

		dbCollection := session.DB("ztpDashboard").C("config")

		// Check if the name has been used before
		count, err := dbCollection.Find(bson.M{"name": config.Name}).Count()
		if err != nil {
			log.Print("Cannot read device table:" + err.Error() + "\n")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Configuration name " + config.Name + " already in use. Please type another name"))
			return
		}

		// Create config file
		d1 := []byte(config.Configuration)
		err = ioutil.WriteFile(basePath+"/public/configs/"+config.Name+".conf", d1, 0644)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't save config in local disk:" + err.Error() + "\n"))
			return
		}

		config.Locationurl = "/configs/" + config.Name + ".conf"

		// Insert new configuration in Database
		err = dbCollection.Insert(&config)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't insert config in database:" + err.Error() + "\n"))
			return
		}

		// Return ok message
		w.Write([]byte("ok"))
		break
	// If method is GET, return all objects
	case "GET":

		var configs []model.Config

		// Open database
		session, err := c.db.OpenSession()
		if err != nil {
			log.Print("Cannot open database:" + err.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer session.Close()
		dbCollection := session.DB("ztpDashboard").C("config")

		err = dbCollection.Find(nil).All(&configs)
		if err != nil {
			log.Print("Cannot read database:" + err.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if configs == nil {
			configs = []model.Config{}
		}
		enc := json.NewEncoder(w)
		enc.Encode(configs)

		break
	}
}
