<form class="flex flex-col panel py-2 px-4 text-sm mb-2">
    <div class="flex flex-row justify-between">
        <h1 class="font-bold mb-1 text-base">Upload Images</h1>
        <div class="flex flex-row justify-start">
            <a class="action self-center rounded text-xs px-1 py-0 mr-2" href="#a">Close</a>
        </div>
    </div>
    <div class="mb-2">
        <label class="block font-bold uppercase text-xs" for="images">Select</label>
        <input class="block border border-gray-500 py-1 px-4 w-full leading-5" id="images" name="images" type="file" multiple>
    </div>
    <div class="mb-2">
        <p class="uppercase italic text-xs">Processing...</p>
    </div>
    <div class="mb-2">
        <p class="font-bold uppercase text-xs">Server error uploading files</p>
    </div>
    <div class="flex flex-row justify-center mb-2 justify-center">
            <button class="inline w-full py-1 px-2 border border-gray-500 font-bold">Upload</button>
    </div>
</form>

