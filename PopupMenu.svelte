<div class="popupmenu" bind:this={popupmenu}>
{#if showmenu}
    <div class="absolute top-auto py-1 bg-gray-200 text-gray-800 w-20 border border-gray-500 shadow-xs text-xs">
    {#each items as item}
        <a href="#a" class="block leading-none px-2 py-1 hover:bg-gray-400 hover:text-gray-900" role="menuitem" on:click='{e => onitem(item.signal)}'>{item.caption}</a>
    {/each}
    </div>
{/if}
</div>

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
let popupmenu = null;

// Ex. "view|View;copy|Copy Link"
export let menu = "";
let items = [];
let mm = menu.split(";");
for (let i=0; i < mm.length; i++) {
    let ss = mm[i].split("|");
    if (ss.length != 2) {
        continue;
    }
    let signal = ss[0].trim();
    let caption = ss[1];
    items.push({signal: signal, caption: caption});
}


onMount(function() {
    // Hide menu when mouse clicked from anywhere on the page.
    popupmenu.addEventListener("globalclick", function(e) {
        hide();
    });
});

function onitem(signal) {
    dispatch(signal); 
}

export let showmenu = false;
export function show() {
    showmenu = true;
}
export function hide() {
    showmenu = false;
}
export function toggle() {
    showmenu = !showmenu;
}
</script>

