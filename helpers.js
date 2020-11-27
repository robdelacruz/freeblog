export function currentSession() {
    let cookies = document.cookie.split(";");
    for (let i=0; i < cookies.length; i++) {
        let cookie = cookies[i].trim();
        let [k,v] = cookie.split("=");
        if (k != "useridtok") {
            continue;
        }
        if (v == undefined) {
            v = "";
        }
        let [suserid, tok] = v.split("|");
        let userid = parseInt(suserid, 10);
        if (tok == undefined) {
            tok = "";
        }
        return {userid: userid, tok: tok};
    }
    return {userid: 0, tok: ""};
}

