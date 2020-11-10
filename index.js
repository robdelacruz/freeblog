import EditUploadEntry from "./EditUploadEntry.svelte";

window.addEditUploadEntry = function(el) {
    return new EditUploadEntry({
        target: el,
        props: {},
    });
}

