function readCookie(name) {
    let cookies = document.cookie.split(";");
    for (let i=0; i < cookies.length; i++) {
        let cookie = cookies[i].trim();
        let [k,v] = cookie.split("=");
        if (k == name) {
            if (v == undefined) {
                v = "";
            }
            return v;
        }
    }
    return "";
}

export function currentSession() {
    let suserid = readCookie("userid");
    if (suserid == "") {
        return {userid: 0, username: "", sig: ""};
    }
    let username = readCookie("username");
    let sig = readCookie("sig");

    let userid = parseInt(suserid, 10);
    return {
        userid: userid,
        username: username,
        sig: sig,
    };
}

export function initPopupHandlers() {
    function onglobalclick(e) {
        // Send signal to close any open pop-up menus.
        let mm = document.querySelectorAll(".popupmenu");
        for (let i=0; i < mm.length; i++) {
            let e = new Event("globalclick");
            mm[i].dispatchEvent(e);
        }
    }
    document.addEventListener("click", onglobalclick, false);
}

// All purpose GET request function
// sreq contains the uri with query params,
// Ex. "/api/entry?id=123" or "/api/entries?userid=123"
// fmt contains "json" or undefined/null to return json object
// any other value to fmt will return plaintext object representation
// Returns [null, err] if an error occured.
// Returns [item, null] if successful, where item contains value returned from request.
export async function find(sreq, fmt) {
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            if (res.status == 404) {
                return [null, null];
            }
            let s = await res.text();
            let err = new Error(s);
            err.status = res.status;
            return [null, err];
        }
        let v;
        if (!fmt || fmt == "json") {
            v = await res.json();
        } else {
            v = await res.text();
        }
        return [v, null];
    } catch(err) {
        return [null, err];
    }
}

// All purpose POST/PUT request function
// sreq contains the uri
// method contains the http method ("POST" or "PUT")
// item contains the object to be submitted
// Returns [null, err] if an error occured.
// Returns [item, null] if successful, where item contains final object saved.
export async function submit(sreq, method, item) {
    try {
        let res = await fetch(sreq, {
            method: method,
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(item),
        });
        if (!res.ok) {
            let s = await res.text();
            let err = new Error(s);
            err.status = res.status;
            return [null, err];
        }
        let saveditem = await res.json();
        return [saveditem, null];
    } catch(err) {
        return [null, err];
    }
}

// Similar to submit(), but ignore any response object and always use POST.
// sreq contains the uri
// Returns err if an error occured, or null if successful.
export async function exec(sreq, item) {
    try {
        let res = await fetch(sreq, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(item),
        });
        if (!res.ok) {
            let s = await res.text();
            let err = new Error(s);
            err.status = res.status;
            return err;
        }
        return null;
    } catch(err) {
        return err;
    }
}

// All purpose DELETE request function
// sreq contains the uri
// Returns err if an error occured, or null if successful.
export async function del(sreq) {
    try {
        let res = await fetch(sreq, {method: "DELETE"});
        if (!res.ok) {
            let s = await res.text();
            let err = new Error(s);
            err.status = res.status;
            return err;
        }
        return null;
    } catch(err) {
        return err;
    }
}

