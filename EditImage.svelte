{#if ui.loadstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init(id)}">Retry</a>
        <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
    </div>
{:else}
    <form class="flex-grow flex flex-col mx-auto px-4 text-sm" on:submit|preventDefault={onsubmit}>
        <div class="flex flex-row py-1">
            <div class="flex-grow">
            {#if id == 0}
                <p class="inline mr-1">New Image:</p>
            {:else}
                <p class="inline mr-1">Editing:</p>
            {/if}
            <a class="action font-bold text-gray-900" href="{ui.file.url}" target="_blank">{ui.file.filename}</a>
            </div>
            <div>
                <button class="inline py-1 px-4 border border-gray-500 font-bold mr-2">Submit</button>
                <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault={oncancel}>Cancel</a>
            </div>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="filename">filename</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="filename" name="filename" type="text" bind:value={ui.file.filename}>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="title">title</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="title" name="title" type="text" bind:value={ui.file.title}>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="file">replace file</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="file" name="file" type="file" accept=".jpg, .jpeg, .png, .gif, .bmp, .tif, .tiff" bind:files={files}>
        </div>
        <div class="mb-2">
    {#if files != null}
        {#each files as previewfile (previewfile.name)}
            <img class="max-w-full" alt="{previewfile.name}" title="{previewfile.name}" use:setimgsrc={previewfile}>
        {/each}
    {:else}
            <img class="max-w-full" alt="{ui.file.title}" title="{ui.file.title}" src="{`${ui.file.url}&${randnum}`}">
    {/if}
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
let blankfile = {
    fileid: 0,
    filename: "",
    title: "",
    bytes: [],
};

let files;
let ui = {};
ui.file = blankfile;
ui.loadstatus = "";
ui.submitstatus = "";

let randnum = performance.now();

init(id);

async function init(qid) {
    ui.loadstatus = "loading image...";

    if (qid == 0) {
        ui.loadstatus = "";
        ui.file = blankfile;
        return;
    }

    let sreq = `${svcurl}/file?id=${qid}`;
    let [f, err] = await find(sreq);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading image";
        ui.file = blankfile;
        return;
    }

    ui.loadstatus = "";
    ui.file = f;
}

function setimgsrc(node, previewfile) {
    // title is filename without the extension
    ui.file.filename = previewfile.name;
    ui.file.title = previewfile.name.replace(/\.[^.]*$/, "");

    let fr = new FileReader();
    fr.readAsDataURL(previewfile)
    fr.onloadend = function() {
        node.setAttribute("src", fr.result);
        //let s = fr.result.replace(/^data:image\/(.*);base64,/, "");
        let s = fr.result.replace(/^data:.*;base64,/, "");
        ui.file.bytes = s;
    }
}

async function onsubmit(e) {
    ui.submitstatus = "processing";

    let sreq = `${svcurl}/file?id=${id}`;
    let method = "PUT";
    if (ui.file.fileid == 0) {
        method = "POST";
    }
    let [savedfile, err] = await submit(sreq, method, ui.file);
    if (err != null) {
        console.error(err);
        ui.submitstatus = "server error submitting image";
        return;
    }

    ui.submitstatus = "";
    ui.file = savedfile;
    dispatch("submit", savedfile);
}

function oncancel(e) {
    dispatch("cancel");
}

// Returns [savedfile, err]
async function submitfile(f) {
    let sreq = `${svcurl}/file/`;
    let method = "";
    if (f.fileid == 0) {
        method = "POST";
    } else {
        method = "PUT";
    }
    try {
        let res = await fetch(sreq, {
            method: method,
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(f),
        });
        if (!res.ok) {
            let s = await res.text();
            console.error(s);
            return [null, new Error(s)];
        }
        let savedfile = await res.json();
        return [savedfile, null];
    } catch(err) {
        return [null, err];
    }
}

</script>

