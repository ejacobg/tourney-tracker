package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/ejacobg/tourney-tracker/http"
	"github.com/ejacobg/tourney-tracker/postgres"
	"html/template"
	"log"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	dsn := flag.String("dsn", "", "PostgreSQL DSN")
	challongeUsername := flag.String("challonge-user", "", "Challonge Username")
	challongePassword := flag.String("challonge-pass", "", "Challonge Password or API Key")
	startggKey := flag.String("startgg-key", "", "start.gg API Key")
	flag.Parse()

	tc, err := newTemplateCache()
	if err != nil {
		log.Fatalln("Failed to create template:", err)
	}

	db, err := openDB(*dsn)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}

	srv := http.NewServer(*challongeUsername, *challongePassword, *startggKey)
	srv.Addr = ":4000"
	srv.Templates = tc
	srv.EntrantService = postgres.EntrantService{db}
	srv.PlayerService = postgres.PlayerService{db}
	srv.TierService = postgres.TierService{db}
	srv.TournamentService = postgres.TournamentService{db}

	fmt.Println("Serving on http://localhost:4000")
	log.Fatalln(srv.ListenAndServe())
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

func newTemplateCache() (cache map[string]*template.Template, err error) {
	cache = make(map[string]*template.Template)

	pages, err := filepath.Glob("ui/html/pages/*/*.go.html")
	if err != nil {
		return nil, err
	}
	pages = append(pages, "ui/html/pages/index.go.html") // Manually adding the index page.

	for _, page := range pages {
		name := strings.TrimPrefix(filepath.ToSlash(page), "ui/html/pages/")
		fmt.Println(name)

		files := []string{
			"ui/html/base.go.html",
			"ui/html/partials/nav.go.html",
			page,
		}

		tmpl, err := template.New(name).ParseFiles(files...)
		if err != nil {
			return nil, err
		}

		cache[name] = tmpl
	}

	return cache, nil
}
