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
type imageController struct {
	imageListTemplate   *template.Template
	imageDetailTemplate *template.Template
	db                  dbController
}

// registerRoutes specifies what are the URL that this controller will respond to
func (i imageController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/ng/images", i.handleImages)
	r.HandleFunc("/ng/images/detail", i.handleImagesDetail)
	r.HandleFunc("/api/images", i.handleAPIImages)
}

// handleConfig will be executed when a request to /ng/images is done
func (i imageController) handleImages(w http.ResponseWriter, r *http.Request) {
	i.imageListTemplate.Execute(w, nil)
}

// handleImagesDetail will be executed when a request to /ng/images/detail is done
func (i imageController) handleImagesDetail(w http.ResponseWriter, r *http.Request) {
	i.imageDetailTemplate.Execute(w, nil)
}

// handleAPIImages will be executed when a request to /api/images is done
func (i imageController) handleAPIImages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// If method is POST, create a new object
	case http.MethodPost:
		deviceTypeName := r.FormValue("deviceType")
		imageName := r.FormValue("name")

		if deviceTypeName == "" || imageName == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Device Type and Image name are required"))
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			log.Print(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIImages (retrieve file from request): "+err.Error(), ErrorSeverity)
			return
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)

		err = ioutil.WriteFile(basePath+"/public/images/"+imageName, data, 0666)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIImages (save image file): "+err.Error(), ErrorSeverity)
			return
		}

		// Open database
		session, err := i.db.OpenSession()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIImages (open database): "+err.Error(), ErrorSeverity)
			return
		}
		defer session.Close()

		// Retrieve and check that a valid device type has been selected

		dbCollection := session.DB("ztpDashboard").C("deviceType")

		// Read database
		deviceTypes := dbCollection.Find(bson.M{"name": deviceTypeName})
		deviceTypesCount, err := deviceTypes.Count()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIImages (read database): "+err.Error(), ErrorSeverity)
			return
		}
		if deviceTypesCount == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid device type selected"))
			return
		}

		dbCollection = session.DB("ztpDashboard").C("image")

		// Check if the name has been used before
		count, err := dbCollection.Find(bson.M{"name": imageName}).Count()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIImages (read database): "+err.Error(), ErrorSeverity)
			return
		}
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Image name already in use"))
			return
		}

		deviceType := model.DeviceType{}
		deviceTypes.One(&deviceType)

		image := &model.Image{
			Name:        imageName,
			DeviceType:  deviceType,
			Locationurl: "/images/" + imageName,
		}

		// Insert new configuration in Database
		err = dbCollection.Insert(&image)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIImages (insert database): "+err.Error(), ErrorSeverity)
			return
		}

		// Return ok message
		w.Write([]byte("ok"))
		break
	// If method is GET, return all objects
	case http.MethodGet:

		var configs []model.Config

		// Open database
		session, err := i.db.OpenSession()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIImages (open database): "+err.Error(), ErrorSeverity)
			return
		}
		defer session.Close()
		dbCollection := session.DB("ztpDashboard").C("image")

		err = dbCollection.Find(nil).All(&configs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPIImages (read database): "+err.Error(), ErrorSeverity)
		}
		if configs == nil {
			configs = []model.Config{}
		}
		enc := json.NewEncoder(w)
		enc.Encode(configs)

		break
	}
}
