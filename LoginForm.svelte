<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
let svcurl = "/api";

export let username = "";
export let pwd = "";
let frm = {};
frm.mode = "";
frm.status = "";

// Post login username/pwd and async returns loginresult:
// {tok: "", error: ""}
async function login(username, pwd) {
    let sreq = `${svcurl}/login/`;
    let reqbody = {
        username: username,
        pwd: pwd,
    };
    try {
        let res = await fetch(sreq, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(reqbody),
        });
        if (!res.ok) {
            let err = await res.text();
            return {tok: "", error: err};
        }
        let result = await res.json();
        return result;
    } catch(err) {
        console.error(err);
        return {tok: "", error: "server/network error"};
    }
}

async function onlogin(e) {
    e.preventDefault();
    frm.mode = "loading";
    frm.status = "Loging in...";

    let result = await login(username, pwd);
    frm.mode = "";
    frm.status = "";
    if (result.error != "") {
        frm.status = result.error;
        return;
    }
    dispatch("login", {username: username, tok: result.tok});
}
function oncancel(e) {
    e.preventDefault();
    username = "";
    pwd = "";
    dispatch("cancel");
}
function oncreatenewaccount(e) {
    e.preventDefault();
    dispatch("createaccount");
}

function ontestlogin(e) {
    e.preventDefault();
    window.location.assign("/entry");
}
</script>

<form class="mx-auto py-4 px-8 max-w-sm">
    <h1 class="font-bold mx-auto text-2xl mb-4 text-center">Sign In</h1>
    <div class="mb-2">
        <label class="block font-bold uppercase text-sm" for="username">username</label>
        <input class="block border border-gray-500 py-1 px-4 w-full" id="username" name="username" type="text" value="robdelacruz">
    </div>
    <div class="mb-4">
        <label class="block font-bold uppercase text-sm" for="pwd">password</label>
        <input class="block border border-gray-500 py-1 px-4 w-full" id="pwd" name="pwd" type="password" value="password">
    </div>
    <div class="mb-2">
        <p class="font-bold uppercase text-xs">Incorrect username or password</p>
    </div>
    <div class="mb-4">
        <button class="inline w-full mx-auto py-1 px-2 border border-gray-500 font-bold mr-2" on:click={ontestlogin}>Login</button>
    </div>
    <div class="flex flex-row justify-between">
        <a class="underline text-sm" href="#a">Create New Account</a>
        <a class="underline text-sm" href="#a">Cancel</a>
    </div>
</form>

