<div class="flex flex-row">
    <h1 class="flex-grow font-bold text-xl my-2 mr-2">Dashboard</h1>
    <div class="self-center">
        <a href="#a" class="action text-sm self-center rounded px-0 py-0 mr-2">link 1</a>
        <a href="#a" class="action text-sm self-center rounded px-0 py-0">link 2</a>
    </div>
</div>
<div class="flex flex-row flex-wrap justify-start">
    <div class="flex-grow mr-2 main-col">
        <div class="panel py-2 px-4 mb-2 h-full">
        {#if ui.action == ""}
            <div class="flex flex-row justify-between mb-4 text-sm">
                <Tablinks links="entries|Entries;images|Images;files|Files" on:sel={tablinks_sel} />
                <div>
                {#if ui.tabsel == "entries"}
                    <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click={onaddentry}>Add Entry</a>
                {/if}
                </div>
            </div>
        {/if}

        {#if ui.tabsel == "entries"}
            {#if ui.action == ""}
                <Entries username={session.username} on:action={entries_action} />
            {:else if ui.action == "edit"}
                <EditEntry entryid={ui.entryid} on:submit={clearaction} on:cancel={clearaction}/>
            {:else if ui.action == "del"}
                <DelEntry entryid={ui.entryid} on:submit={clearaction} on:cancel={clearaction}/>
            {/if}
        {/if}
        </div>
    </div>
    <div class="side-col">
        <div class="panel py-2 px-4 text-sm mb-2">
            <SearchImages />
        </div>
        <div class="panel py-2 px-4 text-sm mb-0">
            <UploadImages />
        </div>
    </div>
</div>

<script>
import {currentSession, initPopupHandlers} from "./helpers.js";
import Tablinks from "./Tablinks.svelte";
import Entries from "./Entries.svelte";
import UploadImages from "./UploadImages.svelte";
import SearchImages from "./SearchImages.svelte";
import EditEntry from "./EditEntry.svelte";
import DelEntry from "./DelEntry.svelte";

let session = currentSession();
let ui = {};
ui.tabsel = "entries";
ui.action = "";
ui.entryid = 0;

let tablinks;

initPopupHandlers();

function tablinks_sel(e) {
    ui.tabsel = e.detail;
}
function onaddentry(e) {
    ui.action = "edit";
    ui.entryid = 0;
}
function entries_action(e) {
    ui.action = e.detail.action;
    ui.entryid = e.detail.entryid;
}
function clearaction(e) {
    ui.action = "";
    ui.entryid = 0;
}

</script>

