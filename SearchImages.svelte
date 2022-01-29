<form class="flex flex-col" autocomplete="off" on:submit|preventDefault="{e => {}}">
    <div class="flex flex-row justify-between mb-2">
        <h1 class="font-bold text-base">Search</h1>
        <div class="text-sm">
            <Tablinks links="images|Images;files|Files" bind:sel={tabsel} on:sel={ontabsel} />
        </div>
    </div>
    <div class="mb-2">
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="search" name="search" type="text" bind:value={qfilter}>
    </div>
{#if status != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{status}</p>
    </div>
{/if}
{#if err != ""}
    <div class="mb-2">
        <p class="font-bold uppercase text-xs">{err}</p>
    </div>
{/if}
{#if tabsel == "images"}
    <div class="flex flex-row flex-wrap mb-2 justify-start">
    {#each display_files as f (f.fileid)}
        <FileThumbnail title={f.title} url={f.url} />
    {/each}
    </div>
{:else if tabsel == "files"}
    <div class="mb-2">
    {#each display_files as f (f.fileid)}
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
let tabsel = "images";
let qfilter = "";
let files = [];
let display_files = [];
let status = "";
let err = "";

$: display_files = filterFiles(files, qfilter);

$: init(tabsel);

async function init(tabsel) {
    status = "loading files...";
    let filetype = "image";
    if (tabsel == "files") {
        filetype = "attachment";
    }
    let [ff, e] = await searchfiles("", filetype);
    if (e != null) {
        console.error(e);
        err = "Server error loading files";
    }
    files = ff;
    status = "";
}

function filterFiles(files, qfilter) {
    let ff = [];
    qfilter = qfilter.trim().toLowerCase();

    for (let i=0; i < files.length; i++) {
        let f = files[i];
        if (f.filename.toLowerCase().includes(qfilter) ||
            f.title.toLowerCase().includes(qfilter) ||
            f.url.toLowerCase().includes(qfilter)) {
            ff.push(f);
        }
    }
    return ff;
}

function ontabsel(e) {
    files = [];
}

async function searchfiles(qsearch, filetype) {
    let sreq = `${svcurl}/files?filetype=${filetype}&userid=${userid}&filename=${qsearch}`;
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

</script>

