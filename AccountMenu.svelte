<div class="flex flex-row">
    <div class="flex-grow truncate mr-2">
        <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("blogsettings")}'>Blog Settings</a>
    </div>
    <div>
    </div>
</div>
<div class="flex flex-row">
    <div class="flex-grow truncate mr-2">
        <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("password")}'>Change Password</a>
    </div>
    <div>
    </div>
</div>
<div class="flex flex-row">
    <div class="flex-grow truncate mr-2">
        <a class="action text-sm text-gray-900" href="#a" on:click|preventDefault='{e => dispatchAction("blogsettings")}'>Deactivate Account</a>
    </div>
    <div>
    </div>
</div>

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();

function dispatchAction(action) {
    dispatch("action", {
        action: action,
    });
}
</script>

