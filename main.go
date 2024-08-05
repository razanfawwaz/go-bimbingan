package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/razanfawwaz/bimbingan/internal/db"
	"github.com/razanfawwaz/bimbingan/internal/handlers"
	"github.com/razanfawwaz/bimbingan/internal/middleware"
	"github.com/razanfawwaz/bimbingan/util"
)

func main() {
	db.Init()

	jwtSecret := util.GetConfig("JWT_SECRET")
	jwtKey := []byte(jwtSecret)

	r := mux.NewRouter()
	r.HandleFunc("/login", handlers.LoginHandler(jwtKey)).Methods("POST")
	r.HandleFunc("/login", handlers.LoginPageHandler).Methods("GET")

	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AuthMiddleware(jwtKey))
	adminRouter.HandleFunc("/dashboard", handlers.AdminHandler).Methods("GET")
	adminRouter.HandleFunc("/graduates", handlers.GetStudentsData).Methods("GET")
	adminRouter.HandleFunc("/update-status", handlers.UpdateStatusHandler).Methods("PATCH")
	adminRouter.HandleFunc("/token", handlers.GetAllToken).Methods("GET")
	adminRouter.HandleFunc("/token", handlers.CreateToken).Methods("POST")

	r.HandleFunc("/graduates", handlers.GraduatesListHandler).Methods("GET")
	r.HandleFunc("/add-data", handlers.AddDataPageHandler).Methods("GET")
	r.HandleFunc("/add-data", handlers.AddDataHandler).Methods("POST")
	r.HandleFunc("/advisor", handlers.AdvisorHandler).Methods("GET")

	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")

	fs := http.FileServer(http.Dir("templates"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Println("server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
