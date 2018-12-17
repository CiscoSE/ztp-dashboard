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
	deviceListTemplate   *template.Template
	deviceDetailTemplate *template.Template
	db                   dbController
}

func (n deviceController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/ng/devices", n.handleDevices)
	r.HandleFunc("/ng/devices/detail", n.handleDevicesDetail)
	r.HandleFunc("/api/devices", n.handleAPIDevices)
	r.HandleFunc("/api/devices/types", n.handleAPIDeviceTypes)
}

// checkDeviceTypes check if NX and XR device types are present in Database
// If not present, will create them
func (n deviceController) checkDeviceTypes() {

	var deviceTypes []model.DeviceType

	// Open database
	session, err := n.db.OpenSession()
	if err != nil {
		log.Fatal("Cannot open database:" + err.Error() + "\n")
	}
	defer session.Close()

	// Read database
	dbCollection := session.DB("ztpDashboard").C("deviceType")
	err = dbCollection.Find(nil).All(&deviceTypes)
	if err != nil {
		log.Fatal("Cannot read database table:" + err.Error() + "\n")
	}

	// Check if deviceTypes exist and have length greater than 0
	if deviceTypes == nil {
		n.createDeviceTypes()
	} else if len(deviceTypes) == 0 {
		n.createDeviceTypes()
	}
}

// createDeviceTypes insert iOS-XR and NX-OS into the database
func (n deviceController) createDeviceTypes() {
	// Open database
	session, err := n.db.OpenSession()
	if err != nil {
		log.Fatal("Cannot open database:" + err.Error() + "\n")
	}
	defer session.Close()
	dbCollection := session.DB("ztpDashboard").C("deviceType")

	// Create IOS-XR device type
	deviceTypeXr := model.DeviceType{Name: "iOS-XR"}
	deviceTypeNx := model.DeviceType{Name: "NX-OS"}

	// Insert new device types in Database
	err = dbCollection.Insert(&deviceTypeXr)
	if err != nil {
		log.Fatal("Couldn't insert in database:" + err.Error() + "\n")
	}
	err = dbCollection.Insert(&deviceTypeNx)
	if err != nil {
		log.Fatal("Couldn't insert in database:" + err.Error() + "\n")
	}
}

func (n deviceController) handleDevices(w http.ResponseWriter, r *http.Request) {
	n.deviceListTemplate.Execute(w, nil)
}

func (n deviceController) handleDevicesDetail(w http.ResponseWriter, r *http.Request) {
	n.deviceDetailTemplate.Execute(w, nil)
}

func (n deviceController) handleAPIDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		// Decode the request body into an Device model.
		dec := json.NewDecoder(r.Body)
		device := &model.Device{}
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
		count, err := dbCollection.Find(bson.M{"hostname": device.Hostname}).Count()
		if err != nil {
			log.Print("Cannot read device table:" + err.Error() + "\n")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Hostname " + device.Hostname + " already in use"))
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
		count, err = dbCollection.Find(bson.M{"fixedip": device.Fixedip}).Count()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Cannot read ncsztp table:" + err.Error() + "\n"))
			return
		}
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Fixed IP " + device.Fixedip + " already in use"))
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
	case http.MethodGet:
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
			log.Print("Cannot read database:" + err.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if devices == nil {
			devices = []model.Device{}
		}
		enc := json.NewEncoder(w)
		enc.Encode(devices)

		break
	case http.MethodPut:

		// Decode the request body into an Device model.
		dec := json.NewDecoder(r.Body)
		device := &model.Device{}
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

		// Update new device in Database (Image and day0 script)
		err = dbCollection.Update(bson.M{"hostname": device.Hostname}, bson.M{"$set": bson.M{"config": device.Config}})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't update in database:" + err.Error() + "\n"))
			return
		}

		err = dbCollection.Update(bson.M{"hostname": device.Hostname}, bson.M{"$set": bson.M{"image": device.Image}})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't update in database:" + err.Error() + "\n"))
			return
		}

		// Regenerate config file and restart dhcp service
		go dhcpController.GenerateConfigFiles()

		// Return ok message
		w.Write([]byte("ok"))
		break
	case http.MethodDelete:
		// Retrieve serial in request
		queryString, present := r.URL.Query()["serial"]

		if !present || len(queryString) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Serial parameter not found"))
		}
		deviceSerial := queryString[0]

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
		count, err := dbCollection.Find(bson.M{"serial": deviceSerial}).Count()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't retrieve object from DB"))
			return
		}
		if count != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't find single object to delete in DB"))
			return
		}
		err = dbCollection.Remove(bson.M{"serial": deviceSerial})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't remove object from DB"))
			return
		}

		// Regenerate dhcp and scripts
		dhcpController.GenerateConfigFiles()

		w.Write([]byte("Ok"))
		break
	}
}

// handleAPIDeviceTypes return a list of device types from the database
func (n deviceController) handleAPIDeviceTypes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var deviceTypes []model.DeviceType

		// Open database
		session, err := n.db.OpenSession()
		if err != nil {
			log.Print("Cannot open database:" + err.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer session.Close()
		dbCollection := session.DB("ztpDashboard").C("deviceType")

		// Read database
		err = dbCollection.Find(nil).All(&deviceTypes)
		if err != nil {
			log.Fatal(err)
		}
		// If result is nil, return an empty slice
		if deviceTypes == nil {
			deviceTypes = []model.DeviceType{}
		}
		enc := json.NewEncoder(w)
		enc.Encode(deviceTypes)

		break
	}
}
