<div class="panel py-2 px-4 mb-2 mr-2">
    <h1 class="font-bold mb-1 text-base">Entries</h1>
{#each ui.entries as entry}
    <div class="flex flex-row py-1">
        <div class="flex-grow">
            <a class="action text-sm text-gray-900" href="/entry?id={entry.entryid}">{entry.title}</a>
        </div>
        <a class="action text-xs text-gray-700 mr-4" href="#a">edit</a>
        <a class="action text-xs text-gray-700" href="#a">delete</a>
    </div>
{/each}
</div>

<script>
export let username = "";

let svcurl = "/api";
let ui = {};
ui.entries = [];
ui.err = "";

init(username);

async function init(username) {
    try {
        ui.entries = await findentries(username);
    } catch(err) {
        console.error(err);
        ui.entries = [];
        ui.err = err;
    }
}

async function findentries(username) {
    let sreq = `${svcurl}/entries?username=${username}`;
    let res = await fetch(sreq, {method: "GET"});
    if (!res.ok) {
        let err = await res.text();
        return Promise.reject(err);
    }
    let entries = await res.json();
    return entries;
}

</script>

