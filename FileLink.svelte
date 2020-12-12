<div class="text-xs haspopupmenu">
    <a class="block" href="#a" title={title} on:click|preventDefault={onclick}>{filename}</a>
    <PopupMenu bind:this={popupmenu} menu="view|View;copy|Copy Link" on:view={onview} on:copy={oncopy} />
</div>

<script>
import {onMount, createEventDispatcher} from "svelte";
import PopupMenu from "./PopupMenu.svelte";
export let filename = "";
export let title = "";
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
    //let s = `<a href="${url}" title="${title}">${filename}</a>`;
    let s = `[${title}](${url})`;
    navigator.clipboard.writeText(s);
}
</script>

