
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

const elem = document.getElementById('canvases'),
  elemLeft = elem.offsetLeft,
  elemTop = elem.offsetTop,
  elements = [];

elem.addEventListener('click', function(event) {
  const x = event.pageX - elemLeft, y = event.pageY - elemTop;

  json_pos = JSON.parse(inverseTranslate(x, y));

  document.getElementById('xcoord').innerText = json_pos.X;
  document.getElementById('ycoord').innerText = json_pos.Y;

}, false);

function printPositions() {
  if (nextFrame()) {
    console.log(JSON.parse(getPositions()))
    console.log(JSON.parse(getBomb()))
  }
}

// {{/*  <label for="x_mult">X-multiplier</label>*/}}
// {{/*  <input type="range" min="1" max="3" value="1" step="0.01" id="x_mult" onchange="updateXMultiplier(this.value);">*/}}
// {{/*  <label for="y_mult">Y-multiplier</label>*/}}
// {{/*  <input type="range" min="1" max="3" value="1" step="0.01" id="y_mult" onchange="updateYMultiplier(this.value);">*/}}


let playersPositions;

function drawFrame() {
  if (nextFrame()) {
    playersPositions = JSON.parse(getPositions());
    displayPlayersPositions(playersPositions)
  }
}

function changeRadar() {
  if (mapName !== "none") {
    const mapCanvas = document.getElementById('mapCanvas');
    const ctx = mapCanvas.getContext('2d');
    const img = new Image;
    img.onload = function(){
      ctx.drawImage(img,0,0);
    };
    if (!radar_lower) {
      img.src = "https://raw.githubusercontent.com/BestAwperEver/csgo-parser-golang-wasm/master/radars/"+mapName+"_lower_radar.png";
      radar_lower = true
    } else {
      img.src = "https://raw.githubusercontent.com/BestAwperEver/csgo-parser-golang-wasm/master/radars/"+mapName+"_radar.png";
      radar_lower = false
    }
  }
}

let draw = true;
let anim;
let mapName = "none";
let radar_lower = false;
// let map_displayed = false;

function drawMap() {
  const mapCanvas = document.getElementById('mapCanvas');
  const ctx = mapCanvas.getContext('2d');
  const img = new Image;
  img.onload = function(){
    ctx.drawImage(img,0,0);
  };
  getHeader((header) => {
    mapName = JSON.parse(header).MapName.toLowerCase()
    img.src = "https://raw.githubusercontent.com/BestAwperEver/csgo-parser-golang-wasm/master/radars/"+mapName+"_radar.png";
    radar_lower = false
  });
  // map_displayed = true;
}

function startDrawFrames() {
  drawMap();
  drawFrames()
}

function drawFrames() {
  drawFrame();
  if (draw) anim = window.requestAnimationFrame(drawFrames);
}

function changeDraw() {
  // draw = !draw;
  window.cancelAnimationFrame(anim)
}

let x_mult = 4526;
let y_mult = 4478;
let x_center = 0.549;
let y_center = 0.724;

/**
 * @return {string}
 */
function LightenDarkenColor(col,amt) {
  var usePound = false;
  if ( col[0] == "#" ) {
    col = col.slice(1);
    usePound = true;
  }

  var num = parseInt(col,16);

  var r = (num >> 16) + amt;

  if ( r > 255 ) r = 255;
  else if  (r < 0) r = 0;

  var b = ((num >> 8) & 0x00FF) + amt;

  if ( b > 255 ) b = 255;
  else if  (b < 0) b = 0;

  var g = (num & 0x0000FF) + amt;

  if ( g > 255 ) g = 255;
  else if  ( g < 0 ) g = 0;

  return (usePound?"#":"") + (g | (b << 8) | (r << 16)).toString(16);
}

