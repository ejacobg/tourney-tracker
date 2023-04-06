package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/ejacobg/tourney-tracker/tournament"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	dsn := flag.String("dsn", "", "PostgreSQL DSN")
	flag.Parse()

	index, err := template.New("index").ParseFiles("ui/html/base.go.html", "ui/html/partials/nav.go.html", "ui/html/pages/tournaments/index.go.html")
	if err != nil {
		log.Fatalln("Failed to create template:", err)
	}

	db, err := openDB(*dsn)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}

	controller := tournament.Controller{
		Model: tournament.Model{db},
		Views: struct {
			Index, View, Edit *template.Template
		}{
			index, nil, nil,
		},
	}

	http.HandleFunc("/", controller.Index)

	fmt.Println("Serving on http://localhost:4000")
	log.Fatalln(http.ListenAndServe(":4000", nil))
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
