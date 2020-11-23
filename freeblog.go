package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	Entryid  int64  `json:"entryid"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Createdt string `json:"createdt"`
	Userid   int64  `json:"userid"`
	Username string `json:"username"`
}
type File struct {
	Fileid   int64  `json:"fileid"`
	Filename string `json:"filename"`
	Url      string `json:"url"`
	Bytes    []byte `json:"-"`
	Createdt string `json:"createdt"`
	Userid   int64  `json:"userid"`
	Username string `json:"username"`
}

func (e *Entry) String() string {
	bs, err := json.MarshalIndent(e, "", "\t")
	if err != nil {
		return ""
	}
	return string(bs)
}
func fileurl(f *File) string {
	return fmt.Sprintf("/file/?filename=%s", f.Filename)
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
	http.HandleFunc("/entry/", entryHandler(db))
	http.HandleFunc("/file/", fileHandler(db))
	http.HandleFunc("/login/", loginHandler(db))
	http.HandleFunc("/logout/", logoutHandler(db))
	http.HandleFunc("/signup/", signupHandler(db))
	http.HandleFunc("/password/", passwordHandler(db))
	http.HandleFunc("/profile/", profileHandler(db))
	http.HandleFunc("/addentry/", addentryHandler(db))
	http.HandleFunc("/editentry/", editentryHandler(db))
	http.HandleFunc("/dashboard/", dashboardHandler(db))

	http.HandleFunc("/api/entry/", apientryHandler(db))
	http.HandleFunc("/api/entries/", apientriesHandler(db))
	http.HandleFunc("/api/uploadfiles/", apiuploadfilesHandler(db))
	http.HandleFunc("/api/files/", apifilesHandler(db))

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
		"CREATE TABLE file (file_id INTEGER PRIMARY KEY NOT NULL, filename TEXT, bytes BLOB, createdt TEXT NOT NULL, user_id INTEGER NOT NULL));",
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
	s := `SELECT entry_id, title, body, createdt, IFNULL(u.user_id, 0), IFNULL(u.username, '') 
FROM entry e
LEFT OUTER JOIN user u ON u.user_id = e.user_id 
WHERE entry_id = ?`
	row := db.QueryRow(s, entryid)
	var e Entry
	err := row.Scan(&e.Entryid, &e.Title, &e.Body, &e.Createdt, &e.Userid, &e.Username)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &e
}
func findEntries(db *sql.DB, qusername string, qlimit, qoffset int) ([]*Entry, error) {
	swhere := "1 = 1"
	var qq []interface{}

	if qusername != "" {
		swhere += " AND u.username = ?"
		qq = append(qq, qusername)
	}
	if qlimit == 0 {
		// Use an arbitrarily large number to indicate no limit
		qlimit = 10000
	}
	qq = append(qq, qlimit, qoffset)

	s := fmt.Sprintf(`SELECT entry_id, title, body, createdt, IFNULL(u.user_id, 0), IFNULL(u.username, '') 
FROM entry e
LEFT OUTER JOIN user u ON u.user_id = e.user_id 
WHERE %s 
ORDER BY entry_id DESC 
LIMIT ? OFFSET ?`, swhere)
	rows, err := db.Query(s, qq...)
	if err != nil {
		return nil, err
	}
	ee := []*Entry{}
	for rows.Next() {
		var e Entry
		rows.Scan(&e.Entryid, &e.Title, &e.Body, &e.Createdt, &e.Userid, &e.Username)
		ee = append(ee, &e)
	}
	return ee, nil
}

func findFile(db *sql.DB, fileid int64) *File {
	s := `SELECT file_id, filename, bytes, createdt, IFNULL(u.user_id, 0), IFNULL(u.username, '') 
FROM file f
LEFT OUTER JOIN user u ON u.user_id = f.user_id 
WHERE file_id = ?`
	row := db.QueryRow(s, fileid)
	var f File
	err := row.Scan(&f.Fileid, &f.Filename, &f.Bytes, &f.Createdt, &f.Userid, &f.Username)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &f
}
func findFileByFilename(db *sql.DB, filename string) *File {
	s := `SELECT file_id, filename, bytes, createdt, IFNULL(u.user_id, 0), IFNULL(u.username, '') 
