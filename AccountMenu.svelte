<div class="flex flex-row">
    <div class="flex-grow truncate mr-2">
        <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("blogsettings")}'>Blog Settings</a>
    </div>
    <div>
    </div>
</div>
<div class="flex flex-row">
    <div class="flex-grow truncate mr-2">
        <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("changepwd")}'>Change Password</a>
    </div>
    <div>
    </div>
</div>
{#if session.userid != 1}
<div class="flex flex-row">
    <div class="flex-grow truncate mr-2">
        <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("deluser")}'>Delete Account</a>
    </div>
    <div>
    </div>
</div>
{/if}

<script>
import {currentSession, initPopupHandlers} from "./helpers.js";
let session = currentSession();
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();

function dispatchAction(action) {
    dispatch("action", {
        action: action,
    });
}
</script>

