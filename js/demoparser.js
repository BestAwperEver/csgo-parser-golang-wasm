
const go = new Go();
// memoryBytes is an Uint8Array pointing to the webassembly memory
let memoryBytes;
// have no idea why do I need "mod"
// just copypasted from a random tutorial
let mod, inst, bytes;
let fileType;

document.getElementById('status').innerText = "Initializing wasm...";

WebAssembly.instantiateStreaming(
  fetch("js/demoparser.wasm", {cache: 'no-cache'}), go.importObject).then((result) => {
  mod = result.module;
  inst = result.instance;
  memoryBytes = new Uint8Array(inst.exports.mem.buffer)
  document.getElementById('status').innerText = "Initialization complete.";
  run();
});

async function run() {
  await go.run(inst);
}

// sets the webassembly memory with the demofile buffer result
// gets called from Go
function gotMem(pointer) {
  memoryBytes.set(bytes, pointer);
  loadDemo();
}

function showHeader() {
  getHeader(onShowHeader);
}

function onShowHeader(header) {
  document.getElementById('status').innerText = "Header"
  displayHeader(JSON.parse(header))
}

// handles file loading
document.getElementById('uploader').addEventListener('change', function() {
  let reader = new FileReader();
  reader.onload = (ev) => {
    bytes = new Uint8Array(ev.target.result);
    initMem(bytes.length);
    let blob = new Blob([bytes], {'type': fileType});
  };
  fileType = this.files[0].type;
  reader.readAsArrayBuffer(this.files[0]);
});

function displayHeader(header) {
  const table = document.getElementById('header');

  Object.keys(header).forEach(function(key) {
    const row = document.createElement('tr');
    row.appendChild(td(key));
    if (key === "PlaybackTime") {
      row.appendChild(td(header[key]/60/1000000000));
    } else {
      row.appendChild(td(header[key]));
    }
    table.appendChild(row);
  });
}

function td(val) {
  const td = document.createElement('td');
  td.innerText = val;
  return td;
}
