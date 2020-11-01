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

	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return fmt.Errorf("Error opening '%s' (%s)\n", dbfile, err)
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/radio.ico") })
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./"))))
	http.HandleFunc("/api/login/", apiloginHandler(db))
	http.HandleFunc("/api/signup/", apisignupHandler(db))
	http.HandleFunc("/api/edituser/", apiedituserHandler(db))
	http.HandleFunc("/api/deluser/", apideluserHandler(db))

	http.HandleFunc("/login/", loginHandler(db))
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

func genTok(u *User) string {
	tok := genHash(fmt.Sprintf("%s_%s", u.Username, u.HashedPwd))
	return tok
}
func validateTok(tok string, u *User) bool {
	return validateHash(tok, fmt.Sprintf("%s_%s", u.Username, u.HashedPwd))
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

type LoginResult struct {
	Tok   string `json:"tok"`
	Error string `json:"error"`
}

func apiloginHandler(db *sql.DB) http.HandlerFunc {
	type LoginReq struct {
		Username string `json:"username"`
		Pwd      string `json:"pwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "apiloginHandler")
			return
		}
		var loginreq LoginReq
		err = json.Unmarshal(bs, &loginreq)
		if err != nil {
			handleErr(w, err, "apiloginHandler")
			return
		}

		var result LoginResult
		tok, err := login(db, loginreq.Username, loginreq.Pwd)
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
		}
		result.Tok = tok

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		bs, _ = json.MarshalIndent(result, "", "\t")
		P("%s\n", string(bs))
	}
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
func apisignupHandler(db *sql.DB) http.HandlerFunc {
	type SignupReq struct {
		Username string `json:"username"`
		Pwd      string `json:"pwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "apisignupHandler")
			return
		}
		var signupreq SignupReq
		err = json.Unmarshal(bs, &signupreq)
		if err != nil {
			handleErr(w, err, "apisignupHandler")
			return
		}
		if signupreq.Username == "" {
			http.Error(w, "username required", 401)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)

		// Attempt to sign up new user.
		var result LoginResult
		err = signup(db, signupreq.Username, signupreq.Pwd)
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
			bs, _ := json.MarshalIndent(result, "", "\t")
			P("%s\n", string(bs))
			return
		}

		// Log in the newly signed up user.
		tok, err := login(db, signupreq.Username, signupreq.Pwd)
		result.Tok = tok
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
		}
		bs, _ = json.MarshalIndent(result, "", "\t")
		P("%s\n", string(bs))
	}
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
func apiedituserHandler(db *sql.DB) http.HandlerFunc {
	type EditUserReq struct {
		Username string `json:"username"`
		Pwd      string `json:"pwd"`
		NewPwd   string `json:"newpwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "apiedituserHandler")
			return
		}
		var req EditUserReq
		err = json.Unmarshal(bs, &req)
		if err != nil {
			handleErr(w, err, "apiedituserHandler")
			return
		}
		if req.Username == "" {
			http.Error(w, "username required", 401)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)

		// Attempt to edit user.
		var result LoginResult
		err = edituser(db, req.Username, req.Pwd, req.NewPwd)
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
			bs, _ := json.MarshalIndent(result, "", "\t")
			P("%s\n", string(bs))
			return
		}

		// Log in the newly edited user.
		tok, err := login(db, req.Username, req.NewPwd)
		result.Tok = tok
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
		}
		bs, _ = json.MarshalIndent(result, "", "\t")
		P("%s\n", string(bs))
	}
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
func apideluserHandler(db *sql.DB) http.HandlerFunc {
	type DelUserReq struct {
		Username string `json:"username"`
		Pwd      string `json:"pwd"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Use POST method", 401)
			return
		}

		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErr(w, err, "apideluserHandler")
			return
		}
		var req DelUserReq
		err = json.Unmarshal(bs, &req)
		if err != nil {
			handleErr(w, err, "apideluserHandler")
			return
		}
		if req.Username == "" {
			http.Error(w, "username required", 401)
			return
		}

		// Attempt to delete user.
		var result LoginResult
		err = deluser(db, req.Username, req.Pwd)
		if err != nil {
			result.Error = fmt.Sprintf("%s", err)
		}

		w.Header().Set("Content-Type", "application/json")
		P := makeFprintf(w)
		bs, _ = json.MarshalIndent(result, "", "\t")
		P("%s\n", string(bs))
	}
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
	P("<body class=\"py-2 bg-white text-black text-base leading-6 myfont light\">\n")
	P("<div id=\"container\" class=\"mx-auto max-w-screen-sm\">\n")
}
func printHtmlClose(P PrintFunc) {
	P("</div>\n")
	P("</body>\n")
	P("</html>\n")
}
func printHeading(P PrintFunc) {
	P("<div class=\"flex flex-row justify-between border-b border-gray-500 pb-1 mb-6 text-sm\"\n>")
	P("    <div>\n")
	P("        <h1 class=\"inline self-end ml-1 mr-2 font-bold\"><a href=\"/\">FreeBlog</a></h1>\n")
	P("        <a href=\"about.html\" class=\"self-end mr-2\">About</a>\n")
	P("        <a href=\"#a\" class=\"bg-gray-400 text-gray-800 self-center rounded px-2 py-1 mr-1\">Add Entry</a>\n")
	P("    </div>\n")
	P("    <div>\n")
	P("        <div class=\"relative inline mr-2\">\n")
	P("            <a class=\"mr-1\" href=\"#a\">\n")
	P("                robdelacruz\n")
	P("            </a>\n")
	P("            <div class=\"hidden absolute top-auto right-0 py-1 bg-gray-200 text-gray-800 w-20 border border-gray-500 shadow-xs w-32\">\n")
	P("                <a href=\"#a\" class=\"block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900\" role=\"menuitem\">Change Password</a>\n")
	P("                <a href=\"#a\" class=\"block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900\" role=\"menuitem\">Delete Account</a>\n")
	P("                <a href=\"#a\" class=\"block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900\" role=\"menuitem\">Reset LocalStorage</a>\n")
	P("            </div>\n")
	P("        </div>\n")
	P("        <a href=\"#a\" class=\"inline self-end mr-1\">Logout</a>\n")
	P("    </div>\n")
	P("</div>\n")
}
func printSampleEntry(P PrintFunc) {
	P("<h1 class=\"font-bold text-2xl mb-4\">The Things We Think and Do Not Say</h1>\n")
	P("<div class=\"content\">\n")
	P(`    <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed purus massa, vulputate quis nibh nec, dictum congue magna. Vivamus in ligula ut massa sollicitudin dictum. Cras nisl ex, dapibus ac ligula et, convallis malesuada eros. Vestibulum porttitor pretium dolor a porta. Integer scelerisque maximus ex at imperdiet. Vestibulum blandit mollis porta. Sed gravida, metus eu lobortis rhoncus, justo nibh rutrum diam, eget auctor arcu nunc fermentum lorem. Sed vel urna sed dolor imperdiet sagittis eget eu leo. Vivamus facilisis ipsum quis cursus feugiat. Suspendisse potenti. Mauris pellentesque mauris at pretium posuere. Quisque accumsan condimentum purus, sed gravida eros rutrum non. Sed feugiat mauris tellus, a sollicitudin sapien gravida sed. Cras mollis suscipit ante et dapibus.

        <p>Morbi mollis, quam vitae ornare fermentum, turpis tellus feugiat dolor, ac auctor lorem orci at velit. Integer at facilisis dui. Praesent in lorem vel nulla dictum convallis. Sed cursus posuere leo, quis iaculis nibh vulputate faucibus. Nulla mollis aliquet dictum. Suspendisse libero tortor, tincidunt eu massa ut, vestibulum suscipit velit. Vivamus vel ornare est. Quisque mollis nec dolor ut sodales. Nunc dolor turpis, finibus sit amet dignissim quis, iaculis mollis augue. Donec sollicitudin nibh a viverra lobortis. Vivamus sed venenatis massa. Praesent in ligula nec nisi placerat fermentum elementum vitae velit. Nam efficitur neque tellus, quis accumsan nisl gravida iaculis. Vestibulum maximus ut est sit amet consequat. Praesent bibendum, massa vel viverra malesuada, tortor turpis ultricies odio, eu vestibulum ante felis facilisis magna.
`)
	P("</div>\n")
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", []string{"/static/bundle.js", "/static/login.js"})
		printHeading(P)

		/*
			P("<form class=\"mx-auto py-4 px-8 max-w-sm bg-gray-200 text-gray-800\">\n")
			P("    <h1 class=\"font-bold mx-auto text-2xl mb-4 text-center\">Sign In</h1>\n")
			P("    <div class=\"mb-2\">\n")
			P("        <label class=\"block font-bold uppercase text-sm\" for=\"username\">username</label>\n")
			P("        <input class=\"block border border-gray-500 py-1 px-4 w-full\" id=\"username\" name=\"username\" type=\"text\" value=\"robdelacruz\">\n")
			P("    </div>\n")
			P("    <div class=\"mb-4\">\n")
			P("        <label class=\"block font-bold uppercase text-sm\" for=\"pwd\">password</label>\n")
			P("        <input class=\"block border border-gray-500 py-1 px-4 w-full\" id=\"pwd\" name=\"pwd\" type=\"password\" value=\"password\">\n")
			P("    </div>\n")
			P("    <div class=\"mb-2\">\n")
			P("        <p class=\"font-bold uppercase text-xs\">Incorrect username or password</p>\n")
			P("    </div>\n")
			P("    <div class=\"mb-4\">\n")
			P("        <button class=\"inline w-full mx-auto py-1 px-2 border border-gray-500 bg-gray-400 font-bold mr-2\">Login</button>\n")
			P("    </div>\n")
			P("    <div class=\"flex flex-row justify-between\">\n")
			P("        <a class=\"underline text-sm\" href=\"#a\">Create New Account</a>\n")
			P("        <a class=\"underline text-sm\" href=\"#a\">Cancel</a>\n")
			P("    </div>\n")
			P("</form>\n")
		*/

		printHtmlClose(P)
	}
}

func entryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		P := makeFprintf(w)
		printHtmlOpen(P, "FreeBlog", nil)
		printHeading(P)
		printSampleEntry(P)
		printHtmlClose(P)
	}
}
