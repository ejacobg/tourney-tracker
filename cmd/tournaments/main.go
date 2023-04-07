package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/ejacobg/tourney-tracker/tournament"
	"github.com/ejacobg/tourney-tracker/tournament/controller"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	dsn := flag.String("dsn", "", "PostgreSQL DSN")
	challongeUsername := flag.String("challonge-user", "", "Challonge Username")
	challongePassword := flag.String("challonge-pass", "", "Challonge Password or API Key")
	startggKey := flag.String("startgg-key", "", "start.gg API Key")
	flag.Parse()

	index, err := template.New("index").ParseFiles("ui/html/base.go.html", "ui/html/partials/nav.go.html", "ui/html/pages/tournaments/index.go.html")
	if err != nil {
		log.Fatalln("Failed to create template:", err)
	}

	db, err := openDB(*dsn)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}

	ctlr := controller.New(*challongeUsername, *challongePassword, *startggKey)
	ctlr.Model = tournament.Model{db}
	ctlr.Views.Index = index

	http.HandleFunc("/", ctlr.Index)
	http.HandleFunc("/tournaments/new", ctlr.New)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("ui/static"))))

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
