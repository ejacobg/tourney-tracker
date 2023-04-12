package main

import (
	"database/sql"
	"flag"
	"fmt"
	controller "github.com/ejacobg/tourney-tracker/http"
	"github.com/ejacobg/tourney-tracker/postgres"
	"github.com/julienschmidt/httprouter"
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

	view, err := template.New("view").ParseFiles("ui/html/base.go.html", "ui/html/partials/nav.go.html", "ui/html/pages/tournaments/view.go.html")
	if err != nil {
		log.Fatalln("Failed to create template:", err)
	}

	edit, err := template.New("edit").ParseFiles("ui/html/pages/tournaments/edit.go.html")
	if err != nil {
		log.Fatalln("Failed to create template:", err)
	}

	db, err := openDB(*dsn)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}

	ctlr := controller.New(*challongeUsername, *challongePassword, *startggKey)
	ctlr.Model = postgres.Model{db}
	ctlr.Views.Index = index
	ctlr.Views.View = view
	ctlr.Views.Edit = edit

	router := httprouter.New()
	router.HandlerFunc("GET", "/", ctlr.Index)
	router.HandlerFunc("POST", "/tournaments/new", ctlr.New)
	router.HandlerFunc("GET", "/tournaments/:id", ctlr.View)
	router.HandlerFunc("GET", "/tournaments/:id/tier", ctlr.ViewTier)
	router.HandlerFunc("GET", "/tournaments/:id/tier/edit", ctlr.EditTier)
	router.Handler("GET", "/static/*filepath", http.FileServer(http.Dir("ui")))

	fmt.Println("Serving on http://localhost:4000")
	log.Fatalln(http.ListenAndServe(":4000", router))
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
