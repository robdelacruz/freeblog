export function currentSession() {
    let cookies = document.cookie.split(";");
    for (let i=0; i < cookies.length; i++) {
        let cookie = cookies[i].trim();
        let [k,v] = cookie.split("=");
        if (k != "usernametok") {
            continue;
        }
        if (v == undefined) {
            v = "";
        }
        let [username, tok] = v.split("|");
        if (tok == undefined) {
            tok = "";
        }
        return {username: username, tok: tok};
    }
    return {username: "", tok: ""};
}

