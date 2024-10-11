package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/harshk200/snippetbox/internal/models"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippetModel  *models.SnippetModel
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
}

func openDB(dns string) (*sql.DB, error) {
	// NOTE: this db here is not a connection it's a connection pool (i.e. no connections are made yet but when they do this will hold them)
	// also this is concurrent safe
	db, err := sql.Open("mysql", dns)
	if err != nil {
		return nil, err
	}

	// NOTE: just testing if the connection to db is good or not?
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	addr := flag.String("addr", ":3000", "HTTP network address")
	dns := flag.String("dns", "web:password@/snippetbox?parseTime=true&interpolateParams=true", "DNS or connection string for MySQl connection")

	flag.Parse()

	infoLog := log.New(os.Stdout, "\u001b[34mINFO\u001b[0m\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "\u001b[31mERROR\u001b[0m\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dns)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	app := &application{
		infoLog:       infoLog,
		errorLog:      errorLog,
		snippetModel:  &models.SnippetModel{DB: db}, // NOTE: creating the new snippetModel Instance here
		templateCache: templateCache,
		formDecoder:   formDecoder,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on port %s\n", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err) // NOTE: ListenAndServe always return a not-nil error so no need to put a if err != nil check here
}
