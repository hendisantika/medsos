package main

import (
	"log"
	"os"

	"github.com/globalsign/mgo"
	"medsos/server"
)

func main() {
	address := os.Getenv("MEDSOS_ADDRESS")
	host := os.Getenv("MEDSOS_MONGO_HOST")
	db := os.Getenv("MEDSOS_MONGO_DB")
	session, err := mgo.Dial(host)
	mongo := session.DB(db)

	server, err := server.New(address, mongo, "/")
	if err != nil {
		log.Fatal("cannot initiate server")
		log.Fatal(err)
	} else {
		log.Fatal(server.ListenAndServe())
	}
}
