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
                <p class="inline mr-1">Deleting File:</p>
                <a class="action font-bold text-gray-900" href="{ui.file.url}" target="_blank">{ui.file.filename}</a>
            </div>
            <div>
                <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Delete</button>
                <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
            </div>
        </div>
        <div class="mb-2">
            <a class="action text-xs" href="{ui.file.url}" target="_blank">{ui.file.filename}</a>
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
import {find, del} from "./helpers.js";
export let id = 0;

let svcurl = "/api";

let files;
let ui = {};
ui.file = null;

ui.loadstatus = "";
ui.submitstatus = "";

init(id);

async function init(qid) {
    ui.loadstatus = "loading file...";

    if (qid == 0) {
        ui.loadstatus = "";
        ui.file = null;
        return;
    }

    let sreq = `${svcurl}/file?id=${qid}`;
    let [f, err] = await find(sreq);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading file";
        ui.file = null;
        return;
    }

    ui.loadstatus = "";
    ui.file = f;
}

async function onsubmit(e) {
    ui.submitstatus = "processing";

    let sreq = `${svcurl}/file?id=${id}`;
    let err = await del(sreq);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error deleting file";
        return;
    }

    ui.submitstatus = "";
    ui.file = null;
    dispatch("submit");
}

function oncancel(e) {
    dispatch("cancel");
}

</script>


