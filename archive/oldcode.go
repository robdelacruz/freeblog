package old

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
				u, sig, err := login(db, u.Username, f.newpwd)
				if err != nil {
					errmsg = fmt.Sprintf("%s", err)
					break
				}
				setLoginCookie(w, u, sig)

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

func accountHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	u, _ := validateLoginCookie(db, r)
	if u == nil {
		http.Error(w, "Must be logged in", 401)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	P := makeFprintf(w)
	pp := getPageParams(r, db)
	printHtmlOpen(P, pp.BlogTitle, nil)
	printContainerOpen(P)
	printHeading(P, u, pp)

	printDivSmallOpen(P, "Account Settings")
	printDivFlex(P, "justify-start")
	P("<div class=\"px-0\">\n")
	P("    <a href=\"/?page=password\" class=\"action block border-b\">Change Password</a>\n")
	if u.Userid != 1 {
		P("    <a href=\"/?page=delaccount\" class=\"action block border-b\">Delete Account</a>\n")
	}
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

func passwordHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
			err := edituser(db, u.Userid, f.pwd, f.newpwd)
			if err != nil {
				errmsg = fmt.Sprintf("%s", err)
				break
			}
			u, sig, err := loginUserid(db, u.Userid, f.newpwd)
			if err != nil {
				errmsg = fmt.Sprintf("%s", err)
				break
			}
			setLoginCookie(w, u, sig)

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	P := makeFprintf(w)
	pp := getPageParams(r, db)
	printHtmlOpen(P, pp.BlogTitle, nil)
	printContainerOpen(P)
	printHeading(P, u, pp)

	printFormSmallOpen(P, "/?page=password", "Change Password")
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

func delaccountHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	u, _ := validateLoginCookie(db, r)
	if u == nil {
		http.Error(w, "Must be logged in", 401)
		return
	}

	var errmsg string
	var f struct{ pwd string }

	if r.Method == "POST" {
		f.pwd = r.FormValue("pwd")
		for {
			// Admin takes ownership of deleted user's entries and files.
			err := transferUserEntries(db, u.Userid, 1)
			if err != nil {
				errmsg = fmt.Sprintf("%s", err)
				break
			}
			err = transferUserFiles(db, u.Userid, 1)
			if err != nil {
				errmsg = fmt.Sprintf("%s", err)
				break
			}

			err = deluser(db, u.Userid, f.pwd)
			if err != nil {
				errmsg = fmt.Sprintf("%s", err)
				break
			}

			// logout
			delLoginCookie(w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	P := makeFprintf(w)
	pp := getPageParams(r, db)
	printHtmlOpen(P, pp.BlogTitle, nil)
	printContainerOpen(P)
	printHeading(P, u, pp)

	//$$ Add a checkbox option to delete user's entries?
	printFormSmallOpen(P, "/?page=delaccount", "Delete Account")
	printFormInputPassword(P, "pwd", "password", f.pwd)
	printFormError(P, errmsg)
	printFormSubmit(P, "Submit")
	printFormLinks(P, "justify-end", "/", "Cancel")
	printFormClose(P)

	printContainerClose(P)
	printHtmlClose(P)
}
