let frm = document.querySelector("#uploadfiles");
let selfiles = document.querySelector("#selfiles");
let upload = document.querySelector("#upload");

frm.addEventListener("submit", onsubmit);

async function onsubmit(e) {
    e.preventDefault();

    if (files.files.length == 0) {
        console.log("Please select a file");
        return;
    }

    let formdata = new FormData(frm);
    let err = await submitform(formdata);
    if (err != null) {
        console.error(err);
    }
    console.log("success");
}

async function submitform(formdata) {
    let sreq = "/api/uploadfiles/";

    try {
        let res = await fetch(sreq, {
            method: "POST",
            body: formdata,
        });
        if (!res.ok) {
            let err = await res.text();
            return err;
        }
    } catch(err) {
        return err;
    }

    return null;
}

