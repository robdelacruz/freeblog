{#if ui.loadstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init(id)}">Retry</a>
        <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
    </div>
{:else}
    <form class="flex-grow flex flex-col mx-auto text-sm" on:submit|preventDefault={onsubmit}>
        <div class="flex flex-row py-1">
            <div class="flex-grow">
                <p class="inline mr-1">Deleting Entry:</p>
                <a class="action font-bold text-gray-900" href="/entry?id={ui.entry.entryid}" target="_blank">{ui.entry.title}</a>
            </div>
            <div>
                <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Delete</button>
                <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
            </div>
        </div>
        <div class="mb-2 content">
            {@html ui.entryhtml}
        </div>
        {#if ui.submitstatus != ""}
            <div class="mb-2">
                <p class="uppercase italic text-xs">{ui.submitstatus}</p>
            </div>
        {/if}
    </form>
{/if}

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
import {currentSession, find, del} from "./helpers.js";
export let id = 0;

let svcurl = "/api";

let blankentry = {
    entryid: 0,
    title: "",
    body: "",
    tags: "",
};

let ui = {};

ui.loadstatus = "";
ui.submitstatus = "";
ui.entry = blankentry;
ui.entryhtml = "";

init(id);

async function init(qentryid) {
    ui.loadstatus = "loading entry...";

    if (qentryid == 0) {
        ui.loadstatus = "";
        ui.entry = blankentry;
        return;
    }

    let sreq = `${svcurl}/entry?id=${qentryid}`;
    let [entry, err] = await find(sreq);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading entry";
        ui.entry = blankentry;
        return;
    }
    ui.loadstatus = "";
    ui.entry = entry;

    sreq = `${svcurl}/entry?id=${qentryid}&fmt=html`;
    let [entryhtml, err2] = await find(sreq, "text");
    if (err2 !=  null) {
        console.error(err2);
        ui.loadstatus = "Server error loading entry";
        ui.entryhtml = "";
        return;
    }
    ui.loadstatus = "";
    ui.entryhtml = entryhtml;
}


function oncancel(e) {
    dispatch("cancel");
}

async function onsubmit(e) {
    ui.submitstatus = "processing";

    let sreq = `${svcurl}/entry?id=${id}`;
    let err = await del(sreq);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error deleting entry";
        return;
    }

    ui.submitstatus = "";
    ui.entryhtml = "";
    dispatch("submit");
}
</script>

