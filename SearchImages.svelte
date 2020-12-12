<form class="flex flex-col" on:submit={onsubmit}>
    <div class="flex flex-row justify-between mb-2">
        <h1 class="font-bold text-base">Search</h1>
        <div class="text-sm">
            <Tablinks links="images|Images;files|Files" />
        </div>
    </div>
    <div class="mb-2">
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
        <FileThumbnail title={f.title} url={f.url} />
{/each}
    </div>
</form>

<script>
import {currentSession} from "./helpers.js";
import Tablinks from "./Tablinks.svelte";
import FileThumbnail from "./FileThumbnail.svelte";

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

