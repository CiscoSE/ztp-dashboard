package controller

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/CiscoSE/ztp-dashboard/model"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

type deviceController struct {
	template *template.Template
	db       dbController
}

func (n deviceController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/ng/devices", n.handleDevices)
	r.HandleFunc("/api/devices", n.handleAPIDevices)
}

func (n deviceController) handleDevices(w http.ResponseWriter, r *http.Request) {
	n.template.Execute(w, nil)
}

func (n deviceController) handleAPIDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Decode the request body into an Device model.
		dec := json.NewDecoder(r.Body)
		var device *model.Device = &model.Device{}
		err := dec.Decode(device)

		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Couldn't decode json:" + err.Error() + "\n"))
			return
		}

		// Open database
		session, err := n.db.OpenSession()
		if err != nil {
			log.Print("Cannot open database:" + err.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer session.Close()

		dbCollection := session.DB("ztpDashboard").C("device")

		// Check if the name has been used before
		count, err := dbCollection.Find(bson.M{"host": device.Host}).Count()
		if err != nil {
			log.Print("Cannot read device table:" + err.Error() + "\n")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Host " + device.Host + " already in use"))
			return
		}

		// Check if the serial has been used before
		count, err = dbCollection.Find(bson.M{"serial": device.Serial}).Count()
		if err != nil {
			log.Print("Cannot read ncsztp table:" + err.Error() + "\n")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Serial " + device.Serial + " already in use"))
			return
		}

		// Check if the fixed IP has been used before
		count, err = dbCollection.Find(bson.M{"fixedip": device.FixedIp}).Count()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Cannot read ncsztp table:" + err.Error() + "\n"))
			return
		}
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Fixed IP " + device.FixedIp + " already in use"))
			return
		}

		// Insert new device in Database
		err = dbCollection.Insert(&device)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't insert in database:" + err.Error() + "\n"))
			return
		}

		// Regenerate config file and restart dhcp service
		go dhcpController.GenerateConfigFiles()

		// Return ok message
		w.Write([]byte("ok"))
		break
	case "GET":
		var devices []model.Device

		// Open database
		session, err := n.db.OpenSession()
		if err != nil {
			log.Print("Cannot open database:" + err.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer session.Close()
		dbCollection := session.DB("ztpDashboard").C("device")

		err = dbCollection.Find(nil).All(&devices)
		if err != nil {
			panic(err)
		}
		enc := json.NewEncoder(w)
		enc.Encode(devices)

		break
	}
}
