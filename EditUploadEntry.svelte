<script>
let svcurl = "/api";
let container = document.querySelector("#container");

let ui = {};
ui.showupload = false;

let entryform = {};
entryform.status = "";

let qid = getqid();
let blankentry = {
    entryid: 0,
    title: "",
    body: "",
    createdt: "",
    userid: "",
    username: "",
};
let entry = blankentry;
let tags = "";
init(qid);

async function init(qid) {
    if (qid != null) {
        try {
            entry = await findentry(qid);
        } catch(err) {
            console.error(err);
        }
    }
}

function getqid() {
    let p = new URLSearchParams(window.location.search);
    return p.get("id");
}

function onupload(e) {
    ui.showupload = !ui.showupload;

    container.classList.remove("max-w-screen-lg");
    container.classList.remove("max-w-screen-sm");
    if (ui.showupload) {
        container.classList.add("max-w-screen-lg");
    } else {
        container.classList.add("max-w-screen-sm");
    }
}
function onuploadclose(e) {
    container.classList.remove("max-w-screen-lg");
    container.classList.add("max-w-screen-sm");
    ui.showupload = false;
}

async function findentry(entryid) {
    let sreq = `${svcurl}/entry?id=${entryid}`;
    let res = await fetch(sreq, {method: "GET"});
    if (!res.ok) {
        let err = await res.text();
        return Promise.reject(err);
    }
    let e = await res.json();
    return e;
}
</script>

<div class="flex-grow flex flex-row justify-center">
    <form class="flex flex-col panel py-2 px-8 text-sm mr-2" style="width: 640px;">
        <div class="flex flex-row justify-between">
{#if entry.entryid == 0}
            <h1 class="font-bold mb-1 text-base">New Entry</h1>
{:else}
            <h1 class="font-bold mb-1 text-base">Edit Entry</h1>
{/if}
            <div class="flex flex-row justify-start">
                <a class="action self-center rounded text-xs px-1 py-0 mr-2" href="#a" on:click={onupload}>Upload Images</a>
                <a class="action self-center rounded text-xs px-1 py-0 mr-2" href="#a">Preview</a>
            </div>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="title">title</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="title" name="title" type="text" bind:value={entry.title}>
        </div>
        <div class="flex-grow flex flex-col mb-2">
            <label class="block font-bold uppercase text-xs" for="username">entry</label>
            <textarea class="flex-grow block border border-gray-500 py-1 px-4 w-full leading-5" id="entry" name="entry" bind:value={entry.body}></textarea>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="tags">tags</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="tags" name="tags" type="text" bind:value={tags}>
        </div>
{#if entryform.status != ""}
        <div class="mb-2">
            <p class="font-bold uppercase text-xs">{entryform.status}</p>
        </div>
{/if}
        <div class="flex flex-row justify-center mb-2 justify-center">
                <button class="inline w-full py-1 px-2 border border-gray-500 font-bold">Submit</button>
        </div>
        <div class="flex flex-row justify-end">
            <a class="text-xs" href="#a">Cancel</a>
        </div>
    </form>

{#if ui.showupload}
    <form class="flex-grow flex flex-col panel mx-auto py-2 px-4 text-sm w-1/3">
        <div class="flex flex-row justify-between">
            <h1 class="font-bold mb-1 text-base">Upload Images</h1>
            <div class="flex flex-row justify-start">
                <a class="action self-center rounded text-xs px-1 py-0 mr-2" href="#a" on:click={onuploadclose}>Close</a>
            </div>
        </div>
        <div class="mb-2">
            <label class="block font-bold uppercase text-xs" for="images">Select</label>
            <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="images" name="images" type="file" multiple>
        </div>
        <div class="flex flex-row justify-center mb-2 justify-center">
                <button class="inline w-full py-1 px-2 border border-gray-500 font-bold">Upload</button>
        </div>
        <div class="flex flex-row flex-wrap mb-2 justify-start">
            <img class="w-20 h-20 border p-2" src="/static/lky.png" alt="lky">
            <img class="w-20 h-20 border p-2" src="/static/sheep_vote.jpeg" alt="lky">
            <img class="w-20 h-20 border p-2" src="/static/buster2.jpg" alt="lky">
            <img class="w-20 h-20 border p-2" src="/static/lucy_youngcage.jpg" alt="lky">
            <img class="w-20 h-20 border p-2" src="/static/small_lucy_ref.png" alt="lky">
            <img class="w-20 h-20 border p-2" src="/static/floppy-disk_emoji.png" alt="lky">
            <img class="w-20 h-20 border p-2" src="/static/lky_remembered.jpg" alt="lky">
            <img class="w-20 h-20 border p-2" src="/static/wewokeuplikethis.jpg" alt="lky">
            <img class="w-20 h-20 border p-2" src="/static/wednesday_coffee.jpg" alt="lky">
        </div>
    </form>
{/if}
</div>
