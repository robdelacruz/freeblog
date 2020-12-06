<div class="border px-1 py-1 haspopupmenu">
    <a href="#a" on:click|preventDefault={onclick}>
        <img class="w-20 h-20" alt={filename} title={filename} src={url}>
    </a>
    <PopupMenu bind:this={popupmenu} menu="view|View;copy|Copy Link" on:view={onview} on:copy={oncopy} />
</div>

<script>
import {onMount, createEventDispatcher} from "svelte";
import PopupMenu from "./PopupMenu.svelte";
export let filename = "";
export let url = "";
let popupmenu;
let ui = {};

function onclick(e) {
    popupmenu.toggle();
    e.stopPropagation();
}
function onview(e) {
    window.open(url, "_blank");
}
function oncopy(e) {
    // copy markdown link
    //let s = `<img alt="${filename}" src="${url}">`;
    let s = `![${filename}](${url})`;
    navigator.clipboard.writeText(s);
}
</script>
