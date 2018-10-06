package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
)

// gracefulShutdown handle program interrupt by user
func gracefulShutdown() {
	var signalChan = make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGTERM)
	signal.Notify(signalChan, syscall.SIGINT)

	go func() {
		<-signalChan
		log.Println("stopping server")
		os.Exit(0)
	}()
}

func runServer() {
	router := mux.NewRouter()
	router.HandleFunc("/admin", AdminPage)
	router.HandleFunc("/{board}", BoardPage).Methods("GET")
	router.HandleFunc("/{board}", AddThread).Methods("POST")
	router.HandleFunc("/thread/{id:[0-9]+}", ThreadPage).Methods("GET")
	router.HandleFunc("/thread/{id:[0-9]+}", AddMessage).Methods("POST")
	router.HandleFunc("/author/{author}", AuthorPage)
	router.HandleFunc("/", MainPage)

	gracefulShutdown()

	log.Println("starting server...")
	err := http.ListenAndServe(":8000", router)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	runServer()
}
