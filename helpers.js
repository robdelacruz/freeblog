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

