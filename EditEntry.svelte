{#if ui.loadstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init(id)}">Retry</a>
        <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
    </div>
{:else}
    <form class="flex-grow flex flex-col mx-auto px-4 text-sm h-80vh" on:submit|preventDefault={onsubmit}>
        <div class="flex flex-row py-1">
            <div class="flex-grow">
            {#if id == 0}
                <p class="inline mr-1">New Entry:</p>
            {:else}
                <p class="inline mr-1">Editing:</p>
            {/if}
                <a class="action font-bold text-gray-900" href="/entry?id={ui.entry.entryid}" target="_blank">{ui.entry.title}</a>
            </div>
            <div>
                <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Submit</button>
                <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
            </div>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="title">title</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="title" name="title" type="text" bind:value={ui.entry.title}>
        </div>
        <div class="flex-grow flex flex-col mb-2">
            <label class="block font-bold uppercase text-xs" for="body">entry</label>
            <textarea class="flex-grow block border border-gray-500 py-1 px-4 w-full leading-5" id="body" name="body" bind:value={ui.entry.body}></textarea>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="tags">tags</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="tags" name="tags" type="text" value="">
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
import {find, submit} from "./helpers.js";
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

init(id);

async function init(qentryid) {
    ui.loadstatus = "loading entry...";

    if (qentryid == 0) {
        ui.loadstatus = "";
        ui.entry = blankentry;
        return;
    }

    let sreq = `${svcurl}/entry?id=${qentryid}`
    let [entry, err] = await find(sreq);
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

    let sreq = `${svcurl}/entry/`;
    let method = "PUT";
    if (ui.entry.entryid == 0) {
        method = "POST";
    }
    let [savedentry, err] = await submit(sreq, method, ui.entry);
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
</script>

