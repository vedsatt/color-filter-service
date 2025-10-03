package main

import (
	"log"
	"net/http"

	"git.miem.hse.ru/kg25-26/aisavelev.git/application/handlers"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.Use(handlers.LoggingMiddleware)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	handlers.SetupRoutes(r)

	// Очистка старых сессий
	go handlers.CleanupSessions()

	log.Println("Server starting on :8080")
	log.Println("Static files served from: ./static")
	log.Fatal(http.ListenAndServe(":8080", r))
}