function displayPlayersPositions(positions) {

  // const de_dust2_wight = 9664, de_dust2_height = 7488;
  const canvas = document.getElementById("myCanvas");
  const ctx = canvas.getContext("2d");
  ctx.clearRect(0, 0, canvas.width, canvas.height); // clear canvas

  ctx.font = "16px Arial";

  let pos_x;
  let pos_y;
  let json_pos;
  for (let i = 0; i < positions.length; i++) {
    const player = positions[i];
    if (player.IsAlive === false) continue;
    // pos_x = (de_dust2_wight/2 + player.Position.X) / de_dust2_wight * canvas.width;
    // pos_x = (.5 + x_mult * player.Position.X / de_dust2_wight) * canvas.width;
    // pos_x = (Number(x_center) + player.Position.X / Number(x_mult)) * canvas.width;
    // pos_y = (de_dust2_height/2 - y_mult * player.Position.Y) / de_dust2_height * canvas.height;
    // pos_y = (Number(y_center) - player.Position.Y / Number(y_mult)) * canvas.height;
    json_pos = JSON.parse(translate(player.Position.X, player.Position.Y));
    pos_x = json_pos.X;
    pos_y = json_pos.Y;
    // ctx.fillStyle = '#FACE8D';
    ctx.fillStyle = '#FFFFFF';
    ctx.fillText(player.SteamID + " (" + player.HP + ", " + player.Armor + ")", pos_x, pos_y - 20);
    ctx.beginPath();
    if (player["Team"] === 2) {
      // terrorist
      ctx.fillStyle = '#FFFF00';
    } else {
      //counter-terrorist
      ctx.fillStyle = '#0000FF';
    }
    if (player.IsBlinded) {
      ctx.fillStyle = '#FFFFFF';
      // ctx.fillStyle = LightenDarkenColor(ctx.fillStyle, player.Flash)
    }
    ctx.arc(pos_x, pos_y, 4, 0, 2 * Math.PI);
    ctx.fill();
    ctx.strokeStyle = '#000000';
    ctx.stroke();

    if (player.HasBomb) {
      ctx.strokeStyle = '#FF0000';
      ctx.fillStyle = '#440000';

      ctx.beginPath();
      ctx.arc(pos_x, pos_y, 2, 0, 2 * Math.PI);
      ctx.fill();
      ctx.stroke();
    }

    ctx.beginPath();
    if (player.IsFiring) {
      ctx.strokeStyle = '#FF0000';
    } else {
      ctx.strokeStyle = '#000000';
    }
    ctx.arc(pos_x, pos_y, 6, -(player.ViewX + 20)/180 * Math.PI,-(player.ViewX - 20)/180 * Math.PI);
    // ctx.fill();
    ctx.stroke();
  }

  bomb_pos = JSON.parse(getBomb());

  if (Number(bomb_pos.X) !== 0) {
    bomb_pos = JSON.parse(translate(bomb_pos.X, bomb_pos.Y));
    ctx.strokeStyle = '#FF0000';
    ctx.fillStyle = '#440000';

    ctx.beginPath();
    ctx.arc(bomb_pos.X, bomb_pos.Y, 2, 0, 2 * Math.PI);
    ctx.fill();
    ctx.stroke();
  }

}

function showHeader() {
  getHeader(onShowHeader);
}

function onShowHeader(header) {
  document.getElementById('status').innerText = "Header";
  displayHeader(JSON.parse(header))
}

// handles file loading
document.getElementById('uploader').addEventListener('change', function() {
  // map_displayed = false;
  let reader = new FileReader();
  reader.onload = (ev) => {
    bytes = new Uint8Array(ev.target.result);
    initMem(bytes.length);
    let blob = new Blob([bytes], {'type': fileType});
  };
  fileType = this.files[0].type;
  reader.readAsArrayBuffer(this.files[0]);
});

function updateXMultiplier(value) {
  x_mult = value;
  displayPlayersPositions(playersPositions)
}

function updateYMultiplier(value) {
  y_mult = value;
  displayPlayersPositions(playersPositions)
}

function updateXCenter(value) {
  x_center = value;
  displayPlayersPositions(playersPositions)
}

function updateYCenter(value) {
  y_center = value;
  displayPlayersPositions(playersPositions)
}

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
