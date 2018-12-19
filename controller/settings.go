package controller

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/CiscoSE/ztp-dashboard/model"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

type settingsController struct {
	template        *template.Template
	situationMgrCtl SituationMgrController
	webexTeamsCtl   WebexTeamsController
	db              dbController
}

func (n settingsController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/ng/settings", n.handleSettings)
	r.HandleFunc("/api/settings", n.handleAPISettings)

}

func (n settingsController) handleSettings(w http.ResponseWriter, r *http.Request) {
	n.template.Execute(w, nil)
}

func (n settingsController) handleAPISettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		// Decode the request body into an Device model.
		dec := json.NewDecoder(r.Body)
		settings := &model.Settings{}
		err := dec.Decode(settings)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPISettings (decode jason): "+err.Error(), ErrorSeverity)
			return
		}

		// Open database
		session, err := n.db.OpenSession()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPISettings (open database): "+err.Error(), ErrorSeverity)
			return
		}
		defer session.Close()

		dbCollection := session.DB("ztpDashboard").C("settings")

		// Delete previous settings
		_, err = dbCollection.RemoveAll(bson.M{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPISettings (remove previous settings): "+err.Error(), ErrorSeverity)
			return
		}

		// Insert new settings in Database
		err = dbCollection.Insert(&settings)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPISettings (insert database): "+err.Error(), ErrorSeverity)
			return
		}

		// Send notification
		go WebexTeamsCtl.SendMessage("#Settings changed \\n Situation manager URL: " + settings.SituationMgrURL + " \\n\\n New webex team room: " + settings.WebexTeamsRoomID)

		// Return ok message
		w.Write([]byte("ok"))
		break
	case http.MethodGet:
		var settings model.Settings

		// Open database
		session, err := n.db.OpenSession()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			go CustomLog("handleAPISettings (open database): "+err.Error(), ErrorSeverity)
			return
		}
		defer session.Close()
		dbCollection := session.DB("ztpDashboard").C("settings")

		count, err := dbCollection.Find(nil).Count()
		if count == 0 {
			settings = model.Settings{
				SituationMgrURL:  "",
				WebexTeamsRoomID: "",
			}
		} else {
			err = dbCollection.Find(nil).One(&settings)
			if err != nil {
				go CustomLog("handleAPISettings (read database): "+err.Error(), ErrorSeverity)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}
		enc := json.NewEncoder(w)
		enc.Encode(settings)

		break
	}
}
