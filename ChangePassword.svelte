<form class="flex-grow flex flex-col mx-auto px-4 text-sm h-85vh" on:submit|preventDefault={onsubmit}>
    <div class="flex flex-row py-1">
        <div class="flex-grow">
            <p class="inline mr-1">Change Password:</p>
        </div>
        <div>
            <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Submit</button>
            <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
        </div>
    </div>
    <div class="mb-2">
        <label class="block font-bold uppercase text-xs" for="pwd">password</label>
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="pwd" name="pwd" type="password" bind:value={ui.pwd}>
    </div>
    <div class="mb-2">
        <label class="block font-bold uppercase text-xs" for="newpwd">new password</label>
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="newpwd" name="newpwd" type="password" bind:value={ui.newpwd}>
    </div>
    <div class="mb-2">
        <label class="block font-bold uppercase text-xs" for="newpwd2">re-enter password</label>
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="newpwd2" name="newpwd2" type="password" bind:value={ui.newpwd2}>
    </div>
{#if ui.submitstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.submitstatus}</p>
    </div>
{/if}
    <div class="flex flex-row justify-center mb-2 justify-center">
            <button class="inline w-full py-1 px-2 border border-gray-500 font-bold">Submit</button>
    </div>
</form>

<script>
import {currentSession, initPopupHandlers} from "./helpers.js";
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();

let svcurl = "/api";
let session = currentSession();
let ui = {};
ui.pwd = "";
ui.newpwd = "";
ui.newpwd2 = "";

ui.submitstatus = "";

async function onsubmit(e) {
    if (ui.newpwd != ui.newpwd2) {
        ui.submitstatus = "passwords don't match";
        return;
    }
    ui.submitstatus = "processing";

    let req = {
        userid: session.userid,
        pwd: ui.pwd,
        newpwd: ui.newpwd,
    };
    let err = await changepwd(req);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error on password change";
        return;
    }

    ui.submitstatus = "";
    dispatch("submit");
}

function oncancel(e) {
    dispatch("cancel");
}

// Returns err
async function changepwd(req) {
    let sreq = `${svcurl}/changepwd/`;
    try {
        let res = await fetch(sreq, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(req),
        });
        if (!res.ok) {
            let s = await res.text();
            console.error(s);
            return new Error(s);
        }
        return null;
    } catch(err) {
        return null;
    }
}
</script>


