<form class="flex-grow flex flex-col mx-auto px-4 text-sm" on:submit|preventDefault={onsubmit}>
    <div class="flex flex-row py-1">
        <div class="flex-grow">
            <p class="inline mr-1">Delete Account:</p>
            <span class="action font-bold text-gray-900">{session.username}</span>
        </div>
        <div>
            <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Delete Account</button>
            <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
        </div>
    </div>
    <div class="mb-2">
        <label class="block font-bold uppercase text-xs" for="pwd">password</label>
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="pwd" name="pwd" type="password" bind:value={ui.pwd}>
    </div>
{#if ui.submitstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.submitstatus}</p>
    </div>
{/if}
</form>

<script>
import {currentSession, initPopupHandlers} from "./helpers.js";
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();

let svcurl = "/api/deluser";
let session = currentSession();
let ui = {};
ui.pwd = "";
ui.newpwd = "";
ui.newpwd2 = "";

ui.submitstatus = "";

async function onsubmit(e) {
    ui.submitstatus = "processing";

    let req = {
        userid: session.userid,
        pwd: ui.pwd,
    };
    let err = await deluser(req);
    if (err != null) {
        console.error(err);
        if (err.status == 401) {
            ui.submitstatus = err.message;
            return;
        }
        ui.submitstatus = "server error while deleting account";
        return;
    }

    ui.submitstatus = "";
    dispatch("submit");

    await logout();
    window.location.replace("/");
}

function oncancel(e) {
    dispatch("cancel");
}

// Returns err
async function deluser(req) {
    let sreq = `${svcurl}/deluser/`;
    try {
        let res = await fetch(sreq, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(req),
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
// Returns err
async function logout() {
    let sreq = `${svcurl}/logout/`;
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            let s = await res.text();
            let err = new Error(s);
            err.status = res.status;
            return err;
        }
        return null;
    } catch(err) {
        return null;
    }
}
</script>


