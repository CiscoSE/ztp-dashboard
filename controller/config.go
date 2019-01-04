package controller

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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
	r.HandleFunc("/configs/{configName}", c.handleConfigFiles)
}

// handleConfigFiles is responsable for serving config to devices and also to update the state of it
func (c configController) handleConfigFiles(w http.ResponseWriter, r *http.Request) {
	remoteIP := strings.Split(r.RemoteAddr, ":")[0]

	requestVars := mux.Vars(r)
	content, err := ioutil.ReadFile(basePath + "/public/configs/" + requestVars["configName"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		CustomLog("handleConfigFiles (reading image file): "+err.Error(), ErrorSeverity)
		return
	}

	var device model.Device
	// Open database
	session, err := c.db.OpenSession()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		go CustomLog("handleConfigFiles (open database): "+err.Error(), ErrorSeverity)
		return
	}
	defer session.Close()
	dbCollection := session.DB("ztpDashboard").C("device")

	// If device not found log the error and continue. Otherwhise update database
	err = dbCollection.Find(bson.M{"fixedip": remoteIP}).One(&device)
	if err != nil {
		go CustomLog("handleConfigFiles (Find request device): "+remoteIP+" "+err.Error(), DebugSeverity)
	} else {
		go CustomLog("handleConfigFiles: Updating device "+device.Hostname+" (serial "+device.Serial+") status to 'Running day 0 config'", DebugSeverity)
		device.Status = "Running day 0 config"
		dbCollection.Update(bson.M{"fixedip": remoteIP}, &device)
		go WebexTeamsCtl.SendMessage("Device " + device.Hostname + " (serial " + device.Serial + ") is running day 0 config " + requestVars["configName"])
	}

	w.Write(content)
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
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIConfigs (decode json): "+err.Error(), ErrorSeverity)
			return
		}

		// Open database
		session, err := c.db.OpenSession()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIConfigs (open database): "+err.Error(), ErrorSeverity)
			return
		}
		defer session.Close()

		dbCollection := session.DB("ztpDashboard").C("config")

		// Check if the name has been used before
		count, err := dbCollection.Find(bson.M{"name": config.Name}).Count()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIConfigs (read database): "+err.Error(), ErrorSeverity)
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
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIConfigs (save config file to local disk): "+err.Error(), ErrorSeverity)
			return
		}

		config.Locationurl = "/configs/" + config.Name + ".conf"

		// Insert new configuration in Database
		err = dbCollection.Insert(&config)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIConfigs (insert database): "+err.Error(), ErrorSeverity)
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
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIConfigs (open database): "+err.Error(), ErrorSeverity)
			return
		}
		defer session.Close()
		dbCollection := session.DB("ztpDashboard").C("config")

		err = dbCollection.Find(nil).All(&configs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIConfigs (read database): "+err.Error(), ErrorSeverity)
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
