package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/razanfawwaz/bimbingan/internal/db"
	"github.com/razanfawwaz/bimbingan/internal/handlers"
)

func main() {
	db.Init()
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/add-graduate", handlers.AddDataHandler).Methods("POST")
	r.HandleFunc("/admin", handlers.AdminHandler).Methods("GET")
	r.HandleFunc("/graduates", handlers.GraduatesListHandler)

	fs := http.FileServer(http.Dir("templates"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Println("server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
