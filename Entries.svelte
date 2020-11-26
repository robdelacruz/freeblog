<div class="panel py-2 px-4 mb-2 h-full">
    <div class="flex flex-row justify-between">
        <h1 class="font-bold mb-1 text-base">Entries</h1>
        <div>
            <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault='{e => switchmode(e, "add", 0)}'>Add Entry</a>
        </div>
    </div>
{#each ui.entries as entry}
    <div class="flex flex-row py-1">
        <div class="flex-grow">
            <a class="action text-sm text-gray-900" href="/entry?id={entry.entryid}">{entry.title}</a>
        </div>
        <div>
            <a class="action text-xs text-gray-700 mr-2" href="#a" on:click|preventDefault='{e => switchmode(e, "edit", entry.entryid)}'>edit</a>
            <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault='{e => switchmode(e, "del", entry.entryid)}'>delete</a>
        </div>
    </div>
{/each}
</div>

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
export let username = "";

let svcurl = "/api";
let ui = {};
ui.entries = [];
ui.status = "";

init(username);

async function init(username) {
    ui.status = "";
    let [ee, err] = await findentries(username);
    if (err != null) {
        console.error(err);
        ui.status = "Server error while fetching entries";
    }
    ui.entries = ee;
}

// Returns []entries, error
async function findentries(username) {
    let sreq = `${svcurl}/entries?username=${username}`;
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

function switchmode(e, mode, entryid) {
    dispatch("mode", {
        mode: mode,
        entryid: entryid,
    });
}

</script>

