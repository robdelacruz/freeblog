{#each ui.files as file}
<div class="flex flex-row">
    <div class="flex-grow">
        <a class="action text-sm text-gray-900" href="/?page=file&id={file.fileid}" target="_blank">{file.filename}</a>
    </div>
    <div>
        <a class="action text-xs text-gray-700 mr-2" href="#a" on:click|preventDefault='{e => dispatchAction("edit", file.fileid)}'>edit</a>
        <a class="action text-xs text-gray-700" href="#a" on:click|preventDefault='{e => dispatchAction("del", file.fileid)}'>delete</a>
    </div>
</div>
{/each}

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();

export let userid = 0;

let svcurl = "/api";
let ui = {};
ui.files = [];
ui.status = "";

init(userid);

async function init(userid) {
    ui.status = "";
    let [ff, err] = await findimages(userid);
    if (err != null) {
        console.error(err);
        ui.status = "Server error while fetching images";
    }
    ui.files = ff;
}

// Returns []files, error
async function findimages(userid) {
    let sreq = `${svcurl}/files?filetype=image&userid=${userid}`;
    // Show all images for admin
    if (userid == 1) {
        sreq = `${svcurl}/files?filetype=image`;
    }
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            let s = await res.text();
            return [[], new Error(s)];
        }
        let files = await res.json();
        return [files, null];
    } catch (err) {
        return [[], err];
    }
}

function dispatchAction(action, itemid) {
    dispatch("action", {
        action: action,
        itemid: itemid,
    });
}

</script>

