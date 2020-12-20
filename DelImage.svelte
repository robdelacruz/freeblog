{#if ui.loadstatus != "" || ui.file == null}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init(id)}">Retry</a>
        <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
    </div>
{:else}
    <form class="flex-grow flex flex-col mx-auto px-4 text-sm" on:submit|preventDefault={onsubmit}>
        <div class="flex flex-row py-1 mb-2">
            <div class="flex-grow">
                <p class="inline mr-1">Deleting Image:</p>
                <a class="action font-bold text-gray-900" href="{ui.file.url}" target="_blank">{ui.file.filename}</a>
            </div>
            <div>
                <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Delete</button>
                <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
            </div>
        </div>
        <div class="mb-2">
            <img class="max-w-full" alt="{ui.file.title}" title="{ui.file.title}" src="{ui.file.url}">
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
export let id = 0;

let svcurl = "/api";

let files;
let ui = {};
ui.file = null;

ui.loadstatus = "";
ui.submitstatus = "";

init(id);

async function init(qid) {
    ui.loadstatus = "loading image...";

    if (qid == 0) {
        ui.loadstatus = "";
        ui.file = null;
        return;
    }

    let [f, err] = await findfile(qid);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading image";
        ui.file = null;
        return;
    }

    ui.loadstatus = "";
    ui.file = f;
}

async function onsubmit(e) {
    ui.submitstatus = "processing";

    let err = await delfile(id);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error deleting image";
        return;
    }

    ui.submitstatus = "";
    ui.file = null;
    dispatch("submit");
}

function oncancel(e) {
    dispatch("cancel");
}

// Returns [file, err]
async function findfile(fileid) {
    let sreq = `${svcurl}/file?id=${fileid}`;
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            if (res.status == 404) {
                return [null, null];
            }
            let s = await res.text();
            return [null, new Error(s)];
        }
        let file = await res.json();
        return [file, null];
    } catch(err) {
        return [null, err];
    }
}

// Returns err
async function delfile(fileid) {
    let sreq = `${svcurl}/file?id=${fileid}`;
    try {
        let res = await fetch(sreq, {method: "DELETE"});
        if (!res.ok) {
            let s = await res.text();
            console.error(s);
            return [null, new Error(s)];
        }
        return null;
    } catch(err) {
        return err;
    }
}

</script>

