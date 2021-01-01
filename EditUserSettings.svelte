{#if ui.loadstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init()}">Retry</a>
        <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
    </div>
{:else}
    <form class="flex-grow flex flex-col mx-auto px-4 text-sm h-80vh" on:submit|preventDefault={onsubmit}>
        <div class="flex flex-row py-1">
            <div class="flex-grow">
                <p class="inline mr-1">Editing:</p>
                <span class="action font-bold text-gray-900">Site Settings</span>
            </div>
            <div>
                <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Submit</button>
                <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
            </div>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="title">blog title</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="title" name="title" type="text" bind:value={ui.us.blogtitle}>
        </div>
        <div class="flex-grow flex flex-col mb-2">
            <label class="block font-bold uppercase text-xs" for="about">about description</label>
            <textarea class="flex-grow block border border-gray-500 py-1 px-4 w-full leading-5" id="about" name="about" bind:value={ui.us.blogabout}></textarea>
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
export let userid = 0;

let svcurl = "/api";
let blankus = {
    userid: 0,
    blogtitle: "",
    blogabout: "",
};

let ui = {};

ui.loadstatus = "";
ui.submitstatus = "";
ui.us = blankus;

init(userid);

async function init(quserid) {
    ui.loadstatus = "loading user settings...";

    let sreq = `${svcurl}/usersettings?id=${quserid}`
    let [us, err] = await find(sreq);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading user settings";
        ui.us = blankus;
        return;
    }

    ui.loadstatus = "";
    ui.us = us;
}

async function onsubmit(e) {
    ui.submitstatus = "processing";

    let sreq = `${svcurl}/usersettings/`;
    let [savedus, err] = await submit(sreq, "POST", ui.us);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error submitting user settings";
        return;
    }

    ui.submitstatus = "";
    ui.us = savedus;
    dispatch("submit", savedus);
}

function oncancel(e) {
    dispatch("cancel");
}
</script>


