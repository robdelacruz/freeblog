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
import {find, exec} from "./helpers.js";

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
    let sreq = `${svcurl}/deluser/`;
    let err = await exec(sreq, req);
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

    sreq = `${svcurl}/logout/`;
    await exec(sreq, {});
    window.location.replace("/");
}

function oncancel(e) {
    dispatch("cancel");
}

</script>


