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
	Userid    int64  `json:"userid"`
	Username  string `json:"username"`
	HashedPwd string `json:"hashedpwd"`
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
	Title    string `json:"title"`
	Url      string `json:"url"`
	Bytes    []byte `json:"bytes"`
	Createdt string `json:"createdt"`
	Userid   int64  `json:"userid"`
	Username string `json:"username"`
}
type Site struct {
	Siteid  int64  `json:"siteid"`
	Title   string `json:"title"`
	About   string `json:"about"`
	IsGroup bool   `json:"isgroup"`
}
type UserSettings struct {
	Userid    int64  `json:"userid"`
	BlogTitle string `json:"blogtitle"`
	BlogAbout string `json:"blogabout"`
}
type PageParams struct {
	IsGroup      bool
	BlogTitle    string
	BlogUsername string
	BlogUserid   int64
}

func jsonstr(v interface{}) string {
	bs, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return ""
	}
	return string(bs)
}
func fileurl(f *File) string {
	return fmt.Sprintf("/?page=file&id=%d", f.Fileid)
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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", rootHandler(db))
	http.HandleFunc("/api/entry/", apientryHandler(db))
	http.HandleFunc("/api/entries/", apientriesHandler(db))
	http.HandleFunc("/api/uploadfiles/", apiuploadfilesHandler(db))
	http.HandleFunc("/api/file/", apifileHandler(db))
	http.HandleFunc("/api/files/", apifilesHandler(db))
	http.HandleFunc("/api/site/", apisiteHandler(db))
	http.HandleFunc("/api/usersettings/", apiusersettingsHandler(db))

	http.HandleFunc("/api/changepwd/", apichangepwdHandler(db))
	http.HandleFunc("/api/deluser/", apideluserHandler(db))
	http.HandleFunc("/api/login/", apiloginHandler(db))
	http.HandleFunc("/api/logout/", apilogoutHandler(db))

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
		"CREATE TABLE site (site_id INTEGER PRIMARY KEY NOT NULL, title TEXT, about TEXT, isgroup INTEGER);",
		"CREATE TABLE user (user_id INTEGER PRIMARY KEY NOT NULL, username TEXT UNIQUE, password TEXT);",
		"CREATE TABLE usersettings (user_id INTEGER PRIMARY KEY NOT NULL, blogtitle TEXT, blogabout TEXT);",
		"INSERT INTO user (user_id, username, password) VALUES (1, 'admin', '');",
		"CREATE TABLE entry (entry_id INTEGER PRIMARY KEY NOT NULL, title TEXT, body TEXT, createdt TEXT NOT NULL, user_id INTEGER NOT NULL);",
		"CREATE TABLE entrytag (entry_id INTEGER NOT NULL, tag TEXT NOT NULL);",
		"CREATE TABLE file (file_id INTEGER PRIMARY KEY NOT NULL, filename TEXT, title TEXT, bytes BLOB, createdt TEXT NOT NULL, user_id INTEGER NOT NULL);",
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
func pathescape(s string) string {
	return url.PathEscape(s)
}
func pathunescape(s string) string {
	us, err := url.PathUnescape(s)
	if err != nil {
		us = s
	}
	return us
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

var DefaultSite = Site{
	Siteid:  1,
	Title:   "FreeBlog",
	About:   "(about text here)",
	IsGroup: false,
}
var DefaultUserSettings = UserSettings{
	Userid:    0,
	BlogTitle: "My Blog",
	BlogAbout: "(about text here)",
}

func findSite(db *sql.DB) *Site {
	s := "SELECT site_id, title, about, isgroup FROM site WHERE site_id = ?"
	row := db.QueryRow(s, 1)
	var site Site
	err := row.Scan(&site.Siteid, &site.Title, &site.About, &site.IsGroup)
	if err != nil {
		return &DefaultSite
	}
	return &site
}
func createSite(db *sql.DB, site *Site) error {
	s := "INSERT OR REPLACE INTO site (site_id, title, about, isgroup) VALUES (?, ?, ?, ?)"
	_, err := sqlexec(db, s, 1, site.Title, site.About, site.IsGroup)
	return err
}

func findUserSettingsById(db *sql.DB, userid int64) *UserSettings {
	s := "SELECT user_id, blogtitle, blogabout FROM usersettings WHERE user_id = ?"
	row := db.QueryRow(s, userid)
	var us UserSettings
	err := row.Scan(&us.Userid, &us.BlogTitle, &us.BlogAbout)
	if err != nil {
		us.Userid = userid
		us.BlogTitle = DefaultUserSettings.BlogTitle
		us.BlogAbout = DefaultUserSettings.BlogAbout
		return &us
	}
	return &us
}
func createUserSettings(db *sql.DB, us *UserSettings) error {
	s := "INSERT OR REPLACE INTO usersettings (user_id, blogtitle, blogabout) VALUES (?, ?, ?)"
	_, err := sqlexec(db, s, us.Userid, us.BlogTitle, us.BlogAbout)
	return err
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
func findUserByUsername(db *sql.DB, username string) *User {
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
	if findUserByUsername(db, username) == nil {
		return false
	}
	return true
}

func genSig(u *User) string {
	sig := genHash(fmt.Sprintf("%s_%s", u.Username, u.HashedPwd))
	return sig
}
func validateSig(sig string, u *User) bool {
	return validateHash(sig, fmt.Sprintf("%s_%s", u.Username, u.HashedPwd))
}

func setCookie(w http.ResponseWriter, name, val string) {
	c := http.Cookie{
		Name:  name,
		Value: val,
		Path:  "/",
		// Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, &c)
}
func delCookie(w http.ResponseWriter, name string) {
	c := http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: 0,
	}
	http.SetCookie(w, &c)
}
func readCookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}
func setLoginCookie(w http.ResponseWriter, u *User, sig string) {
	setCookie(w, "userid", itoa(u.Userid))
	setCookie(w, "username", u.Username)
	setCookie(w, "sig", sig)
}
func delLoginCookie(w http.ResponseWriter) {
	delCookie(w, "userid")
	delCookie(w, "username")
	delCookie(w, "sig")
}
func readLoginCookie(r *http.Request) (*User, string) {
	var u User
	u.Userid = idtoi(readCookie(r, "userid"))
	u.Username = readCookie(r, "username")
	sig := readCookie(r, "sig")
	return &u, sig
}

// Reads and validates login cookie. If invalid user/sig, return no user.
func validateLoginCookie(db *sql.DB, r *http.Request) (*User, string) {
	utmp, sig := readLoginCookie(r)
	if utmp.Userid == 0 {
		return nil, ""
	}
	u := findUserById(db, utmp.Userid)
	if u == nil {
		return nil, ""
	}
	if !validateSig(sig, u) {
		log.Printf("Invalid signature for '%s' ", u.Username)
		return nil, ""
	}
	return u, sig
}

// Pass either userid or username (userid takes precedence) and signature to validate.
func validateUserSig(db *sql.DB, userid int64, username string, sig string) (*User, string) {
	var u *User
	if userid != 0 {
		u = findUserById(db, userid)
	} else {
		u = findUserByUsername(db, username)
	}
	if u == nil {
		return nil, ""
	}
	if !validateSig(sig, u) {
		log.Printf("Invalid signature for '%s' ", u.Username)
		return nil, ""
	}
	return u, sig
}

func validateApiUser(db *sql.DB, r *http.Request) *User {
	// Get user making the request. There are two ways to specify user:
	// - Through querystring either userid or username and sig
	// - Through http cookies 'userid' and 'sig'
	quserid := idtoi(r.FormValue("userid"))
	qusername := r.FormValue("username")
	qsig := r.FormValue("sig")

	var u *User
	if quserid > 0 || qusername != "" {
		u, _ = validateUserSig(db, quserid, qusername, qsig)
	} else {
		u, _ = validateLoginCookie(db, r)
	}
	return u
}

var ErrLoginIncorrect = errors.New("Incorrect username or password")

func loginUserid(db *sql.DB, userid int64, pwd string) (*User, string, error) {
	u := findUserById(db, userid)
	if u == nil {
		return nil, "", ErrLoginIncorrect
	}
	sig, err := login(u, pwd)
	if err != nil {
		return nil, "", err
	}
	return u, sig, nil
}
func loginUsername(db *sql.DB, username, pwd string) (*User, string, error) {
	u := findUserByUsername(db, username)
	if u == nil {
		return nil, "", ErrLoginIncorrect
	}
	sig, err := login(u, pwd)
	if err != nil {
		return nil, "", err
	}
	return u, sig, nil
}
func login(u *User, pwd string) (string, error) {
	if !validateHash(u.HashedPwd, pwd) {
		return "", ErrLoginIncorrect
	}
	// Return user signature, this will be used to authenticate user per request.
	sig := genSig(u)
	return sig, nil
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

func edituser(db *sql.DB, userid int64, pwd string, newpwd string) error {
	// Validate existing password
	_, _, err := loginUserid(db, userid, pwd)
	if err != nil {
		return err
	}

	// Set new password
	hashedPwd := genHash(newpwd)
	s := "UPDATE user SET password = ? WHERE user_id = ?"
	_, err = sqlexec(db, s, hashedPwd, userid)
	if err != nil {
		return fmt.Errorf("DB error updating user password: %s", err)
	}
	return nil
}

func deluser(db *sql.DB, userid int64, pwd string) error {
	// Validate existing password
	_, _, err := loginUserid(db, userid, pwd)
	if err != nil {
		return err
	}

	// Delete user
	s := "DELETE FROM user WHERE user_id = ?"
	_, err = sqlexec(db, s, userid)
	if err != nil {
		return fmt.Errorf("DB error deleting user: %s", err)
	}
	return nil
}
func transferUserEntries(db *sql.DB, fromUserid, toUserid int64) error {
	s := "UPDATE entry SET user_id = ? WHERE user_id = ?"
	_, err := sqlexec(db, s, toUserid, fromUserid)
	if err != nil {
		return err
	}
	return nil
}
func transferUserFiles(db *sql.DB, fromUserid, toUserid int64) error {
	s := "UPDATE file SET user_id = ? WHERE user_id = ?"
	_, err := sqlexec(db, s, toUserid, fromUserid)
	if err != nil {
		return err
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
func findEntries(db *sql.DB, quserid int64, qlimit, qoffset int) ([]*Entry, error) {
	swhere := "1 = 1"
	var qq []interface{}

	if quserid != 0 {
		swhere += " AND u.user_id = ?"
		qq = append(qq, quserid)
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
	s := `SELECT file_id, filename, title, bytes, createdt, IFNULL(u.user_id, 0), IFNULL(u.username, '') 
FROM file f
LEFT OUTER JOIN user u ON u.user_id = f.user_id 
WHERE file_id = ?`
	row := db.QueryRow(s, fileid)
	var f File
	err := row.Scan(&f.Fileid, &f.Filename, &f.Title, &f.Bytes, &f.Createdt, &f.Userid, &f.Username)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &f
}
func findFileByFilename(db *sql.DB, filename string) *File {
	s := `SELECT file_id, filename, title, bytes, createdt, IFNULL(u.user_id, 0), IFNULL(u.username, '') 
FROM file f
LEFT OUTER JOIN user u ON u.user_id = f.user_id 
WHERE filename = ?`
	row := db.QueryRow(s, filename)
	var f File
	err := row.Scan(&f.Fileid, &f.Filename, &f.Title, &f.Bytes, &f.Createdt, &f.Userid, &f.Username)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &f
}
func findFiles(db *sql.DB, quserid int64, qfilename string, qlimit, qoffset int) ([]*File, error) {
	swhere := "1 = 1"
	var qq []interface{}

	if quserid != 0 {
		swhere += " AND u.user_id = ?"
		qq = append(qq, quserid)
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

	return findFilesWithParams(db, swhere, qq)
}
func findImageFiles(db *sql.DB, quserid int64, qfilename string, qlimit, qoffset int) ([]*File, error) {
	var qq []interface{}

	swhere := "(filename LIKE '%.png' OR filename LIKE '%.jpg' OR filename LIKE '%.jpeg' OR filename LIKE '%.gif' OR filename LIKE '%.bmp' OR filename LIKE '%.tif' OR filename LIKE '%.tiff')"

	if quserid != 0 {
		swhere += " AND u.user_id = ?"
		qq = append(qq, quserid)
	}
	if qfilename != "" {
		swhere += " AND (filename LIKE ?)"
		qq = append(qq, fmt.Sprintf("%%%s%%", qfilename))
	}
	if qlimit == 0 {
		// Use an arbitrarily large number to indicate no limit
		qlimit = 10000
	}
	qq = append(qq, qlimit, qoffset)

	return findFilesWithParams(db, swhere, qq)
}
func findAttachmentFiles(db *sql.DB, quserid int64, qfilename string, qlimit, qoffset int) ([]*File, error) {
	var qq []interface{}

	swhere := "(NOT filename LIKE '%.png' AND NOT filename LIKE '%.jpg' AND NOT filename LIKE '%.jpeg' AND NOT filename LIKE '%.gif' AND NOT filename LIKE '%.bmp' AND NOT filename LIKE '%.tif' AND NOT filename LIKE '%.tiff')"

	if quserid != 0 {
		swhere += " AND u.user_id = ?"
		qq = append(qq, quserid)
	}
	if qfilename != "" {
		swhere += " AND (filename LIKE ?)"
		qq = append(qq, fmt.Sprintf("%%%s%%", qfilename))
	}
	if qlimit == 0 {
		// Use an arbitrarily large number to indicate no limit
		qlimit = 10000
	}
	qq = append(qq, qlimit, qoffset)

	return findFilesWithParams(db, swhere, qq)
}
func findFilesWithParams(db *sql.DB, swhere string, qq []interface{}) ([]*File, error) {
	s := fmt.Sprintf(`SELECT file_id, filename, title, createdt, IFNULL(u.user_id, 0), IFNULL(u.username, '') 
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
		rows.Scan(&f.Fileid, &f.Filename, &f.Title, &f.Createdt, &f.Userid, &f.Username)
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
func printHeading(P PrintFunc, u *User, pp *PageParams) {
	P("<div class=\"flex flex-row justify-between border-b border-gray-500 pb-1 mb-4 text-sm\"\n>")
	P("    <div>\n")
	P("        <h1 class=\"inline self-end ml-1 mr-2 font-bold\"><a href=\"/%s\">%s</a></h1>\n", pp.BlogUsername, pp.BlogTitle)
	P("        <a href=\"/%s?page=about\" class=\"self-end mr-2\">About</a>\n", pp.BlogUsername)
	P("    </div>\n")
	P("    <div>\n")
	if u != nil {
		P("        <div class=\"relative inline mr-2\">\n")
		P("            <a class=\"mr-1\" href=\"/%s?page=dashboard\">%s</a>\n", pp.BlogUsername, escape(u.Username))
		P("        </div>\n")
		P("        <a href=\"/%s?page=logout\" class=\"inline self-end mr-1\">Logout</a>\n", pp.BlogUsername)
	} else {
		P("        <a href=\"/%s?page=login\" class=\"inline self-end mr-1\">Login</a>\n", pp.BlogUsername)
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
		P("    <h1 class=\"font-bold mb-2 text-base\">%s</h1>\n", heading)
	}
}
func printDivFlex(P PrintFunc, justify string) {
	P("<div class=\"flex flex-row %s\">\n", justify)
}
func printDivClose(P PrintFunc) {
	P("</div>\n")
}

func rootHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := r.FormValue("page")
		if page == "index" || page == "" {
			indexHandler(w, r, db)
		} else if page == "entry" {
			entryHandler(w, r, db)
		} else if page == "file" {
			fileHandler(w, r, db)
		} else if page == "login" {
			loginHandler(w, r, db)
		} else if page == "logout" {
			logoutHandler(w, r, db)
		} else if page == "signup" {
			signupHandler(w, r, db)
		} else if page == "dashboard" {
			dashboardHandler(w, r, db)
		}
	}
}

func parsePageUrl(r *http.Request) (string, string) {
	surl := strings.Trim(r.URL.Path, "/")
	ss := strings.Split(surl, "/")
	sslen := len(ss)
	if sslen == 0 {
		return "", ""
	} else if sslen == 1 {
		return qunescape(ss[0]), ""
	}
	return qunescape(ss[0]), qunescape(ss[1])
}
func getPageParams(r *http.Request, db *sql.DB) *PageParams {
	var pp PageParams

	site := findSite(db)
	pp.IsGroup = site.IsGroup
	pp.BlogTitle = site.Title

	var bloguser *User
	blogusername, _ := parsePageUrl(r)
	if blogusername != "" {
		bloguser = findUserByUsername(db, blogusername)
	}
	if bloguser != nil {
		pp.BlogUsername = qescape(bloguser.Username)
		pp.BlogUserid = bloguser.Userid

		if !site.IsGroup {
			us := findUserSettingsById(db, bloguser.Userid)
			pp.BlogTitle = escape(us.BlogTitle)
		}
	}

	return &pp
}
func indexHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	u, _ := validateLoginCookie(db, r)

	pp := getPageParams(r, db)
	ee, err := findEntries(db, pp.BlogUserid, 0, 0)
	if handleDbErr(w, err, "indexHandler") {
		return
	}

	w.Header().Set("Content-Type", "text/html")
	P := makeFprintf(w)
	printHtmlOpen(P, pp.BlogTitle, nil)
	printContainerOpen(P)
	printHeading(P, u, pp)

	P("<h1 class=\"font-bold text-lg mb-2\">Latest Posts</h1>\n")
	for _, e := range ee {
		P("<div class=\"flex flex-row py-1\">\n")
		P("    <p class=\"text-xs text-gray-700\">%s</p>\n", formatdate(e.Createdt))
		P("    <p class=\"flex-grow px-4\">\n")
		P("        <a class=\"action font-bold\" href=\"/%s?page=entry&id=%d\">%s</a>\n", pp.BlogUsername, e.Entryid, escape(e.Title))
		P("    </p>\n")
		P("    <a class=\"text-xs text-gray-700 px-2\" href=\"/%s\">%s</a>\n", qescape(e.Username), escape(e.Username))
		P("</div>\n")
	}

	printContainerClose(P)
	printHtmlClose(P)
}

func entryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	pp := getPageParams(r, db)
	printHtmlOpen(P, pp.BlogTitle, nil)
	printContainerOpen(P)
	printHeading(P, u, pp)

	printEntry(P, db, e)

	printContainerClose(P)
	printHtmlClose(P)
}
func printEntry(P PrintFunc, db *sql.DB, e *Entry) {
	P("<h1 class=\"font-bold text-2xl mb-2\">%s</h1>\n", escape(e.Title))
	if e.Username != "" {
		P("<p class=\"mb-4 text-sm\">Posted on \n")
		P("    <span class=\"italic\">%s</span> by \n", formatdate(e.Createdt))
		P("    <a href=\"/%s\" class=\"action\">%s</a>\n", qescape(e.Username), escape(e.Username))
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

// GET /?page=file&id=123
// GET /?page=file&filename=file1.jpg
func fileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

func loginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	u, _ := validateLoginCookie(db, r)
	pp := getPageParams(r, db)

	var errmsg string
	var f struct{ username, pwd string }

	if r.Method == "POST" {
		f.username = r.FormValue("username")
		f.pwd = r.FormValue("pwd")
		for {
			u, sig, err := loginUsername(db, f.username, f.pwd)
			if err != nil {
				errmsg = fmt.Sprintf("%s", err)
				break
			}
			setLoginCookie(w, u, sig)

			http.Redirect(w, r, fmt.Sprintf("/%s", qescape(u.Username)), http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	P := makeFprintf(w)
	printHtmlOpen(P, pp.BlogTitle, nil)
	printContainerOpen(P)
	printHeading(P, u, pp)

	printFormSmallOpen(P, fmt.Sprintf("/%s?page=login", pp.BlogUsername), "Log In")
	printFormInput(P, "username", "username", f.username)
	printFormInputPassword(P, "pwd", "password", f.pwd)
	printFormError(P, errmsg)
	printFormSubmit(P, "Login")
	printFormLinks(P, "", fmt.Sprintf("/%s?page=signup", pp.BlogUsername), "Create New Account", "/", "Cancel")
	printFormClose(P)

	printContainerClose(P)
	printHtmlClose(P)
}
func logoutHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	delLoginCookie(w)

	pp := getPageParams(r, db)
	http.Redirect(w, r, fmt.Sprintf("/%s", pp.BlogUsername), http.StatusSeeOther)
}
func signupHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	u, _ := validateLoginCookie(db, r)
	pp := getPageParams(r, db)

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
			u, sig, err := loginUsername(db, f.username, f.pwd)
			if err != nil {
				errmsg = fmt.Sprintf("%s", err)
				break
			}
			setLoginCookie(w, u, sig)

			http.Redirect(w, r, fmt.Sprintf("/%s", qescape(u.Username)), http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	P := makeFprintf(w)
	printHtmlOpen(P, pp.BlogTitle, nil)
	printContainerOpen(P)
	printHeading(P, u, pp)

	printFormSmallOpen(P, fmt.Sprintf("/%s?page=signup", pp.BlogUsername), "Sign Up")
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
func delEntry(db *sql.DB, entryid int64) error {
	s := `DELETE FROM entry WHERE entry_id = ?`
	_, err := sqlexec(db, s, entryid)
	if err != nil {
		return err
	}
	return nil
}
func createFile(db *sql.DB, f *File) (int64, error) {
	f.Filename = makeUniqueFilename(db, f.Filename)

	s := "INSERT INTO file (filename, title, bytes, createdt, user_id) VALUES (?, ?, ?, ?, ?)"
	result, err := sqlexec(db, s, f.Filename, f.Title, f.Bytes, f.Createdt, f.Userid)
	if err != nil {
		return 0, err
	}
	fileid, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return fileid, nil
}
func editFile(db *sql.DB, f *File) error {
	f.Filename = makeUniqueFilename(db, f.Filename)
	s := "UPDATE file SET filename = ?, title = ?, bytes = ? WHERE file_id = ?"
	_, err := sqlexec(db, s, f.Filename, f.Title, f.Bytes, f.Fileid)
	if err != nil {
		return err
	}
	return nil
}
func delFile(db *sql.DB, fileid int64) error {
	s := `DELETE FROM file WHERE file_id = ?`
	_, err := sqlexec(db, s, fileid)
	if err != nil {
		return err
	}
	return nil
}

func dashboardHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	u, _ := validateLoginCookie(db, r)
	if u == nil {
		http.Error(w, "Must be logged in", 401)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	P := makeFprintf(w)
	pp := getPageParams(r, db)
	printHtmlOpen(P, pp.BlogTitle, []string{"/static/bundle.js", "/static/dashboard.js"})
	printWideContainerOpen(P)
	printHeading(P, u, pp)

	printContainerClose(P)
	printHtmlClose(P)
}

func apichangepwdHandler(db *sql.DB) http.HandlerFunc {
	type Req struct {
		Userid int64  `json:"userid"`
		Pwd    string `json:"pwd"`
		Newpwd string `json:"newpwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		u := validateApiUser(db, r)
		if u == nil {
			http.Error(w, "Invalid user", 401)
			return
		}
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "POST apichangepwdHandler")
			return
		}
		var req Req
		err = json.Unmarshal(bs, &req)
		if err != nil {
			handleErr(w, err, "POST apichangepwdHandler")
			return
		}
		if u.Userid != 1 && req.Userid != u.Userid {
			http.Error(w, "Not authorized", 401)
			return
		}
		err = edituser(db, req.Userid, req.Pwd, req.Newpwd)
		if err == ErrLoginIncorrect {
			http.Error(w, err.Error(), 401)
			return
		}
		if err != nil {
			handleErr(w, err, "POST apichangepwdHandler")
			return
		}
	}
}

func apideluserHandler(db *sql.DB) http.HandlerFunc {
	type Req struct {
		Userid int64  `json:"userid"`
		Pwd    string `json:"pwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		u := validateApiUser(db, r)
		if u == nil {
			http.Error(w, "Invalid user", 401)
			return
		}
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "POST apideluserHandler")
			return
		}
		var req Req
		err = json.Unmarshal(bs, &req)
		if err != nil {
			handleErr(w, err, "POST apideluserHandler")
			return
		}
		if u.Userid != 1 && req.Userid != u.Userid {
			http.Error(w, "Not authorized", 401)
			return
		}

		err = deluser(db, req.Userid, req.Pwd)
		if err == ErrLoginIncorrect {
			http.Error(w, err.Error(), 401)
			return
		}
		if err != nil {
			handleErr(w, err, "POST apideluserHandler")
			return
		}
		// Admin takes ownership of deleted user's entries and files.
		err = transferUserEntries(db, req.Userid, 1)
		if err != nil {
			handleErr(w, err, "POST apideluserHandler")
			return
		}
		err = transferUserFiles(db, req.Userid, 1)
		if err != nil {
			handleErr(w, err, "POST apideluserHandler")
			return
		}
	}
}

func apiloginHandler(db *sql.DB) http.HandlerFunc {
	type Req struct {
		Userid int64  `json:"userid"`
		Pwd    string `json:"pwd"`
	}
	type Resp struct {
		Userid int64  `json:"userid"`
		Sig    string `json:"sig"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "POST apiloginHandler")
			return
		}
		var req Req
		err = json.Unmarshal(bs, &req)
		if err != nil {
			handleErr(w, err, "POST apiloginHandler")
			return
		}

		u, sig, err := loginUserid(db, req.Userid, req.Pwd)
		if err == ErrLoginIncorrect {
			http.Error(w, err.Error(), 401)
			return
		}
		if err != nil {
			handleErr(w, err, "POST apiloginHandler")
			return
		}
		setLoginCookie(w, u, sig)

		var resp Resp
		resp.Userid = u.Userid
		resp.Sig = sig

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		P("%s", jsonstr(resp))
	}
}

func apilogoutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST", 401)
			return
		}

		delLoginCookie(w)
	}
}

