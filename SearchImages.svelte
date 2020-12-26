<form class="flex flex-col" on:submit={onsubmit}>
    <div class="flex flex-row justify-between mb-2">
        <h1 class="font-bold text-base">Search</h1>
        <div class="text-sm">
            <Tablinks links="images|Images;files|Files" bind:sel={ui.tabsel} on:sel={ontabsel} />
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
{#if ui.tabsel == "images"}
    <div class="flex flex-row flex-wrap mb-2 justify-start">
    {#each ui.files as f (f.fileid)}
        <FileThumbnail title={f.title} url={f.url} />
    {/each}
    </div>
{:else if ui.tabsel == "files"}
    <div class="mb-2">
    {#each ui.files as f (f.fileid)}
        <FileLink filename={f.filename} title={f.title} url={f.url} />
    {/each}
    </div>
{/if}
</form>

<script>
import Tablinks from "./Tablinks.svelte";
import FileThumbnail from "./FileThumbnail.svelte";
import FileLink from "./FileLink.svelte";
import {currentSession} from "./helpers.js";
let session = currentSession();

export let userid = 0;
if (userid == 0) {
    userid = session.userid;
}

let svcurl = "/api";
let ui = {};
ui.tabsel = "images";
ui.qsearch = "";
ui.files = [];
ui.status = "";
ui.err = "";

function ontabsel(e) {
    ui.files = [];
}

async function searchfiles(qsearch) {
    let sreq = `${svcurl}/files?filetype=image&userid=${userid}&filename=${qsearch}`;
    if (ui.tabsel == "files") {
        sreq = `${svcurl}/files?filetype=attachment&userid=${userid}&filename=${qsearch}`;
    }
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

