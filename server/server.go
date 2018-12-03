package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

type Medsos struct {
	address string
	db      *mgo.Database
	root    string
	router  *mux.Router
}

func New(address string, db *mgo.Database, root string) (*Medsos, error) {
	medsos := Medsos{address: address, db: db, root: root}
	medsos.initRouter(root)
	return &medsos, nil
}

func (m *Medsos) initRouter(root string) {
	m.router = mux.NewRouter()
	m.router.HandleFunc(root+"feeds", m.postActivity).Methods("POST")
	m.router.HandleFunc(root+"feeds/{actor}/", m.getFeeds)
}

func (m *Medsos) ListenAndServe() error {
	return http.ListenAndServe(m.address, m.router)
}

type Actor struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name    string        `json:"actor" bson:"name"`
	Friends []Actor       `json:"friends" bson:"friends"`
}

type Activity struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Actor   string        `json:"actor" bson:"actor"`
	Verb    string        `json:"verb" bson:"verb"`
	Object  string        `json:"object" bson:"object"`
	Target  string        `json:"target" bson:"target"`
	Related bson.ObjectId `json:"related_id" bson:"related_id,omitempty"`
}

func jsonResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(payload)
	if err != nil {
		log.Printf("error encoding json: %v", err.Error())
		jsonErrorResponse(w, http.StatusInternalServerError, "error encoding json")
	}
}

func jsonErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	jsonResponse(w, statusCode, map[string]string{"message": message})
}

func (m *Medsos) registerActor(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var actor Actor
	err := decoder.Decode(&actor)
	if err != nil {
		log.Println(err)
		jsonErrorResponse(w, http.StatusBadRequest, "invalid data posted")
	} else {
		actors := m.db.C("actor")
		actor.Id = bson.NewObjectId()
		err := actors.Insert(&actor)
		if err != nil {
			log.Printf("error inserting actor: %v", err.Error())
			jsonErrorResponse(w, http.StatusInternalServerError, "error inserting actor, try again later.")
		} else {
			jsonResponse(w, http.StatusCreated, actor)
		}
	}
}

func (m *Medsos) postActivity(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var activity Activity
	err := decoder.Decode(&activity)
	if err != nil {
		log.Println(err)
		jsonErrorResponse(w, http.StatusBadRequest, "invalid data posted")
		return
	}
	activity.Id = bson.NewObjectId()

	feeds := m.db.C("feeds")
	err = feeds.Insert(&activity)
	if err != nil {
		log.Printf("Error inserting activity feed: %v\n", err.Error())
		jsonErrorResponse(w, http.StatusInternalServerError, "error insert activity, try again later.")
	} else {
		jsonResponse(w, http.StatusCreated, activity)
	}
}

func (m *Medsos) getFeeds(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actor, ok := vars["actor"]
	if !ok {
		jsonErrorResponse(w, http.StatusNotFound, "actor not found")
	} else {
		w.WriteHeader(http.StatusOK)
		feeds := m.db.C("feeds")
		var activity []Activity
		err := feeds.Find(bson.M{"actor": actor}).All(&activity)
		if err != nil {
			log.Printf("Error querying feed: %v\n", err.Error())
			jsonErrorResponse(w, http.StatusInternalServerError, "error querying feeds, try again later.")
		} else {
			jsonResponse(w, http.StatusOK, activity)
		}
	}
}