// GET /api/entry?id=123
// DELETE /api/entry?id=123
// POST /api/entry {...}
// PUT /api/entry {...}
func apientryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			qid := idtoi(r.FormValue("id"))
			qfmt := r.FormValue("fmt")
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
			if qfmt == "html" {
				printViewEntry(P, db, e)
				return
			}
			P("%s", jsonstr(e))
			return
		} else if r.Method == "POST" {
			u := validateApiUser(db, r)
			if u == nil {
				http.Error(w, "Invalid user", 401)
				return
			}
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
			e.Userid = u.Userid
			e.Createdt = isodate(time.Now())
			newid, err := createEntry(db, &e)
			if err != nil {
				handleErr(w, err, "POST apientryHandler")
				return
			}
			e.Entryid = newid

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(e))
			return
		} else if r.Method == "PUT" {
			u := validateApiUser(db, r)
			if u == nil {
				http.Error(w, "Invalid user", 401)
				return
			}
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
			if u.Userid != 1 && e.Userid != u.Userid {
				http.Error(w, "Not authorized", 401)
				return
			}
			err = editEntry(db, &e)
			if err != nil {
				handleErr(w, err, "PUT apientryHandler")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(e))
			return
		} else if r.Method == "DELETE" {
			u := validateApiUser(db, r)
			if u == nil {
				http.Error(w, "Invalid user", 401)
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
			if u.Userid != 1 && e.Userid != u.Userid {
				http.Error(w, "Not authorized", 401)
				return
			}

			err := delEntry(db, qid)
			if err != nil {
				handleErr(w, err, "DEL apientryHandler")
				return
			}
			return
		}

		http.Error(w, "Use GET/POST/PUT/DELETE", 401)
	}
}

