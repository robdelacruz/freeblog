{#if ui.loadstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init(entryid)}">Retry</a>
        <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
    </div>
{:else}
    <form class="flex-grow flex flex-col panel mx-auto py-2 px-8 text-sm" on:submit|preventDefault={onsubmit}>
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
import {currentSession} from "./helpers.js";
export let entryid = 0;

let svcurl = "/api";
let session = currentSession();

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

init(entryid);

async function init(qentryid) {
    ui.loadstatus = "loading entry...";

    if (qentryid == 0) {
        ui.loadstatus = "";
        ui.entry = blankentry;
        return;
    }

    let [entry, err] = await findentry(qentryid);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading entry";
        ui.entry = blankentry;
        return;
    }
    ui.loadstatus = "";
    ui.entry = entry;

    let [entryhtml, err2] = await findentryhtml(qentryid);
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

    let err = await delentry(entryid);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error deleting entry";
        return;
    }

    ui.submitstatus = "";
    ui.entryhtml = "";
    dispatch("submit");
}

// Returns [entry, err]
async function findentry(entryid) {
    let sreq = `${svcurl}/entry?id=${entryid}`;
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            if (res.status == 404) {
                return [null, null];
            }
            let s = await res.text();
            return [null, new Error(s)];
        }
        let entry = await res.json();
        return [entry, null];
    } catch(err) {
        return [null, err];
    }
}

// Returns [entryhtml, err]
async function findentryhtml(entryid) {
    let sreq = `${svcurl}/entry?id=${entryid}&fmt=html`;
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            if (res.status == 404) {
                return [null, null];
            }
            let s = await res.text();
            return [null, new Error(s)];
        }
        let entryhtml = await res.text();
        return [entryhtml, null];
    } catch(err) {
        return [null, err];
    }
}

// Returns err
async function delentry(entryid) {
    let sreq = `${svcurl}/entry?id=${entryid}`;

    try {
        let res = await fetch(sreq, {method: "DELETE"});
        if (!res.ok) {
            let s = await res.text();
            console.error(s);
            return new Error(s);
        }
        return null;
    } catch(err) {
        return err;
    }
}
</script>

