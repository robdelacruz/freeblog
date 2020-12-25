{#if ui.loadstatus != ""}
    <div class="mb-2">
        <p class="uppercase italic text-xs">{ui.loadstatus}</p>
    </div>
    <div class="mb-2">
        <a class="action self-center rounded text-xs px-0 py-0 mr-2" href="#a" on:click|preventDefault="{e => init()}">Retry</a>
    </div>
{:else}
    {#if userid == 1}
    <div class="flex flex-row">
        <div class="flex-grow truncate mr-2">
            <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("site")}'>Site Settings</a>
        </div>
        <div>
        </div>
    </div>
    {/if}
    {#if !ui.site.isgroup}
    <div class="flex flex-row">
        <div class="flex-grow truncate mr-2">
            <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("usersettings")}'>User Settings</a>
        </div>
        <div>
        </div>
    </div>
    {/if}
    <div class="flex flex-row">
        <div class="flex-grow truncate mr-2">
            <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("changepwd")}'>Change Password</a>
        </div>
        <div>
        </div>
    </div>
    {#if userid > 1}
    <div class="flex flex-row">
        <div class="flex-grow truncate mr-2">
            <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("deluser")}'>Delete Account</a>
        </div>
        <div>
        </div>
    </div>
    {/if}
{/if}

<script>
import {currentSession, initPopupHandlers, find} from "./helpers.js";
let session = currentSession();
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
export let userid = 0;

let svcurl = "/api";
let ui = {};
ui.site = null;

init();

async function init() {
    ui.loadstatus = "loading site settings...";

    let sreq = `${svcurl}/site`
    let [site, err] = await find(sreq);
    if (err !=  null) {
        console.error(err);
        ui.loadstatus = "Server error loading site settings";
        ui.site = null;
        return;
    }

    ui.loadstatus = "";
    ui.site = site;
}

function dispatchAction(action) {
    dispatch("action", {
        action: action,
    });
}

</script>