func printViewEntry(P PrintFunc, db *sql.DB, e *Entry) {
	P("<h1 class=\"text-2xl mb-2\">%s</h1>\n", escape(e.Title))
	if e.Username != "" {
		P("<p class=\"mb-4 text-sm\">Posted on \n")
		P("    <span class=\"italic\">%s</span> by %s\n", formatdate(e.Createdt), e.Username)
		P("</p>\n")
	} else {
		P("<p class=\"mb-4 text-sm\">Posted on <span class=\"italic\">%s</span></p>\n", formatdate(e.Createdt))
	}
	P("<div class=\"content\">\n")
	P("%s\n", parseMarkdown(e.Body))
	P("</div>\n")
}

// GET /api/entries
// GET /api/entries?userid=2
// GET /api/entries?limit=10
// GET /api/entries?limit=10&offset=20
func apientriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ee []*Entry
		var err error

		quserid := idtoi(r.FormValue("userid"))
		qlimit := atoi(r.FormValue("limit"))
		qoffset := atoi(r.FormValue("offset"))

		ee, err = findEntries(db, quserid, qlimit, qoffset)
		if err != nil {
			handleErr(w, err, "apientriesHandler")
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		P("%s", jsonstr(ee))
	}
}

// If another file has the same filename, add a --n to make unique.
// Ex. "Abc File", "Abc File--1", "Abc File--2", etc.
func makeUniqueFilename(db *sql.DB, filename string) string {
	uniqueFilename := filename
	for i := 1; i < 100; i++ {
		f := findFileByFilename(db, uniqueFilename)
		if f == nil {
			return uniqueFilename
		}
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		uniqueFilename = fmt.Sprintf("%s--%d%s", base, i, ext)
	}

	// If code reached this point, that means we're at (100) which is very unusual.
	// Rather than risk an overly long loop, let's just cap it off at (100).
	return uniqueFilename
}

