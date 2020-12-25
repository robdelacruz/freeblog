<div class="flex flex-row">
    <h1 class="flex-grow font-bold text-xl my-2 mr-2">Dashboard</h1>
    <div class="self-center">
        <a href="/?page=account" class="action text-sm self-center rounded px-0 py-0 mr-2">Account</a>
    </div>
</div>
<div class="flex flex-row flex-wrap justify-start">
    <div class="flex-grow mr-2 main-col">
        <div class="panel py-2 px-4 mb-2 h-full">
        {#if ui.action == ""}
            <div class="flex flex-row justify-between mb-4 text-sm">
                <Tablinks links="entries|Entries;images|Images;files|Files;account|Account" bind:sel={ui.tabsel} />
                <div>
                {#if ui.tabsel == "entries"}
                    <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click={onadditem}>Add Entry</a>
                {:else if ui.tabsel == "images"}
                    <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click={onadditem}>Add Image</a>
                {:else if ui.tabsel == "file"}
                    <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click={onadditem}>Add File</a>
                {/if}
                </div>
            </div>
        {/if}

        {#if ui.tabsel == "entries"}
            {#if ui.action == ""}
                <Entries userid={session.userid} on:action={item_action} />
            {:else if ui.action == "edit"}
                <EditEntry id={ui.itemid} on:submit={clearaction} on:cancel={clearaction}/>
            {:else if ui.action == "del"}
                <DelEntry id={ui.itemid} on:submit={clearaction} on:cancel={clearaction}/>
            {/if}
        {:else if ui.tabsel == "images"}
            {#if ui.action == ""}
                <Images userid={session.userid} on:action={item_action} />
            {:else if ui.action == "edit"}
                <EditImage id={ui.itemid} on:submit={clearaction} on:cancel={clearaction}/>
            {:else if ui.action == "del"}
                <DelImage id={ui.itemid} on:submit={clearaction} on:cancel={clearaction}/>
            {/if}
        {:else if ui.tabsel == "files"}
            {#if ui.action == ""}
                <Files userid={session.userid} on:action={item_action} />
            {:else if ui.action == "edit"}
                <EditFile id={ui.itemid} on:submit={clearaction} on:cancel={clearaction}/>
            {:else if ui.action == "del"}
                <DelFile id={ui.itemid} on:submit={clearaction} on:cancel={clearaction}/>
            {/if}
        {:else if ui.tabsel == "account"}
            {#if ui.action == ""}
                <AccountMenu userid={session.userid} on:action={item_action} />
            {:else if ui.action == "site"}
                <EditSite on:submit={clearaction} on:cancel={clearaction}/>
            {:else if ui.action == "usersettings"}
                <EditUserSettings userid={session.userid} on:submit={clearaction} on:cancel={clearaction}/>
            {:else if ui.action == "changepwd"}
                <ChangePassword userid={session.userid} on:submit={clearaction} on:cancel={clearaction}/>
            {:else if ui.action == "deluser"}
            <DelUser userid={session.userid} username={session.username} on:submit={clearaction} on:cancel={clearaction}/>
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
import EditEntry from "./EditEntry.svelte";
import DelEntry from "./DelEntry.svelte";
import Images from "./Images.svelte";
import EditImage from "./EditImage.svelte";
import DelImage from "./DelImage.svelte";
import Files from "./Files.svelte";
import EditFile from "./EditFile.svelte";
import DelFile from "./DelFile.svelte";
import AccountMenu from "./AccountMenu.svelte";
import EditSite from "./EditSite.svelte";
import EditUserSettings from "./EditUserSettings.svelte";
import ChangePassword from "./ChangePassword.svelte";
import DelUser from "./DelUser.svelte";
import UploadImages from "./UploadImages.svelte";
import SearchImages from "./SearchImages.svelte";

let session = currentSession();
let ui = {};
ui.tabsel = "entries";
ui.action = "";
ui.itemid = 0;

let tablinks;

initPopupHandlers();

function onadditem(e) {
    ui.action = "edit";
    ui.itemid = 0;
}
function item_action(e) {
    ui.action = e.detail.action;
    ui.itemid = e.detail.itemid;
}
function clearaction(e) {
    ui.action = "";
    ui.itemid = 0;
}

</script>

