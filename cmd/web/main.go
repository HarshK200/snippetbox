package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	// NOTE: yesterday the book finished till
	// NOTE: => parsing the runtime config
	addr := flag.String("addr", ":3000", "HTTP network address")
	flag.Parse()

	// NOTE: => establishing the dependencies for the handlers (that are on the app struct btw)
	infoLog := log.New(os.Stdout, "\u001b[34mINFO\u001b[0m\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "\u001b[31mERROR\u001b[0m\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		infoLog:  infoLog,
		errorLog: errorLog,
	}

	// NOTE: => running the HTittyPee server
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on port %s\n", *addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err) // NOTE: ListenAndServe always return a not-nil error so no need to put a if err != nil check here
}