func baseFilename(filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	return base
}

func apiuploadfilesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		u := validateApiUser(db, r)
		if u == nil {
			http.Error(w, "Invalid user", 401)
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
			file.Title = baseFilename(h.Filename)
			file.Bytes = bs
			file.Createdt = isodate(time.Now())
			file.Userid = u.Userid
			_, err = createFile(db, &file)
			if err != nil {
				handleErr(w, err, "apiuploadfilesHandler")
				return
			}
		}
	}
}

// GET /api/file?id=123
// DELETE /api/file?id=123
// POST /api/file {...}
// PUT /api/file {...}
func apifileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			qid := idtoi(r.FormValue("id"))
			if qid == 0 {
				http.Error(w, "Specify file id (id=nnn)", 401)
				return
			}
			f := findFile(db, qid)
			if f == nil {
				http.Error(w, "Not found.", 404)
				return
			}
			f.Url = fileurl(f)

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(f))
			return
		} else if r.Method == "POST" {
			u := validateApiUser(db, r)
			if u == nil {
				http.Error(w, "Invalid user", 401)
				return
			}
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				handleErr(w, err, "POST apientryHandler")
				return
			}
			var f File
			err = json.Unmarshal(bs, &f)
			if err != nil {
				handleErr(w, err, "POST apifileHandler")
				return
			}
			f.Userid = u.Userid
			f.Createdt = isodate(time.Now())
			newid, err := createFile(db, &f)
			if err != nil {
				handleErr(w, err, "POST apifileHandler")
				return
			}
			f.Fileid = newid

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(f))
			return
		} else if r.Method == "PUT" {
			u := validateApiUser(db, r)
			if u == nil {
				http.Error(w, "Invalid user", 401)
				return
			}
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				handleErr(w, err, "PUT apifileHandler 2")
				return
			}
			var f File
			err = json.Unmarshal(bs, &f)
			if err != nil {
				handleErr(w, err, "PUT apifileHandler 3")
				return
			}
			if u.Userid != 1 && f.Userid != u.Userid {
				http.Error(w, "Not authorized", 401)
				return
			}
			err = editFile(db, &f)
			if err != nil {
				handleErr(w, err, "PUT apifileHandler 4")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(f))
			return
		} else if r.Method == "DELETE" {
			u := validateApiUser(db, r)
			if u == nil {
				http.Error(w, "Invalid user", 401)
				return
			}
			qid := idtoi(r.FormValue("id"))
			if qid == 0 {
				http.Error(w, "Not found.", 404)
				return
			}

			f := findFile(db, qid)
			if f == nil {
				http.Error(w, "Not found.", 404)
				return
			}
			if u.Userid != 1 && f.Userid != u.Userid {
				http.Error(w, "Not authorized", 401)
				return
			}

			err := delFile(db, qid)
			if err != nil {
				handleErr(w, err, "DEL apifileHandler")
				return
			}
			return
		}

		http.Error(w, "Use GET/POST/PUT/DELETE", 401)
	}
}

func apifilesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ff []*File
		var err error

		quserid := idtoi(r.FormValue("userid"))
		qfilename := r.FormValue("filename")
		qlimit := atoi(r.FormValue("limit"))
		qoffset := atoi(r.FormValue("offset"))
		qfiletype := r.FormValue("filetype")

		if qfiletype == "image" {
			ff, err = findImageFiles(db, quserid, qfilename, qlimit, qoffset)
		} else if qfiletype == "attachment" {
			ff, err = findAttachmentFiles(db, quserid, qfilename, qlimit, qoffset)
		} else {
			ff, err = findFiles(db, quserid, qfilename, qlimit, qoffset)
		}
		if err != nil {
			handleErr(w, err, "apifilesHandler")
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		P("%s", jsonstr(ff))
	}
}

// GET /api/site
// POST /api/site {...}
func apisiteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			site := findSite(db)
			if site == nil {
				http.Error(w, "Not found.", 404)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(site))
			return
		} else if r.Method == "POST" || r.Method == "PUT" {
			u := validateApiUser(db, r)
			if u == nil {
				http.Error(w, "Invalid user", 401)
				return
			}
			if u.Userid != 1 {
				http.Error(w, "Not authorized", 401)
				return
			}
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				handleErr(w, err, "POST apisiteHandler")
				return
			}
			var site Site
			err = json.Unmarshal(bs, &site)
			site.Siteid = 1
			if err != nil {
				handleErr(w, err, "POST apisiteHandler")
				return
			}
			err = createSite(db, &site)
			if err != nil {
				handleErr(w, err, "POST apisiteHandler")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(site))
			return
		}

		http.Error(w, "Use GET/PUT/POST", 401)
	}
}

// GET /api/usersettings?id=123
// POST /api/usersettings {...}
func apiusersettingsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			qid := idtoi(r.FormValue("id"))
			if qid == 0 {
				http.Error(w, "Not found.", 404)
				return
			}

			us := findUserSettingsById(db, qid)
			if us == nil {
				http.Error(w, "Not found.", 404)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(us))
			return
		} else if r.Method == "POST" || r.Method == "PUT" {
			u := validateApiUser(db, r)
			if u == nil {
				http.Error(w, "Invalid user", 401)
				return
			}
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				handleErr(w, err, "POST apiusersettingsHandler")
				return
			}
			var us UserSettings
			err = json.Unmarshal(bs, &us)
			if err != nil {
				handleErr(w, err, "POST apiusersettingsHandler")
				return
			}
			us.Userid = u.Userid
			err = createUserSettings(db, &us)
			if err != nil {
				handleErr(w, err, "POST apiusersettingsHandler")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			P := makeFprintf(w)
			P("%s", jsonstr(us))
			return
		}

		http.Error(w, "Use GET/PUT/POST", 401)
	}
}
