<form class="flex flex-col panel py-2 px-4 text-sm" on:submit={onsubmit}>
    <div class="flex flex-row justify-between">
        <h1 class="font-bold mb-1 text-base">Search Images</h1>
        <div class="flex flex-row justify-start">
            <a class="action self-center rounded text-xs px-1 py-0 mr-2" href="#a">Close</a>
        </div>
    </div>
    <div class="mb-2">
        <label class="block font-bold uppercase text-xs" for="search">search</label>
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="search" name="search" type="text" bind:value={ui.qsearch}>
    </div>
{#if ui.status != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.status}</p>
    </div>
{/if}
{#if ui.err != ""}
    <div class="mb-2">
        <p class="font-bold uppercase text-xs">{ui.err}</p>
    </div>
{/if}
    <div class="flex flex-row flex-wrap mb-2 justify-start">
{#each ui.files as f (f.fileid)}
        <div class="flex flex-col items-center border px-1 pt-1">
            <a href={f.url} target="_blank">
                <img class="w-20 h-20" alt={f.filename} title={f.filename} src={f.url}>
            </a>
        </div>
{/each}
    </div>
</form>

<script>
import {currentSession} from "./helpers.js";

let svcurl = "/api";
let session = currentSession();
let ui = {};
ui.qsearch = "";
ui.files = [];
ui.status = "";
ui.err = "";

async function searchfiles(qsearch) {
    let sreq = `${svcurl}/files?filename=${qsearch}`;
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            let s = await res.text();
            return [[], new Error(s)];
        }
        let files = await res.json();
        return [files, null];
    } catch(err) {
        return [[], err];
    }
}

async function onsubmit(e) {
    e.preventDefault();

    ui.status = "searching...";
    ui.err = "";
    ui.files = [];

    let [ff, err] = await searchfiles(ui.qsearch);
    if (err != null) {
        console.error(err);
        ui.err = "Server error occured during search";
    }
    ui.files = ff;
    ui.status = "";
}

</script>

