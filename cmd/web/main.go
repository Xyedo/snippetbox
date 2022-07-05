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

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
	"github.com/xyedo/snippetbox/pkg/models/mysql"
)

type ContextKey string

var ContextKeyUser = ContextKey("user")

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	session       *sessions.Session
	snippets      *mysql.SnippetModel
	users         *mysql.UserModel
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	// dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "mySQL databases")
	pass := flag.String("passDB", "web:pass@/snippetbox?parseTime=true", "MYSQL DB Password for user:web\n for parsing web:{pass}@/snippetbox?parseTime=true")
	secret := flag.String("secret", "K/2TiTnFoRJWpHM3sksO5i6zpmJ9ryczujFwVjiy5Tk=", "Secret Key For Session")
	flag.Parse()
	dsn := fmt.Sprintf("web:%s@/snippetbox?parseTime=true", *pass)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache("../../ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true
	app := &application{
		infoLog:  infoLog,
		errorLog: errorLog,
		session:  session,
		snippets: &mysql.SnippetModel{
			DB: db,
		},
		users: &mysql.UserModel{
			DB: db,
		},
		templateCache: templateCache,
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
	err = srv.ListenAndServeTLS("../../tls/cert.pem", "../../tls/key.pem")
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
