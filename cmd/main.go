package main

import (
	"log"
	"net/http"
	"os"

	"git.miem.hse.ru/kg25-26/aisavelev.git/application/handlers"

	"github.com/gorilla/mux"
)

func main() {
	createDir()

	r := mux.NewRouter()

	r.Use(handlers.LoggingMiddleware)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	handlers.SetupRoutes(r)

	go handlers.CleanupSessions()

	log.Println("Server starting on :8080")
	log.Println("Static files served from: ./static")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createDir() {
	err := os.Mkdir("static/uploads", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("Failed to create uploads directory: %v", err)
		return
	}
}
