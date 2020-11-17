import Dashboard from "./Dashboard.svelte";

window.addDashboard = function(el) {
    return new Dashboard({
        target: el,
        props: {},
    });
}

