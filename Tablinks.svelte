<div class="tablinks">
{#each ui.links as link}
    {#if link.signal == sel}
        <a href="#a" class="pb-1 border-b border-gray-500 font-bold mr-2" on:click='{e => onlink(link.signal)}'>{link.caption}</a>
    {:else}
        <a href="#a" class="mr-2" on:click='{e => onlink(link.signal)}'>{link.caption}</a>
    {/if}
{/each}
</div>

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();

// Ex. "entries|Entries;images|Images;files|Files"
export let links = "";
export let sel = "";

let ui = {};
ui.links = [];

let ll = links.split(";");
for (let i=0; i < ll.length; i++) {
    let ss = ll[i].split("|");
    if (ss.length != 2) {
        continue;
    }
    let signal = ss[0].trim();
    let caption = ss[1];
    ui.links.push({signal: signal, caption: caption});

    // Make the first link active by default.
    if (sel == "") {
        sel = signal;
    }
}

function onlink(signal) {
    sel = signal;
    dispatch("sel", signal);
}

</script>

