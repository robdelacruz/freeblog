<div class="flex flex-row">
    <h1 class="flex-grow font-bold text-xl my-2 mr-2">Dashboard</h1>
    <div class="self-center">
        <a href="#a" class="action text-sm self-center rounded px-0 py-0 mr-2">link 1</a>
        <a href="#a" class="action text-sm self-center rounded px-0 py-0">link 2</a>
    </div>
</div>
<div class="flex flex-row flex-wrap justify-start">
    <div class="flex-grow mr-2 main-col min-h-85">
{#if ui.mode == ""}
        <Entries username={session.username} on:mode={entries_mode} />
{:else if ui.mode == "add"}
        <EditEntry on:submit={resetmode} on:cancel={resetmode}/>
{:else if ui.mode == "edit"}
        <EditEntry entryid={ui.entryid} on:submit={resetmode} on:cancel={resetmode}/>
{:else if ui.mode == "del"}
        <DelEntry entryid={ui.entryid} on:submit={resetmode} on:cancel={resetmode}/>
{/if}
    </div>
    <div class="side-col">
        <div class="panel py-2 px-4 text-sm mb-2">
            <h1 class="font-bold mb-1 text-base">{session.username}</h1>
            <div class="flex flex-col">
                <a href="/password" class="">Change Password</a>
                <a href="#a" class="">Delete Account</a>
            </div>
        </div>

        <UploadImages />
        <SearchImages />

    </div>
</div>

<script>
import {currentSession} from "./helpers.js";
import Entries from "./Entries.svelte";
import UploadImages from "./UploadImages.svelte";
import SearchImages from "./SearchImages.svelte";
import EditEntry from "./EditEntry.svelte";
import DelEntry from "./DelEntry.svelte";

let session = currentSession();
let ui = {};
ui.mode = "";
ui.entryid = 0;

function entries_mode(e) {
    ui.mode = e.detail.mode;
    ui.entryid = e.detail.entryid;
}
function resetmode() {
    ui.mode = "";
}

</script>

