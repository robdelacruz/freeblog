{#if ui.loadstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init(entryid)}">Retry</a>
        <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
    </div>
{:else}
    <form class="flex-grow flex flex-col panel mx-auto py-2 px-8 text-sm h-full">
        <div class="flex flex-row justify-between mb-4">
        {#if optdel != null}
            <h1 class="font-bold mb-1 text-base">Delete Entry</h1>
        {:else}
            <h1 class="font-bold mb-1 text-base">View Entry</h1>
        {/if}
            <div>
                <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Back</a>
            </div>
        </div>
        <div class="mb-2">
            {@html ui.entryhtml}
        </div>
        {#if ui.submitstatus != ""}
            <div class="mb-2">
                <p class="uppercase italic text-xs">{ui.submitstatus}</p>
            </div>
        {/if}
        {#if optdel != null}
            <label class="block font-bold uppercase text-xs" for="tags">Warning: Entry will be permanently deleted</label>
            <div class="flex flex-row justify-center mb-2 justify-center">
                <button class="inline w-full py-1 px-2 border border-gray-500 font-bold" on:click|preventDefault={ondel}>Delete Entry</button>
            </div>
        {/if}
    </form>
{/if}

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
import {currentSession} from "./helpers.js";
export let entryid = 0;
export let optdel = null;

let svcurl = "/api";
let session = currentSession();

let ui = {};

ui.loadstatus = "";
ui.submitstatus = "";
ui.entryhtml = "";

init(entryid);

async function init(qentryid) {
    ui.loadstatus = "loading entry...";
    ui.entryhtml = "";

    if (qentryid == 0) {
        ui.loadstatus = "";
        ui.entryhtml = "";
        return;
    }

    let [entryhtml, err] = await findentryhtml(qentryid);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading entry";
        return;
    }

    ui.loadstatus = "";
    ui.entryhtml = entryhtml;
}

function oncancel(e) {
    dispatch("cancel");
}

async function ondel(e) {
    ui.submitstatus = "processing";

    let err = await delentry(entryid);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error deleting entry";
        return;
    }

    ui.submitstatus = "";
    ui.entryhtml = "";
    dispatch("del");
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

