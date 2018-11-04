package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go/relay"
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
	modelCtx := getmodelContext()
	requestHandler := newRequestHandler(modelCtx)

	schema, err := getSchema("./schema.graphql", modelCtx.repoConnection)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	router.Handle("/api", &relay.Handler{Schema: schema})

	router.HandleFunc("/admin", requestHandler.AdminPage)
	router.HandleFunc("/{board}", requestHandler.BoardPage).Methods("GET")
	router.HandleFunc("/{board}", requestHandler.AddThread).Methods("POST")
	router.HandleFunc("/thread/{id:[0-9]+}", requestHandler.ThreadPage).Methods("GET")
	router.HandleFunc("/thread/{id:[0-9]+}", requestHandler.AddMessage).Methods("POST")
	router.HandleFunc("/author/{author}", requestHandler.AuthorPage)
	router.HandleFunc("/", requestHandler.MainPage)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	router.PathPrefix("/media/").Handler(http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))

	gracefulShutdown()

	log.Println("starting server...")
	err = http.ListenAndServe(":8000", router)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	runServer()
}
