{#if ui.loadstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init(id)}">Retry</a>
        <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
    </div>
{:else}
    <form class="flex-grow flex flex-col mx-auto px-4 text-sm h-85vh" on:submit|preventDefault={onsubmit}>
        <div class="flex flex-row py-1">
            <div class="flex-grow">
            {#if id == 0}
                <p class="inline mr-1">New Image:</p>
            {:else}
                <p class="inline mr-1">Editing:</p>
            {/if}
            </div>
            <div>
                <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Submit</button>
                <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
            </div>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="title">title</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="title" name="title" type="text">
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="file">replace file</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="file" name="file" type="file" bind:files={files}>
        </div>
    {#if files != null}
        <div class="mb-2">
        {#each files as previewfile (previewfile.name)}
            <img class="max-w-full" alt="{previewfile.name}" title="{previewfile.name}" use:setimgsrc={previewfile}>
        {/each}
        </div>
    {/if}
    {#if ui.submitstatus != ""}
        <div class="mb-2">
            <p class="uppercase italic text-xs">{ui.submitstatus}</p>
        </div>
    {/if}
        <div class="flex flex-row justify-center mb-2 justify-center">
                <button class="inline w-full py-1 px-2 border border-gray-500 font-bold">Submit</button>
        </div>
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

ui.loadstatus = "";
ui.submitstatus = "";

/*
init(id);

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
}

async function onsubmit(e) {
    ui.submitstatus = "processing";

    let [savedentry, err] = await submitentry(ui.entry);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error submitting entry";
        return;
    }

    ui.submitstatus = "";
    ui.entry = savedentry;
    dispatch("submit", savedentry);
}

function oncancel(e) {
    dispatch("cancel");
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

// Returns [savedentry, err]
async function submitentry(e) {
    let sreq = `${svcurl}/entry/`;
    let method = "";
    if (e.entryid == 0) {
        method = "POST";
    } else {
        method = "PUT";
    }

    try {
        let res = await fetch(sreq, {
            method: method,
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(e),
        });
        if (!res.ok) {
            let s = await res.text();
            console.error(s);
            return [null, new Error(s)];
        }
        let savedentry = await res.json();
        return [savedentry, null];
    } catch(err) {
        return [null, err];
    }
}
*/

function onsubmit(e) {
}
function oncancel(e) {
}
function init(id) {
}

function setimgsrc(node, previewfile) {
    let fr = new FileReader();
    fr.readAsDataURL(previewfile)
    fr.onloadend = function() {
        node.setAttribute("src", fr.result);
    }
}
</script>

