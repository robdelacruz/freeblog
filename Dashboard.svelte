<div class="flex flex-row">
    <h1 class="flex-grow font-bold text-xl my-2 mr-2">Dashboard</h1>
    <div class="self-center">
        <a href="#a" class="action text-sm self-center rounded px-0 py-0 mr-2">link 1</a>
        <a href="#a" class="action text-sm self-center rounded px-0 py-0">link 2</a>
    </div>
</div>
<div class="flex flex-row flex-wrap justify-start">
    <div class="flex-grow" style="flex-basis: 640px;">
        <Entries username={session.username} />
    </div>
    <div class="" style="flex-basis: 320px;">
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

let ui = {};
let session = currentSession();

</script>

