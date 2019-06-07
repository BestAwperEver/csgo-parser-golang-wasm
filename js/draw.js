document.addEventListener('DOMContentLoaded', function(){
    // state('init')

    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("js/draw.wasm"), go.importObject).then((result) => {
        go.run(result.instance);
        // state('ready')
    });
}, false);

function draw() {
    drawReplay(document.getElementById('demofile'))
}