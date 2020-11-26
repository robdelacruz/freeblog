<form class="flex-grow flex flex-col panel mx-auto py-2 px-8 text-sm h-full">
    <div class="flex flex-row justify-between">
        <h1 class="font-bold mb-1 text-base">Edit Entry</h1>
        <div>
            <a class="action self-center rounded text-xs px-0 py-0" href="#a" on:click|preventDefault=''>Cancel</a>
        </div>
    </div>
    <div class="mb-2">
        <label class="block font-bold uppercase text-xs" for="title">title</label>
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="title" name="title" type="text" value="">
    </div>
    <div class="flex-grow flex flex-col mb-2">
        <label class="block font-bold uppercase text-xs" for="username">entry</label>
        <textarea class="flex-grow block border border-gray-500 py-1 px-4 w-full leading-5" id="entry" name="entry"></textarea>
    </div>
    <div class="mb-2">
        <label class="block font-bold uppercase text-xs" for="tags">tags</label>
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="tags" name="tags" type="text" value="">
    </div>
    <div class="mb-2">
        <p class="font-bold uppercase text-xs"></p>
    </div>
    <div class="flex flex-row justify-center mb-2 justify-center">
            <button class="inline w-full py-1 px-2 border border-gray-500 font-bold">Submit</button>
    </div>
</form>

<script>
import {onMount, createEventDispatcher} from "svelte";
let dispatch = createEventDispatcher();
import {currentSession} from "./helpers.js";
export let entryid;

let svcurl = "/api";
let session = currentSession();

let ui = {};
ui.entryid = entryid;
ui.mode = "";

// Returns [entry, err]
async function findentry(entryid) {
    let sreq = `${svcurl}/entry?id=${entryid}`;
    try {
        let res = await fetch(sreq, {method: "GET"});
        if (!res.ok) {
            if (res.status == 404) {
                return [null, null];
            }
            let s = await res.text();
            return [null, new Error(s)];
        }
        let entry = await res.json();
        return [entry, null];
    } catch(err) {
        return [null, err];
    }
}

// Returns [savedentry, err]
async function submitentry(e) {
    let sreq = `${svcurl}/entry/`;
    let method = "";
    if (e.entryid == 0) {
        method = "POST";
    } else {
        method = "PUT";
    }

    try {
        let res = await fetch(sreq, {
            method: method,
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(entry),
        });
        if (!res.ok) {
            let s = await res.text();
            console.error(s);
            return [null, new Error(s)];
        }
        let savedentry = await res.json();
        return [savedentry, null];
    } catch(err) {
        return [null, err];
    }
}
</script>

