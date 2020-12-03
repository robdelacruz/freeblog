<div class="border px-1 py-1 haspopupmenu" bind:this={container}>
    <a href="#a" on:click|preventDefault={onclick}>
        <img class="w-20 h-20" alt={filename} title={filename} src={url}>
    </a>
    {#if ui.showmenu}
    <div class="absolute top-auto py-1 bg-gray-200 text-gray-800 w-20 border border-gray-500 shadow-xs text-xs">
        <a href={url} target="_blank" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem" on:click={hidemenu}>View</a>
        <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem" on:click='{e => copyMarkdownLink(filename, url)}'>Copy Link</a>
    </div>
    {/if}
</div>

<script>
import {onMount, createEventDispatcher} from "svelte";
export let filename = "";
export let url = "";
let container;
let ui = {};
ui.showmenu = false;

onMount(function() {
    // Close any open pop-up menus when globalclick signal received.
    container.addEventListener("globalclick", function(e) {
        ui.showmenu = false;
    });
});

function hidemenu() {
    ui.showmenu = false;
}
function onclick(e) {
    ui.showmenu = !ui.showmenu;
    e.stopPropagation();
}
function copyMarkdownLink(filename, url) {
    //let s = `<img alt="${filename}" src="${url}">`;
    let s = `![${filename}](${url})`;
    navigator.clipboard.writeText(s);

    hidemenu();
}

</script>
