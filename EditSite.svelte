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
            <label class="block font-bold uppercase text-xs" for="title">site name</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="title" name="title" type="text" bind:value={ui.site.title}>
        </div>
        <div class="flex flex-row items-center mb-2">
            <input class="mr-2" id="groupblog" name="groupblog" type="checkbox" bind:checked={ui.site.isgroup}>
            <label class="font-bold uppercase text-xs" for="groupblog">group blog</label>
        </div>
        <div class="flex-grow flex flex-col mb-2">
            <label class="block font-bold uppercase text-xs" for="about">about description</label>
            <textarea class="flex-grow block border border-gray-500 py-1 px-4 w-full leading-5" id="about" name="about" bind:value={ui.site.about}></textarea>
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

let svcurl = "/api";
let blanksite = {
    siteid: 1,
    title: "",
    about: "",
    isgroup: false,
};

let ui = {};

ui.loadstatus = "";
ui.submitstatus = "";
ui.site = blanksite;

init();

async function init() {
    ui.loadstatus = "loading site settings...";

    let sreq = `${svcurl}/site`
    let [site, err] = await find(sreq);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading site settings";
        ui.site = blanksite;
        return;
    }

    ui.loadstatus = "";
    ui.site = site;
}

async function onsubmit(e) {
    ui.submitstatus = "processing";

    let sreq = `${svcurl}/site/`;
    let [savedsite, err] = await submit(sreq, "POST", ui.site);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error submitting site settings";
        return;
    }

    ui.submitstatus = "";
    ui.site = savedsite;
    dispatch("submit", savedsite);
}

function oncancel(e) {
    dispatch("cancel");
}
</script>


