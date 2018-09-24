package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func runServer() {
	router := mux.NewRouter()
	router.HandleFunc("/admin", AdminPage)
	router.HandleFunc("/{board}", BoardPage).Methods("GET")
	router.HandleFunc("/{board}", AddThread).Methods("POST")
	router.HandleFunc("/thread/{id:[0-9]+}", ThreadPage).Methods("GET")
	router.HandleFunc("/thread/{id:[0-9]+}", AddMessage).Methods("POST")
	router.HandleFunc("/author/{author}", AuthorPage)
	router.HandleFunc("/", MainPage)

	log.Println("starting server...")
	err := http.ListenAndServe(":8000", router)
	if err != nil {
		panic(err)
	}
}

func main() {
	runServer()
}
