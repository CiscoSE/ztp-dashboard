package controller

import (
	"github.com/globalsign/mgo"
	"log"
	"os"
)

type dbController struct {

}

func (d dbController) OpenSession() (*mgo.Session, error) {
	// Open database
	session, err := mgo.Dial(os.Getenv("DB_URI"));
	if err != nil {
		log.Fatalf("Cannot open database:" + err.Error() + "\n")

	} else {
		session.SetMode(mgo.Monotonic, true)
	}

	return session, err
}
