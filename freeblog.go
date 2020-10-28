package main

import (
	"database/sql"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	//	"github.com/gorilla/feeds"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type PrintFunc func(format string, a ...interface{}) (n int, err error)

type User struct {
	Userid    int64
	Username  string
	HashedPwd string
}

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	sw, parms := parseArgs(args)

	// [-i new_file]  Create and initialize db file
	if sw["i"] != "" {
		dbfile := sw["i"]
		if fileExists(dbfile) {
			return fmt.Errorf("File '%s' already exists. Can't initialize it.\n", dbfile)
		}
		createTables(dbfile)
		return nil
	}

	// Need to specify a db file as first parameter.
	if len(parms) == 0 {
		s := `Usage:

   Start webservice using database file:
	freeblog <db file> [port]

   Initialize new database file:
	freeblog -i <new db file>

`
		fmt.Printf(s)
		return nil
	}

	// Exit if db file doesn't exist.
	dbfile := parms[0]
	if !fileExists(dbfile) {
		return fmt.Errorf(`Database file '%s' doesn't exist. Create one using:
	freeblog -i <instance.db>
   `, dbfile)
	}

	_, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return fmt.Errorf("Error opening '%s' (%s)\n", dbfile, err)
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/radio.ico") })
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./"))))

	port := "8000"
	if len(parms) > 1 {
		port = parms[1]
	}
	fmt.Printf("Listening on %s...\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	return err
}

func createTables(newfile string) {
	if fileExists(newfile) {
		s := fmt.Sprintf("File '%s' already exists. Can't initialize it.\n", newfile)
		fmt.Printf(s)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", newfile)
	if err != nil {
		fmt.Printf("Error opening '%s' (%s)\n", newfile, err)
		os.Exit(1)
	}

	ss := []string{
		"CREATE TABLE user (user_id INTEGER PRIMARY KEY NOT NULL, username TEXT UNIQUE, password TEXT);",
		"INSERT INTO user (user_id, username, password) VALUES (1, 'admin', '');",
		"CREATE TABLE savedgrid (user_id INTEGER PRIMARY KEY NOT NULL, gridjson TEXT);",
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("DB error (%s)\n", err)
		os.Exit(1)
	}
	for _, s := range ss {
		_, err := txexec(tx, s)
		if err != nil {
			tx.Rollback()
			log.Printf("DB error (%s)\n", err)
			os.Exit(1)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("DB error (%s)\n", err)
		os.Exit(1)
	}
}

//*** DB functions ***
func sqlstmt(db *sql.DB, s string) *sql.Stmt {
	stmt, err := db.Prepare(s)
	if err != nil {
		log.Fatalf("db.Prepare() sql: '%s'\nerror: '%s'", s, err)
	}
	return stmt
}
func sqlexec(db *sql.DB, s string, pp ...interface{}) (sql.Result, error) {
	stmt := sqlstmt(db, s)
	defer stmt.Close()
	return stmt.Exec(pp...)
}
func txstmt(tx *sql.Tx, s string) *sql.Stmt {
	stmt, err := tx.Prepare(s)
	if err != nil {
		log.Fatalf("tx.Prepare() sql: '%s'\nerror: '%s'", s, err)
	}
	return stmt
}
func txexec(tx *sql.Tx, s string, pp ...interface{}) (sql.Result, error) {
	stmt := txstmt(tx, s)
	defer stmt.Close()
	return stmt.Exec(pp...)
}

//*** Helper functions ***

// Helper function to make fmt.Fprintf(w, ...) calls shorter.
// Ex.
// Replace:
//   fmt.Fprintf(w, "<p>Some text %s.</p>", str)
//   fmt.Fprintf(w, "<p>Some other text %s.</p>", str)
// with the shorter version:
//   P := makeFprintf(w)
//   P("<p>Some text %s.</p>", str)
//   P("<p>Some other text %s.</p>", str)
func makeFprintf(w io.Writer) func(format string, a ...interface{}) (n int, err error) {
	return func(format string, a ...interface{}) (n int, err error) {
		return fmt.Fprintf(w, format, a...)
	}
}
func listContains(ss []string, v string) bool {
	for _, s := range ss {
		if v == s {
			return true
		}
	}
	return false
}
func fileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
func makePrintFunc(w io.Writer) func(format string, a ...interface{}) (n int, err error) {
	// Return closure enclosing io.Writer.
	return func(format string, a ...interface{}) (n int, err error) {
		return fmt.Fprintf(w, format, a...)
	}
}
func atoi(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
func idtoi(sid string) int64 {
	return int64(atoi(sid))
}
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
func atof(s string) float64 {
	if s == "" {
		return 0.0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return f
}

func unescapeUrl(qurl string) string {
	returl := "/"
	if qurl != "" {
		returl, _ = url.QueryUnescape(qurl)
	}
	return returl
}
func escape(s string) string {
	return html.EscapeString(s)
}

func parseArgs(args []string) (map[string]string, []string) {
	switches := map[string]string{}
	parms := []string{}

	standaloneSwitches := []string{}
	definitionSwitches := []string{"i"}
	fNoMoreSwitches := false
	curKey := ""

	for _, arg := range args {
		if fNoMoreSwitches {
			// any arg after "--" is a standalone parameter
			parms = append(parms, arg)
		} else if arg == "--" {
			// "--" means no more switches to come
			fNoMoreSwitches = true
		} else if strings.HasPrefix(arg, "--") {
			switches[arg[2:]] = "y"
			curKey = ""
		} else if strings.HasPrefix(arg, "-") {
			if listContains(definitionSwitches, arg[1:]) {
				// -a "val"
				curKey = arg[1:]
				continue
			}
			for _, ch := range arg[1:] {
				// -a, -b, -ab
				sch := string(ch)
				if listContains(standaloneSwitches, sch) {
					switches[sch] = "y"
				}
			}
		} else if curKey != "" {
			switches[curKey] = arg
			curKey = ""
		} else {
			// standalone parameter
			parms = append(parms, arg)
		}
	}

	return switches, parms
}

func handleErr(w http.ResponseWriter, err error, sfunc string) {
	log.Printf("%s: server error (%s)\n", sfunc, err)
	http.Error(w, fmt.Sprintf("%s", err), 500)
}
func handleDbErr(w http.ResponseWriter, err error, sfunc string) bool {
	if err == sql.ErrNoRows {
		http.Error(w, "Not found.", 404)
		return true
	}
	if err != nil {
		log.Printf("%s: database error (%s)\n", sfunc, err)
		http.Error(w, "Server database error.", 500)
		return true
	}
	return false
}
func handleTxErr(tx *sql.Tx, err error) bool {
	if err != nil {
		tx.Rollback()
		return true
	}
	return false
}

func genHash(sinput string) string {
	bsHash, err := bcrypt.GenerateFromPassword([]byte(sinput), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(bsHash)
}
func validateHash(shash, sinput string) bool {
	if shash == "" && sinput == "" {
		return true
	}
	err := bcrypt.CompareHashAndPassword([]byte(shash), []byte(sinput))
	if err != nil {
		return false
	}
	return true
}

func findUser(db *sql.DB, username string) *User {
	s := "SELECT user_id, username, password FROM user WHERE username = ?"
	row := db.QueryRow(s, username)
	var u User
	err := row.Scan(&u.Userid, &u.Username, &u.HashedPwd)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &u
}
func isUsernameExists(db *sql.DB, username string) bool {
	if findUser(db, username) == nil {
		return false
	}
	return true
}
