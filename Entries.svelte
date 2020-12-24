{#each ui.entries as entry}
<div class="flex flex-row">
    <div class="flex-grow truncate mr-2">
        <a class="action text-sm text-gray-900" href="/?page=entry&id={entry.entryid}" target="_blank">{entry.title}</a>
    </div>
    <div>
    {#if userid == 1}
        <span class="text-xs text-gray-700 italic mr-4">{entry.username}</span>
    {/if}
        <a class="action text-xs text-gray-700 mr-2" href="#a" on:click|preventDefault='{e => dispatchAction("edit", entry.entryid)}'>edit</a>
        <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault='{e => dispatchAction("del", entry.entryid)}'>delete</a>
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
ui.entries = [];
ui.status = "";

init(userid);

async function init(userid) {
    ui.status = "";

    let sreq = `${svcurl}/entries?userid=${userid}`;
    // Show all entries for admin
    if (userid == 1) {
        sreq = `${svcurl}/entries`;
    }
    let [ee, err] = await find(sreq);
    if (err != null) {
        console.error(err);
        ui.status = "Server error while fetching entries";
    }
    if (ee == null) {
        ee = [];
    }
    ui.entries = ee;
}

function dispatchAction(action, entryid) {
    dispatch("action", {
        action: action,
        itemid: entryid,
    });
}

</script>

