package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xyedo/snippetbox/internal/models"
)

type application struct {
	debug          bool
	errorLog       *log.Logger
	infoLog        *log.Logger
	sessionManager *scs.SessionManager
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface

	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	// dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "mySQL databases")
	pass := flag.String("passDB", "web:pass@/snippetbox?parseTime=true", "MYSQL DB Password for user:web\n for parsing web:{pass}@/snippetbox?parseTime=true")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()
	dsn := fmt.Sprintf("web:%s@/snippetbox?parseTime=true", *pass)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}
	formDecoder := form.NewDecoder()
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	app := &application{
		debug:          *debug,
		infoLog:        infoLog,
		errorLog:       errorLog,
		sessionManager: sessionManager,
		snippets: &models.SnippetModel{
			DB: db,
		},
		users: &models.UserModel{
			DB: db,
		},
		templateCache: templateCache,
		formDecoder:   formDecoder,
	}
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1%s", *addr),
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	infoLog.Printf("starting server on %s\n", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	if err != nil {
		errorLog.Fatal(err)
	}
}
func openDB(dsn string) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