FROM file f
LEFT OUTER JOIN user u ON u.user_id = f.user_id 
WHERE filename = ?`
	row := db.QueryRow(s, filename)
	var f File
	err := row.Scan(&f.Fileid, &f.Filename, &f.Bytes, &f.Createdt, &f.Userid, &f.Username)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &f
}
func findFiles(db *sql.DB, qusername, qfilename string, qlimit, qoffset int) ([]*File, error) {
	swhere := "1 = 1"
	var qq []interface{}

	if qusername != "" {
		swhere += " AND u.username = ?"
		qq = append(qq, qusername)
	}
	if qfilename != "" {
		swhere += " AND filename LIKE ?"
		qq = append(qq, fmt.Sprintf("%%%s%%", qfilename))
	}
	if qlimit == 0 {
		// Use an arbitrarily large number to indicate no limit
		qlimit = 10000
	}
	qq = append(qq, qlimit, qoffset)

	s := fmt.Sprintf(`SELECT file_id, filename, createdt, IFNULL(u.user_id, 0), IFNULL(u.username, '') 
FROM file f
LEFT OUTER JOIN user u ON u.user_id = f.user_id 
WHERE %s 
ORDER BY file_id DESC 
LIMIT ? OFFSET ?`, swhere)
	rows, err := db.Query(s, qq...)
	if err != nil {
		return nil, err
	}
	ff := []*File{}
	for rows.Next() {
		var f File
		rows.Scan(&f.Fileid, &f.Filename, &f.Createdt, &f.Userid, &f.Username)
		f.Url = fileurl(&f)
		ff = append(ff, &f)
	}
	return ff, nil
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
	P("<body class=\"text-base leading-6 myfont light\">\n")
}
func printHtmlClose(P PrintFunc) {
	P("</body>\n")
	P("</html>\n")
}
func printContainerOpen(P PrintFunc) {
	P("<div id=\"container\" class=\"flex flex-col py-2 mx-auto max-w-screen-sm\">\n")
}
func printWideContainerOpen(P PrintFunc) {
	P("<div id=\"container\" class=\"flex flex-col py-2 mx-auto max-w-screen-lg\">\n")
}
func printContainerHscreenOpen(P PrintFunc) {
	P("<div id=\"container\" class=\"flex flex-col py-2 mx-auto max-w-screen-sm h-screen\">\n")
}
func printContainerClose(P PrintFunc) {
	P("</div>\n")
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
	P("<form action=\"%s\" method=\"post\" class=\"flex-grow flex flex-col panel mx-auto py-2 px-8 text-sm w-full\">\n", action)
	if heading != "" {
		P("    <h1 class=\"font-bold mx-auto mb-2 text-center text-base\">%s</h1>\n", heading)
	}
}
func printFormSmallOpen(P PrintFunc, action, heading string) {
	P("<form action=\"%s\" method=\"post\" class=\"flex-grow flex flex-col panel mx-auto py-2 px-8 text-sm max-w-sm\">\n", action)
	if heading != "" {
		P("    <h1 class=\"font-bold mx-auto mb-2 text-center text-base\">%s</h1>\n", heading)
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
func printFormTextarea(P PrintFunc, id, label, val string) {
	P("<div class=\"flex-grow flex flex-col mb-2\">\n")
	P("    <label class=\"block font-bold uppercase text-xs\" for=\"%s\">%s</label>\n", id, label)
	P("    <textarea class=\"flex-grow block border border-gray-500 py-1 px-4 w-full leading-5\" id=\"%[1]s\" name=\"%[1]s\">%s</textarea>\n", id, val)
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
		P("    <h1 class=\"font-bold mx-auto mb-2 text-center text-base\">%s</h1>\n", heading)
	}
}
func printDivSmallOpen(P PrintFunc, heading string) {
	P("<div class=\"panel mx-auto py-4 px-8 text-sm max-w-sm\">\n")
	if heading != "" {
		P("    <h1 class=\"font-bold mx-auto mb-2 text-center text-base\">%s</h1>\n", heading)
	}
}
func printDivFlex(P PrintFunc, justify string) {
	P("<div class=\"flex flex-row %s\">\n", justify)
}
func printDivClose(P PrintFunc) {
	P("</div>\n")
}

func indexHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		ee, err := findEntries(db, "", 0, 0)
		if handleDbErr(w, err, "indexHandler") {
			return
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printContainerOpen(P)
		printHeading(P, u)

		P("<h1 class=\"font-bold text-lg mb-2\">Latest Posts</h1>\n")
		for _, e := range ee {
			P("<div class=\"flex flex-row py-1\">\n")
			P("    <p class=\"text-xs text-gray-700\">%s</p>\n", formatdate(e.Createdt))
			P("    <p class=\"flex-grow px-4\">\n")
			P("        <a class=\"action font-bold\" href=\"/entry?id=%d\">%s</a>\n", e.Entryid, escape(e.Title))
			P("    </p>\n")
			P("    <a class=\"text-xs text-gray-700 px-2\" href=\"/?username=%s\">%s</a>\n", qescape(e.Username), escape(e.Username))
			P("</div>\n")

			/*
				if u.Userid == e.Userid {
					P("        <a class=\"px-2 py-1 rounded mx-1 pill text-xs\" href=\"/editentry?id=%d\">Edit</a>\n", e.Entryid)
				}
			*/
		}

		printContainerClose(P)
		printHtmlClose(P)
	}
}

func entryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)

		qid := idtoi(r.FormValue("id"))
		if qid == 0 {
			http.Error(w, "Not found.", 404)
			return
		}
		e := findEntry(db, qid)
		if e == nil {
			http.Error(w, "Not found.", 404)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printContainerOpen(P)
		printHeading(P, u)

		printEntry(P, db, e, u)

		printContainerClose(P)
		printHtmlClose(P)
	}
}
func printEntry(P PrintFunc, db *sql.DB, e *Entry, u *User) {
	P("<h1 class=\"font-bold text-2xl mb-2\">%s</h1>\n", escape(e.Title))
	if e.Username != "" {
		P("<p class=\"mb-4 text-sm\">Posted on \n")
		P("    <span class=\"italic\">%s</span> by \n", formatdate(e.Createdt))
		P("    <a href=\"#\" class=\"action\">%s</a>\n", e.Username)
		if u != nil && e.Userid == u.Userid {
			P("    <a href=\"/editentry?id=%d\" class=\"pill rounded px-2 py-1 mx-1\">Edit</a>\n", e.Entryid)
		}
		P("</p>\n")
	} else {
		P("<p class=\"mb-4 text-sm\">Posted on <span class=\"italic\">%s</span></p>\n", formatdate(e.Createdt))
	}
	P("<div class=\"content\">\n")
	P("%s\n", parseMarkdown(e.Body))
	P("</div>\n")
}

func fileext(filename string) string {
	ss := strings.Split(filename, ".")
	if len(ss) < 2 {
		return ""
	}
	return strings.ToLower(ss[len(ss)-1])
}

// GET /file?id=123
// GET /file?filename=file1.jpg
func fileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qid := idtoi(r.FormValue("id"))
		qfilename := r.FormValue("filename")
		if qid == 0 && qfilename == "" {
			http.Error(w, "Specify id or filename (id=nnn or filename=file1.jpg)", 401)
			return
		}
		var f *File
		if qid > 0 {
			f = findFile(db, qid)
		} else if qfilename != "" {
			f = findFileByFilename(db, qfilename)
		}
		if f == nil {
			http.Error(w, "Not found.", 404)
			return
		}

		ext := fileext(f.Filename)
		if ext == "" {
			w.Header().Set("Content-Type", "application")
		} else if ext == "png" || ext == "gif" || ext == "bmp" {
			w.Header().Set("Content-Type", fmt.Sprintf("image/%s", ext))
		} else if ext == "jpg" || ext == "jpeg" {
			w.Header().Set("Content-Type", fmt.Sprintf("image/jpeg"))
		} else {
			w.Header().Set("Content-Type", fmt.Sprintf("application/%s", ext))
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", f.Filename))
		_, err := w.Write(f.Bytes)
		if err != nil {
			log.Printf("Error writing file '%s' (%s)\n", f.Filename, err)
		}
	}
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
		printContainerOpen(P)
		printHeading(P, u)

		printFormSmallOpen(P, "/login/", "Log In")
		printFormInput(P, "username", "username", f.username)
		printFormInputPassword(P, "pwd", "password", f.pwd)
		printFormError(P, errmsg)
		printFormSubmit(P, "Login")
		printFormLinks(P, "", "/signup", "Create New Account", "/", "Cancel")
		printFormClose(P)

		printContainerClose(P)
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
		printContainerOpen(P)
		printHeading(P, u)

		printFormSmallOpen(P, "/signup/", "Sign Up")
		printFormInput(P, "username", "username", f.username)
		printFormInputPassword(P, "pwd", "password", f.pwd)
		printFormInputPassword(P, "pwd2", "re-enter password", f.pwd2)
		printFormError(P, errmsg)
		printFormSubmit(P, "Sign Up")
		printFormLinks(P, "justify-end", "/", "Cancel")
		printFormClose(P)

		printContainerClose(P)
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
		printContainerOpen(P)
		printHeading(P, u)

		printFormSmallOpen(P, "/password/", "Change Password")
		printFormInputPassword(P, "pwd", "password", f.pwd)
		printFormInputPassword(P, "newpwd", "new password", f.newpwd)
		printFormInputPassword(P, "newpwd2", "re-enter password", f.newpwd2)
		printFormError(P, errmsg)
		printFormSubmit(P, "Submit")
		printFormLinks(P, "justify-end", "/", "Cancel")
		printFormClose(P)

		printContainerClose(P)
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
		printContainerOpen(P)
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

		printContainerClose(P)
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
func editEntry(db *sql.DB, e *Entry) error {
	s := "UPDATE entry SET title = ?, body = ? WHERE entry_id = ?"
	_, err := sqlexec(db, s, e.Title, e.Body, e.Entryid)
	if err != nil {
		return err
	}
	return nil
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
		printContainerHscreenOpen(P)
		printHeading(P, u)

		printFormOpen(P, "/addentry/", "New Entry")
		printFormInput(P, "title", "title", e.Title)
		printFormTextarea(P, "body", "entry", e.Body)
		printFormInput(P, "tags", "tags", tags)
		printFormError(P, errmsg)
		printFormSubmit(P, "Submit")
		printFormLinks(P, "justify-end", "/", "Cancel")
		printFormClose(P)

		printContainerClose(P)
		printHtmlClose(P)
	}
}
func editentryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		if u == nil {
			http.Error(w, "Must be logged in", 401)
			return
		}

		qid := idtoi(r.FormValue("id"))
		if qid == 0 {
			http.Error(w, "Not found.", 404)
			return
		}
		e := findEntry(db, qid)
		if e == nil {
			http.Error(w, "Not found.", 404)
			return
		}

		var errmsg string
		var tags string

		if r.Method == "POST" {
			e.Title = r.FormValue("title")
			e.Body = r.FormValue("body")
			tags = r.FormValue("tags")

			for {
				if e.Title == "" {
					errmsg = "enter a title"
					break
				}
				err := editEntry(db, e)
				if err != nil {
					logErr("editEntry", err)
					errmsg = "server error edit entry"
					break
				}

				http.Redirect(w, r, fmt.Sprintf("/entry?id=%d", e.Entryid), http.StatusSeeOther)
				return
			}
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printContainerHscreenOpen(P)
		printHeading(P, u)

		printFormOpen(P, fmt.Sprintf("/editentry/?id=%d", e.Entryid), "Edit Entry")
		printFormInput(P, "title", "title", e.Title)
		printFormTextarea(P, "body", "entry", e.Body)
		printFormInput(P, "tags", "tags", tags)
		printFormError(P, errmsg)
		printFormSubmit(P, "Submit")
		printFormLinks(P, "justify-end", "/", "Cancel")
		printFormClose(P)

		printContainerClose(P)
		printHtmlClose(P)
	}
}

func dashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := validateLoginCookie(db, r)
		if u == nil {
			http.Error(w, "Must be logged in", 401)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", []string{"/static/bundle.js", "/static/dashboard.js"})
		printWideContainerOpen(P)
		printHeading(P, u)

		printContainerClose(P)
		printHtmlClose(P)
	}
}

// GET /api/entry?id=123
// POST /api/entry {...}
// PUT /api/entry {...}
func apientryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			qid := idtoi(r.FormValue("id"))
			if qid == 0 {
				http.Error(w, "Not found.", 404)
				return
			}
			e := findEntry(db, qid)
			if e == nil {
				http.Error(w, "Not found.", 404)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s\n", e)
			return
		} else if r.Method == "POST" {
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				handleErr(w, err, "POST apientryHandler")
				return
			}
			var e Entry
			err = json.Unmarshal(bs, &e)
			if err != nil {
				handleErr(w, err, "POST apientryHandler")
				return
			}
			newid, err := createEntry(db, &e)
			if err != nil {
				handleErr(w, err, "POST apientryHandler")
				return
			}
			e.Entryid = newid

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s\n", &e)
			return
		} else if r.Method == "PUT" {
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				handleErr(w, err, "PUT apientryHandler")
				return
			}
			var e Entry
			err = json.Unmarshal(bs, &e)
			if err != nil {
				handleErr(w, err, "PUT apientryHandler")
				return
			}
			err = editEntry(db, &e)
			if err != nil {
				handleErr(w, err, "PUT apientryHandler")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s\n", &e)
			return
		}

		http.Error(w, "Use GET/POST/PUT", 401)
	}
}

// GET /api/entries
// GET /api/entries?username=rob
// GET /api/entries?limit=10
// GET /api/entries?limit=10&offset=20
func apientriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ee []*Entry
		var err error

		qusername := r.FormValue("username")
		qlimit := atoi(r.FormValue("limit"))
		qoffset := atoi(r.FormValue("offset"))

		ee, err = findEntries(db, qusername, qlimit, qoffset)
		if err != nil {
			handleErr(w, err, "apientriesHandler")
		}

		bs, err := json.MarshalIndent(ee, "", "\t")
		if err != nil {
			handleErr(w, err, "apientriesHandler")
		}
		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		P("%s\n", string(bs))
	}
}

// If another file has the same filename, add a (n) to make unique.
// Ex. "Abc File", "Abc File (1)", "Abc File (2)", etc.
func makeUniqueFilename(db *sql.DB, filename string) string {
	uniqueFilename := filename
	for i := 1; i < 100; i++ {
		f := findFileByFilename(db, uniqueFilename)
		if f == nil {
			return uniqueFilename
		}
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		uniqueFilename = fmt.Sprintf("%s (%d)%s", base, i, ext)
	}

	// If code reached this point, that means we're at (100) which is very unusual.
	// Rather than risk an overly long loop, let's just cap it off at (100).
	return uniqueFilename
}

func createFile(db *sql.DB, f *File) (int64, error) {
	f.Filename = makeUniqueFilename(db, f.Filename)

	s := "INSERT INTO file (filename, bytes, createdt, user_id) VALUES (?, ?, ?, ?)"
	result, err := sqlexec(db, s, f.Filename, f.Bytes, f.Createdt, f.Userid)
	if err != nil {
		return 0, err
	}
	fileid, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return fileid, nil
}

func apiuploadfilesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		r.ParseMultipartForm(32 << 20)
		hh := r.MultipartForm.File["files"]
		for _, h := range hh {
			f, err := h.Open()
			if err != nil {
				handleErr(w, err, "apiuploadfilesHandler")
				return
			}
			defer f.Close()

			var file File
			bs, err := ioutil.ReadAll(f)
			if err != nil {
				handleErr(w, err, "apiuploadfilesHandler")
				return
			}
			file.Filename = h.Filename
			file.Bytes = bs
			file.Createdt = isodate(time.Now())

			_, err = createFile(db, &file)
			if err != nil {
				handleErr(w, err, "apiuploadfilesHandler")
				return
			}
		}
	}
}

func apifilesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ff []*File
		var err error

		qusername := r.FormValue("username")
		qfilename := r.FormValue("filename")
		qlimit := atoi(r.FormValue("limit"))
		qoffset := atoi(r.FormValue("offset"))

		ff, err = findFiles(db, qusername, qfilename, qlimit, qoffset)
		if err != nil {
			handleErr(w, err, "apifilesHandler")
		}

		bs, err := json.MarshalIndent(ff, "", "\t")
		if err != nil {
			handleErr(w, err, "apifilesHandler")
		}
		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		P("%s\n", string(bs))
	}
}
