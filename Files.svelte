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

export let userid = 0;

let svcurl = "/api";
let ui = {};
ui.files = [];
ui.status = "";

init(userid);

async function init(userid) {
    ui.status = "";

    let sreq = `${svcurl}/files?filetype=attachment&userid=${userid}`;
    // Show all files for admin
    if (userid == 1) {
        sreq = `${svcurl}/files?filetype=attachment`;
    }
    let [ff, err] = await find(sreq);
    if (err != null) {
        console.error(err);
        ui.status = "Server error while fetching files";
    }
    if (ff == null) {
        ff = [];
    }
    ui.files = ff;
}

function dispatchAction(action, itemid) {
    dispatch("action", {
        action: action,
        itemid: itemid,
    });
}

</script>


