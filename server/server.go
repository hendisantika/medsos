package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

// Medsos is media social API server
type Medsos struct {
	address string
	db      *mgo.Database
	root    string
	router  *mux.Router
}

// New creates new Medsos instance
func New(address string, db *mgo.Database, root string) (*Medsos, error) {
	medsos := Medsos{address: address, db: db, root: root}
	medsos.initRouter(root)
	return &medsos, nil
}

// initRouter initialize router and handler
func (m *Medsos) initRouter(root string) {
	m.router = mux.NewRouter()
	m.router.HandleFunc(root+"register", m.registerActor).Methods("POST")
	m.router.HandleFunc(root+"feeds", m.postActivity).Methods("POST")
	m.router.HandleFunc(root+"feeds/{actor}/", m.getFeeds)
	m.router.HandleFunc(root+"follow/{actor}/", m.follow).Methods("POST")
	m.router.HandleFunc(root+"follow/{actor}/{friend}", m.unfollow).Methods("DELETE")
}

// ListendAndServe listens for request
func (m *Medsos) ListenAndServe() error {
	return http.ListenAndServe(m.address, m.router)
}

// ActorName is used for encode/decode json/bson data
type ActorName struct {
	Name string `json:"actor" bson:"actor"`
}

// Actor is Medsos actor
type Actor struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name    string        `json:"actor" bson:"actor"`
	Friends []ActorName   `json:"friends" bson:"friends"`
}

// Activity is Medsos activity
type Activity struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Actor   string        `json:"actor" bson:"actor"`
	Verb    string        `json:"verb" bson:"verb"`
	Object  string        `json:"object" bson:"object"`
	Target  string        `json:"target" bson:"target"`
	Related bson.ObjectId `json:"related_id" bson:"related_id,omitempty"`
}

// jsonResponse writes json data as response
func jsonResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		encoder := json.NewEncoder(w)
		err := encoder.Encode(payload)
		if err != nil {
			log.Printf("error encoding json: %v", err.Error())
			jsonErrorResponse(w, http.StatusInternalServerError, "error encoding json")
		}
	}
}

// jsonErrorResponse writes error message as json
func jsonErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	jsonResponse(w, statusCode, map[string]string{"message": message})
}

// registerActor registers new Medsos actor
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

// postActivity posts new activity
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

// getFeeds get feeds for specific actor
func (m *Medsos) getFeeds(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actor, ok := vars["actor"]
	if !ok {
		jsonErrorResponse(w, http.StatusNotFound, "actor not found")
	} else {
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

// follow follows specific actor
func (m *Medsos) follow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["actor"]
	var friend ActorName
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&friend)
	if !ok {
		jsonErrorResponse(w, http.StatusNotFound, "actor not found")
	} else {
		actors := m.db.C("actor")
		var actor Actor
		err := actors.Find(bson.M{"actor": name}).One(&actor)
		if err != nil {
			log.Printf("Error querying actor: %v\n", err.Error())
			jsonErrorResponse(w, http.StatusInternalServerError, "error querying actor, try again later")
		} else {
			change := bson.M{"$push": bson.M{"friends": friend.Name}}
			err = actors.Update(bson.M{"_id": actor.Id}, change)
			if err != nil {
				log.Printf("Error updating actor: %v\n", err.Error())
				jsonErrorResponse(w, http.StatusInternalServerError, "error updating actor, try again later")
			} else {
				jsonResponse(w, http.StatusNoContent, nil)
			}
		}
	}
}

// unfollow delete specific friend from friend list
func (m *Medsos) unfollow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["actor"]
	if !ok {
		jsonErrorResponse(w, http.StatusNotFound, "actor not found")
		return
	}

	friend, ok := vars["friend"]
	if !ok {
		jsonErrorResponse(w, http.StatusNotFound, "friend not found")
		return
	}

	actors := m.db.C("actor")
	var actor Actor
	err := actors.Find(bson.M{"actor": name, "friends": friend}).One(&actor)
	if err != nil {
		log.Printf("Error querying actor: %v\n", err.Error())
		jsonErrorResponse(w, http.StatusInternalServerError, "error querying actor, try again later")
		return
	}

	change := bson.M{"$pull": bson.M{"friends": friend}}
	err = actors.Update(bson.M{"_id": actor.Id}, change)
	if err != nil {
		log.Printf("Error updating actor: %v\n", err.Error())
		jsonErrorResponse(w, http.StatusInternalServerError, "error updating actor, try again later")
	} else {
		jsonResponse(w, http.StatusNoContent, nil)
	}
}
