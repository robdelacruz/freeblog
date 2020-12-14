{#each ui.entries as entry}
<div class="flex flex-row">
    <div class="flex-grow">
        <a class="action text-sm text-gray-900" href="/?page=entry&id={entry.entryid}" target="_blank">{entry.title}</a>
    </div>
    <div>
        <a class="action text-xs text-gray-700 mr-2" href="#a" on:click|preventDefault='{e => dispatchAction("edit", entry.entryid)}'>edit</a>
        <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault='{e => dispatchAction("del", entry.entryid)}'>delete</a>
    </div>
</div>
{/each}

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();

export let userid = 0;

let svcurl = "/api";
let ui = {};
ui.entries = [];
ui.status = "";

init(userid);

async function init(userid) {
    ui.status = "";
    let [ee, err] = await findentries(userid);
    if (err != null) {
        console.error(err);
        ui.status = "Server error while fetching entries";
    }
    ui.entries = ee;
}

// Returns []entries, error
async function findentries(userid) {
    let sreq = `${svcurl}/entries?userid=${userid}`;
    // Show all entries for admin
    if (userid == 1) {
        sreq = `${svcurl}/entries`;
    }
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            let s = await res.text();
            return [[], new Error(s)];
        }
        let entries = await res.json();
        return [entries, null];
    } catch (err) {
        return [[], err];
    }
}

function dispatchAction(action, entryid) {
    dispatch("action", {
        action: action,
        itemid: entryid,
    });
}

</script>

