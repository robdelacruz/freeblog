package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	//	"github.com/gorilla/feeds"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shurcooL/github_flavored_markdown"
	"golang.org/x/crypto/bcrypt"
)

type PrintFunc func(format string, a ...interface{}) (n int, err error)

type User struct {
	Userid    int64
	Username  string
	HashedPwd string
}

type Entry struct {
	Entryid  int64
	Title    string
	Body     string
	Createdt string
	Userid   int64
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

	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return fmt.Errorf("Error opening '%s' (%s)\n", dbfile, err)
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/radio.ico") })
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", indexHandler(db))
	http.HandleFunc("/login/", loginHandler(db))
	http.HandleFunc("/logout/", logoutHandler(db))
	http.HandleFunc("/signup/", signupHandler(db))
	http.HandleFunc("/password/", passwordHandler(db))
	http.HandleFunc("/profile/", profileHandler(db))
	http.HandleFunc("/addentry/", addentryHandler(db))
	http.HandleFunc("/entry/", entryHandler(db))

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
		"CREATE TABLE entry (entry_id INTEGER PRIMARY KEY NOT NULL, title TEXT, body TEXT, createdt TEXT NOT NULL, user_id INTEGER NOT NULL);",
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

func qunescape(s string) string {
	us, err := url.QueryUnescape(s)
	if err != nil {
		us = s
	}
	return us
}
func qescape(s string) string {
	return url.QueryEscape(s)
}
func escape(s string) string {
	return html.EscapeString(s)
}
func unescape(s string) string {
	return html.UnescapeString(s)
}

func isodate(t time.Time) string {
	return t.Format(time.RFC3339)
}
func parseisodate(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
func formatisodate(s string) string {
	t := parseisodate(s)
	return t.Format("2 Jan 2006")
}
func formatdate(s string) string {
	t := parseisodate(s)
	return t.Format("2 Jan 2006")
}

func parseMarkdown(s string) string {
	s = strings.ReplaceAll(s, "%", "%%")
	return string(github_flavored_markdown.Markdown([]byte(s)))
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
func logErr(sfunc string, err error) {
	log.Printf("%s error (%s)\n", sfunc, err)
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

func findUserById(db *sql.DB, userid int64) *User {
	s := "SELECT user_id, username, password FROM user WHERE user_id = ?"
	row := db.QueryRow(s, userid)
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

func genTok(u *User) string {
	tok := genHash(fmt.Sprintf("%s_%s", u.Username, u.HashedPwd))
	return tok
}
func validateTok(tok string, u *User) bool {
	return validateHash(tok, fmt.Sprintf("%s_%s", u.Username, u.HashedPwd))
}

func setLoginCookie(w http.ResponseWriter, username, tok string) {
	c := http.Cookie{
		Name:  "usernametok",
		Value: fmt.Sprintf("%s|%s", username, tok),
		Path:  "/",
		// Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, &c)
}
func delLoginCookie(w http.ResponseWriter) {
	c := http.Cookie{
		Name:   "usernametok",
		Value:  "",
		Path:   "/",
		MaxAge: 0,
	}
	http.SetCookie(w, &c)
}
func readLoginCookie(r *http.Request) (string, string) {
	c, err := r.Cookie("usernametok")
	if err != nil {
		return "", ""
	}

	var username, tok string
	ss := strings.Split(c.Value, "|")
	username = ss[0]
	if len(ss) > 1 {
		tok = ss[1]
	}
	return username, tok
}

// Reads and validates login cookie. If invalid username/token, return no user.
func validateLoginCookie(db *sql.DB, r *http.Request) (*User, string) {
	username, tok := readLoginCookie(r)
	if username == "" {
		return nil, ""
	}
	u := findUser(db, username)
	if u == nil {
		return nil, ""
	}
	if !validateTok(tok, u) {
		log.Printf("Token not validated for '%s' ", u.Username)
		return nil, ""
	}
	return u, tok
}

var ErrLoginIncorrect = errors.New("Incorrect username or password")

func login(db *sql.DB, username, pwd string) (string, error) {
	var u User
	s := "SELECT user_id, username, password FROM user WHERE username = ?"
	row := db.QueryRow(s, username)
	err := row.Scan(&u.Userid, &u.Username, &u.HashedPwd)
	if err == sql.ErrNoRows {
		return "", ErrLoginIncorrect
	}
	if err != nil {
		return "", err
	}
	if !validateHash(u.HashedPwd, pwd) {
		return "", ErrLoginIncorrect
	}

	// Return session token, this will be used to authenticate username
	// on every request by calling validateTok()
	tok := genTok(&u)
	return tok, nil
}

func signup(db *sql.DB, username, pwd string) error {
	if isUsernameExists(db, username) {
		return fmt.Errorf("username '%s' already exists", username)
	}

	hashedPwd := genHash(pwd)
	s := "INSERT INTO user (username, password) VALUES (?, ?);"
	_, err := sqlexec(db, s, username, hashedPwd)
	if err != nil {
		return fmt.Errorf("DB error creating user: %s", err)
	}
	return nil
}

func edituser(db *sql.DB, username, pwd string, newpwd string) error {
	// Validate existing password
	_, err := login(db, username, pwd)
	if err != nil {
		return err
	}

	// Set new password
	hashedPwd := genHash(newpwd)
	s := "UPDATE user SET password = ? WHERE username = ?"
	_, err = sqlexec(db, s, hashedPwd, username)
	if err != nil {
		return fmt.Errorf("DB error updating user password: %s", err)
	}
	return nil
}

func deluser(db *sql.DB, username, pwd string) error {
	// Validate existing password
	_, err := login(db, username, pwd)
	if err != nil {
		return err
	}

	// Delete user
	s := "DELETE FROM user WHERE username = ?"
	_, err = sqlexec(db, s, username)
	if err != nil {
		return fmt.Errorf("DB error deleting user: %s", err)
	}
	return nil
}

func findEntry(db *sql.DB, entryid int64) *Entry {
	s := "SELECT entry_id, title, body, createdt, user_id FROM entry WHERE entry_id = ?"
	row := db.QueryRow(s, entryid)
	var e Entry
	err := row.Scan(&e.Entryid, &e.Title, &e.Body, &e.Createdt, &e.Userid)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &e
}
func findEntries(db *sql.DB) ([]*Entry, error) {
	s := "SELECT entry_id, title, body, createdt, user_id FROM entry ORDER BY entry_id DESC"
	rows, err := db.Query(s)
	if err != nil {
		return nil, err
	}
	var ee []*Entry
	for rows.Next() {
		var e Entry
		rows.Scan(&e.Entryid, &e.Title, &e.Body, &e.Createdt, &e.Userid)
		ee = append(ee, &e)
	}
	return ee, nil
}

//*** HTML template functions ***
func printHtmlOpen(P PrintFunc, title string, jsurls []string) {
	P("<!DOCTYPE html>\n")
	P("<html>\n")
	P("<head>\n")
	P("<meta charset=\"utf-8\">\n")
	P("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	P("<title>%s</title>\n", title)
	P("<link rel=\"stylesheet\" type=\"text/css\" href=\"/static/style.css\">\n")
	for _, jsurl := range jsurls {
		P("<script defer src=\"%s\"></script>\n", jsurl)
	}
	P("<style>\n")
	P(".myfont {font-family: Helvetica Neue,Helvetica,Arial,sans-serif;}\n")
	P("</style>\n")
	P("</head>\n")
	P("<body class=\"py-2 text-base leading-6 myfont light\">\n")
	P("<div id=\"container\" class=\"mx-auto max-w-screen-sm\">\n")
}
func printHtmlClose(P PrintFunc) {
	P("</div>\n")
	P("</body>\n")
	P("</html>\n")
}
func printHeading(P PrintFunc, u *User) {
	P("<div class=\"flex flex-row justify-between border-b border-gray-500 pb-1 mb-4 text-sm\"\n>")
	P("    <div>\n")
	P("        <h1 class=\"inline self-end ml-1 mr-2 font-bold\"><a href=\"/\">FreeBlog</a></h1>\n")
	P("        <a href=\"about.html\" class=\"self-end mr-2\">About</a>\n")
	if u != nil {
		P("        <a href=\"/addentry\" class=\"pill self-center rounded px-2 py-1 mr-1\">Add Entry</a>\n")
	}
	P("    </div>\n")
	P("    <div>\n")
	if u != nil {
		P("        <div class=\"relative inline mr-2\">\n")
		P("            <a class=\"mr-1\" href=\"/profile\">%s</a>\n", escape(u.Username))
		P("            <div class=\"hidden popmenu absolute top-auto right-0 py-1 w-20 border border-gray-500 shadow-xs w-40\">\n")
		P("                <a href=\"#a\" class=\"block leading-none px-2 py-1 border-b\" role=\"menuitem\">Change Password</a>\n")
		P("                <a href=\"#a\" class=\"block leading-none px-2 py-1 border-b\" role=\"menuitem\">Delete Account</a>\n")
		P("                <a href=\"#a\" class=\"block leading-none px-2 py-1 border-b\" role=\"menuitem\">Reset LocalStorage</a>\n")
		P("            </div>\n")
		P("        </div>\n")
		P("        <a href=\"/logout\" class=\"inline self-end mr-1\">Logout</a>\n")
	} else {
		P("        <a href=\"/login\" class=\"inline self-end mr-1\">Login</a>\n")
	}
	P("    </div>\n")
	P("</div>\n")
}
func printFormOpen(P PrintFunc, action, heading string) {
	P("<form action=\"%s\" method=\"post\" class=\"panel mx-auto py-4 px-8 text-sm\">\n", action)
	if heading != "" {
		P("    <h1 class=\"font-bold mx-auto mb-2 text-center text-xl\">%s</h1>\n", heading)
	}
}
func printFormSmallOpen(P PrintFunc, action, heading string) {
	P("<form action=\"%s\" method=\"post\" class=\"panel mx-auto py-4 px-8 text-sm max-w-sm\">\n", action)
	if heading != "" {
		P("    <h1 class=\"font-bold mx-auto mb-2 text-center text-xl\">%s</h1>\n", heading)
	}
}
func printFormClose(P PrintFunc) {
	P("</form>\n")
}
func printFormInput(P PrintFunc, id, label, val string) {
	P("<div class=\"mb-2\">\n")
	P("    <label class=\"block font-bold uppercase text-xs\" for=\"%s\">%s</label>\n", id, label)
	P("    <input class=\"block border border-gray-500 py-1 px-4 w-full\" id=\"%[1]s\" name=\"%[1]s\" type=\"text\" value=\"%s\">\n", id, val)
	P("</div>\n")
}
func printFormInputPassword(P PrintFunc, id, label, val string) {
	P("<div class=\"mb-2\">\n")
	P("    <label class=\"block font-bold uppercase text-xs\" for=\"%s\">%s</label>\n", id, label)
	P("    <input class=\"block border border-gray-500 py-1 px-4 w-full\" id=\"%[1]s\" name=\"%[1]s\" type=\"password\" value=\"%s\">\n", id, val)
	P("</div>\n")
}
func printFormTextarea(P PrintFunc, id, label, val, rows string) {
	if rows == "" {
		rows = "22"
	}
	P("<div class=\"mb-2\">\n")
	P("    <label class=\"block font-bold uppercase text-xs\" for=\"%s\">%s</label>\n", id, label)
	P("    <textarea class=\"block border border-gray-500 py-1 px-4 w-full leading-5\" id=\"%[1]s\" name=\"%[1]s\" rows=\"%s\">%s</textarea>\n", id, rows, val)
	P("</div>\n")
}
func printFormError(P PrintFunc, errmsg string) {
	if errmsg == "" {
		return
	}
	P("<div class=\"mb-2\">\n")
	P("    <p class=\"font-bold uppercase text-xs\">%s</p>\n", errmsg)
	P("</div>\n")
}
func printFormSubmit(P PrintFunc, caption string) {
	P("<div class=\"mb-2\">\n")
	P("    <button type=\"submit\" class=\"inline w-full mx-auto py-1 px-2 border border-gray-500 font-bold mr-2\">%s</button>\n", caption)
	P("</div>\n")
}

// Ex. printFormLinks(P, "justify-end", "/signup", "Sign Up", "/login", "Login")
func printFormLinks(P PrintFunc, justify string, ss ...string) {
	type Link struct {
		Href    string
		Caption string
	}

	if justify == "" {
		justify = "justify-between"
	}

	var ll []Link
	var l Link
	for _, s := range ss {
		if l.Href == "" {
			l.Href = s
			continue
		}
		if l.Caption == "" {
			l.Caption = s
			ll = append(ll, l)

			l.Href = ""
			l.Caption = ""
		}

	}

	P("<div class=\"flex flex-row %s\">\n", justify)
	for _, l := range ll {
		P("    <a class=\"text-xs\" href=\"%s\">%s</a>\n", l.Href, l.Caption)
	}
	P("</div>\n")
}
func printDivOpen(P PrintFunc, heading string) {
	P("<div class=\"panel mx-auto py-4 px-8 text-sm\">\n")
	if heading != "" {
		P("    <h1 class=\"font-bold mx-auto mb-2 text-center text-xl\">%s</h1>\n", heading)
	}
}
func printDivSmallOpen(P PrintFunc, heading string) {
	P("<div class=\"panel mx-auto py-4 px-8 text-sm max-w-sm\">\n")
	if heading != "" {
		P("    <h1 class=\"font-bold mx-auto mb-2 text-center text-xl\">%s</h1>\n", heading)
	}
}
func printDivFlex(P PrintFunc, justify string) {
	P("<div class=\"flex flex-row %s\">\n", justify)
}
func printDivClose(P PrintFunc) {
	P("</div>\n")
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		var errmsg string
		var f struct{ username, pwd string }

		if r.Method == "POST" {
			f.username = r.FormValue("username")
			f.pwd = r.FormValue("pwd")
			for {
				tok, err := login(db, f.username, f.pwd)
				if err != nil {
					errmsg = fmt.Sprintf("%s", err)
					break
				}
				setLoginCookie(w, f.username, tok)

				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printHeading(P, u)

		printFormSmallOpen(P, "/login/", "Log In")
		printFormInput(P, "username", "username", f.username)
		printFormInputPassword(P, "pwd", "password", f.pwd)
		printFormError(P, errmsg)
		printFormSubmit(P, "Login")
		printFormLinks(P, "", "/signup", "Create New Account", "/", "Cancel")
		printFormClose(P)

		printHtmlClose(P)
	}
}
func logoutHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		delLoginCookie(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
func signupHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		var errmsg string
		var f struct{ username, pwd, pwd2 string }

		if r.Method == "POST" {
			f.username = r.FormValue("username")
			f.pwd = r.FormValue("pwd")
			f.pwd2 = r.FormValue("pwd2")
			for {
				if f.pwd != f.pwd2 {
					errmsg = "passwords don't match"
					break
				}
				err := signup(db, f.username, f.pwd)
				if err != nil {
					errmsg = fmt.Sprintf("%s", err)
					break
				}
				tok, err := login(db, f.username, f.pwd)
				if err != nil {
					errmsg = fmt.Sprintf("%s", err)
					break
				}
				setLoginCookie(w, f.username, tok)

				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printHeading(P, u)

		printFormSmallOpen(P, "/signup/", "Sign Up")
		printFormInput(P, "username", "username", f.username)
		printFormInputPassword(P, "pwd", "password", f.pwd)
		printFormInputPassword(P, "pwd2", "re-enter password", f.pwd2)
		printFormError(P, errmsg)
		printFormSubmit(P, "Sign Up")
		printFormLinks(P, "justify-end", "/", "Cancel")
		printFormClose(P)

		printHtmlClose(P)
	}
}

func passwordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		if u == nil {
			http.Error(w, "Must be logged in", 401)
			return
		}

		var errmsg string
		var f struct{ pwd, newpwd, newpwd2 string }

		if r.Method == "POST" {
			f.pwd = r.FormValue("pwd")
			f.newpwd = r.FormValue("newpwd")
			f.newpwd2 = r.FormValue("newpwd2")
			for {
				if f.newpwd != f.newpwd2 {
					errmsg = "passwords don't match"
					break
				}
				err := edituser(db, u.Username, f.pwd, f.newpwd)
				if err != nil {
					errmsg = fmt.Sprintf("%s", err)
					break
				}
				tok, err := login(db, u.Username, f.newpwd)
				if err != nil {
					errmsg = fmt.Sprintf("%s", err)
					break
				}
				setLoginCookie(w, u.Username, tok)

				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printHeading(P, u)

		printFormSmallOpen(P, "/password/", "Change Password")
		printFormInputPassword(P, "pwd", "password", f.pwd)
		printFormInputPassword(P, "newpwd", "new password", f.newpwd)
		printFormInputPassword(P, "newpwd2", "re-enter password", f.newpwd2)
		printFormError(P, errmsg)
		printFormSubmit(P, "Submit")
		printFormLinks(P, "justify-end", "/", "Cancel")
		printFormClose(P)

		printHtmlClose(P)
	}
}

func profileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		if u == nil {
			http.Error(w, "Must be logged in", 401)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printHeading(P, u)

		printDivSmallOpen(P, escape(u.Username))
		printDivFlex(P, "justify-start")
		P("<div class=\"px-4\">\n")
		P("    <a href=\"/password\" class=\"action block border-b\">Change Password</a>\n")
		P("    <a href=\"#\" class=\"action block border-b\">Delete Account</a>\n")
		P("</div>\n")
		P("<div class=\"px-4\">\n")
		P("</div>\n")
		P("<div class=\"px-4\">\n")
		P("</div>\n")
		printDivClose(P)
		printDivClose(P)

		printHtmlClose(P)
	}
}

func createEntry(db *sql.DB, e *Entry) (int64, error) {
	s := "INSERT INTO entry (title, body, createdt, user_id) VALUES (?, ?, ?, ?)"
	result, err := sqlexec(db, s, e.Title, e.Body, e.Createdt, e.Userid)
	if err != nil {
		return 0, err
	}
	entryid, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return entryid, nil
}
func addentryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		if u == nil {
			http.Error(w, "Must be logged in", 401)
			return
		}

		var errmsg string
		var e Entry
		var tags string

		if r.Method == "POST" {
			e.Title = r.FormValue("title")
			e.Body = r.FormValue("body")
			e.Createdt = isodate(time.Now())
			e.Userid = u.Userid
			tags = r.FormValue("tags")

			for {
				if e.Title == "" {
					errmsg = "enter a title"
					break
				}
				_, err := createEntry(db, &e)
				if err != nil {
					logErr("createEntry", err)
					errmsg = "server error adding entry"
					break
				}

				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printHeading(P, u)

		printFormOpen(P, "/addentry/", "New Entry")
		printFormInput(P, "title", "title", e.Title)
		printFormTextarea(P, "body", "entry", e.Body, "")
		printFormInput(P, "tags", "tags", tags)
		printFormError(P, errmsg)
		printFormSubmit(P, "Submit")
		printFormLinks(P, "justify-end", "/", "Cancel")
		printFormClose(P)

		printHtmlClose(P)
	}
}

func entryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)

		id := idtoi(r.FormValue("id"))
		if id == 0 {
			http.Error(w, "Not found.", 404)
			return
		}
		entry := findEntry(db, id)
		if entry == nil {
			http.Error(w, "Not found.", 404)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printHeading(P, u)
		printEntry(P, db, u, entry)
		printHtmlClose(P)
	}
}
func printEntry(P PrintFunc, db *sql.DB, u *User, e *Entry) {
	var authorname string
	author := findUserById(db, e.Userid)
	if author != nil {
		authorname = author.Username
	}

	P("<h1 class=\"font-bold text-2xl mb-2\">%s</h1>\n", escape(e.Title))
	if authorname != "" {
		P("<p class=\"mb-4 text-sm\">Posted on <span class=\"italic\">%s</span> by <a href=\"#\" class=\"action\">%s</a></p>\n", formatdate(e.Createdt), authorname)
	} else {
		P("<p class=\"mb-4 text-sm\">Posted on <span class=\"italic\">%s</span></p>\n", formatdate(e.Createdt))
	}
	P("<div class=\"content\">\n")
	P("%s\n", parseMarkdown(e.Body))
	P("</div>\n")
}

func indexHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		ee, err := findEntries(db)
		if handleDbErr(w, err, "indexHandler") {
			return
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printHeading(P, u)

		P("<h1 class=\"font-bold text-xl mb-2\">Latest Posts</h1>\n")
		for _, e := range ee {
			P("<div class=\"flex flex-row py-1\">\n")
			P("    <p class=\"text-xs text-gray-700\">%s</p>\n", formatdate(e.Createdt))
			P("    <p class=\"flex-grow px-4\">\n")
			P("        <a class=\"action font-bold\" href=\"/entry?id=%d\">%s</a>\n", e.Entryid, escape(e.Title))
			P("    </p>\n")
			P("    <a class=\"text-xs text-gray-700 px-2\" href=\"/?username=%s\">%s</a>\n", qescape(u.Username), escape(u.Username))
			P("</div>\n")

			/*
				if u.Userid == e.Userid {
					P("        <a class=\"px-2 py-1 rounded mx-1 pill text-xs\" href=\"/editentry?id=%d\">Edit</a>\n", e.Entryid)
				}
			*/
		}

		printHtmlClose(P)
	}
}
