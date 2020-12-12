<form class="flex flex-col" bind:this={frm} on:submit={onsubmit}>
    <div class="flex flex-row justify-between">
        <h1 class="font-bold mb-1 text-base">Upload</h1>
        <div class="text-sm">
            <Tablinks links="images|Images;files|Files" bind:sel={ui.tabsel} />
        </div>
    </div>
    <div class="mb-2">
{#if ui.tabsel == "images"}
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="files" name="files" type="file" accept=".jpg, .jpeg, .png, .gif, .bmp, .tif, .tiff" multiple bind:files={files}>
{:else if ui.tabsel == "files"}
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="files" name="files" type="file" multiple bind:files={files}>
{/if}
    </div>
{#if files != null}
    {#if ui.tabsel == "images"}
    <div class="flex flex-row flex-wrap mb-2 justify-start">
        {#each files as previewfile (previewfile.name)}
        <div class="flex flex-col items-center border px-1 pt-1">
            <img class="w-20 h-20" alt="{previewfile.name}" title="{previewfile.name}" use:setimgsrc={previewfile}>
        </div>
        {/each}
    </div>
    {:else if ui.tabsel == "files"}
    <div class="flex flex-col mb-2 justify-start">
        {#each files as previewfile (previewfile.name)}
        <p class="action text-xs">{previewfile.name}</p>
        {/each}
    </div>
    {/if}
{/if}
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
    <div class="flex flex-row justify-center mb-2 justify-center">
            <button class="inline w-full py-1 px-2 border border-gray-500 font-bold">Upload</button>
    </div>
</form>

<script>
import Tablinks from "./Tablinks.svelte";

let svcurl = "/api";
let frm;
let inputfiles;
let files;
let ui = {};
ui.tabsel = "images";
ui.status = "";
ui.err = "";

function setimgsrc(node, previewfile) {
    let fr = new FileReader();
    fr.readAsDataURL(previewfile)
    fr.onloadend = function() {
        node.setAttribute("src", fr.result);
    }
}

async function onsubmit(e) {
    e.preventDefault();
    if (files == null) {
        ui.err = "Please select image(s) to upload.";
        return;
    }

    ui.status = "Processing...";
    ui.err = "";

    let fd = new FormData(frm);
    let err = await submitform(fd);
    if (err != null) {
        console.error(err);
        ui.err = "Server error while uploading images";
        ui.status = "";
        return;
    }

    files = null;
    frm.reset();
    ui.status = "";
}

async function submitform(formdata) {
    let sreq = `${svcurl}/uploadfiles/`;
    try {
        let res = await fetch(sreq, {
            method: "POST",
            body: formdata,
        });
        if (!res.ok) {
            let s = await res.text();
            return new Error(s);
        }
        return null;
    } catch(err) {
        return err;
    }
}
</script>

