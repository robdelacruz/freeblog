import LoginForm from "./LoginForm.svelte";

window.AddLoginForm = function(el) {
    return new LoginForm({target: el});
}

