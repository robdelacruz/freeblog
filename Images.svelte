{#each ui.files as file}
<div class="flex flex-row">
    <div class="flex-grow truncate mr-2">
        <a class="action text-sm text-gray-900" href="/?page=file&id={file.fileid}" target="_blank">{file.filename}</a>
    </div>
    <div>
    {#if userid == 1}
        <span class="text-xs text-gray-700 italic mr-4">{file.username}</span>
    {/if}
        <a class="action text-xs text-gray-700 mr-2" href="#a" on:click|preventDefault='{e => dispatchAction("edit", file.fileid)}'>edit</a>
        <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault='{e => dispatchAction("del", file.fileid)}'>delete</a>
    </div>
</div>
{/each}

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
import {find, submit} from "./helpers.js";
import {currentSession} from "./helpers.js";
let session = currentSession();

export let userid = 0;
if (userid == 0) {
    userid = session.userid;
}

let svcurl = "/api";
let ui = {};
ui.files = [];
ui.status = "";

init(userid);

async function init(userid) {
    ui.status = "";

    let sreq = `${svcurl}/files?filetype=image&userid=${userid}`;
    // Show all images for admin
    if (userid == 1) {
        sreq = `${svcurl}/files?filetype=image`;
    }
    let [ff, err] = await find(sreq);
    if (err != null) {
        console.error(err);
        ui.status = "Server error while fetching images";
    }
    if (ff == null) {
        ff = [];
    }
    ui.files = ff;
}

// Returns []files, error
async function findimages(userid) {
    let sreq = `${svcurl}/files?filetype=image&userid=${userid}`;
    // Show all images for admin
    if (userid == 1) {
        sreq = `${svcurl}/files?filetype=image`;
    }
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            let s = await res.text();
            return [[], new Error(s)];
        }
        let files = await res.json();
        return [files, null];
    } catch (err) {
        return [[], err];
    }
}

function dispatchAction(action, itemid) {
    dispatch("action", {
        action: action,
        itemid: itemid,
    });
}

</script>

