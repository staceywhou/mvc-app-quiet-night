package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"log"

	"github.com/go-chi/chi"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./visitors.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS visitors (domain TEXT PRIMARY KEY, count INTEGER)`)
	if err != nil {
		log.Fatal(err)
	}
}

func createRoutes() chi.Router {
	// We're using chi as the router. You'll want to read
	// the documentation https://github.com/go-chi/chi
	// so that you can capture parameters like /events/5
	// or /api/events/4 -- where you want to get the
	// event id (5 and 4, respectively).

	r := chi.NewRouter()

	// set up both get and post
	r.Get("/about", aboutController)
	r.Get("/events/{id:[0-9]+}", detailsController)
	r.Post("/events/{id:[0-9]+}", detailsController)
	r.Get("/", indexController)
	r.Get("/visitor", visitorController)
	r.Get("/events/new", createController)
	r.Post("/events/new", createController)
	r.Get("/api/events", apiController)
	r.Get("/api/events/{id:[0-9]+}", apiEventController)

	addStaticFileServer(r, "/static/", "staticfiles")
	return r
}
