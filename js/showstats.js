document.addEventListener('DOMContentLoaded', function(){
	state('init');

	const go = new Go();
	WebAssembly.instantiateStreaming(fetch("js/showstats.wasm"), go.importObject).then((result) => {
		go.run(result.instance);
		state('ready')
	});
}, false);

const demoBufferSize = 1024 * 2048; // 2 MB

function parseHeader() {
	state("constructing parser");
	newParser((parser) => {
		// state("parser is constructed");
		state("getting header");
		parser.getHeader((header) => {
			console.log('done');
			state("done");
			// let jsonHeader = JSON.parse(header);
			displayHeader(JSON.parse(header))
		});

		const reader = new FileReader();
		reader.onload = function() {
			const data = reader.result;
			console.log('writing data');
			for (let offset = 0; offset < data.byteLength; offset += demoBufferSize) {
				const arr = readDataIntoBuffer(data, offset);
				let base64 = btoa(arr.reduce((data, byte) => (data.push(String.fromCharCode(byte)), data), []).join(''));
				parser.write(base64);
			}
			// console.log('closing pipe');
			// parser.close();
		};
		reader.readAsArrayBuffer(document.getElementById('demofile').files[0]);
	});
}

function parseFinalStats() {
	state("creating parser");
	newParser((parser) => {
		state("parsing");
		console.log('parse');
		parser.parseFinalStats((stats) => {
			console.log('done');
			state("done");
			displayStats(JSON.parse(stats));
		});

		const reader = new FileReader();
		reader.onload = function() {
			const data = reader.result;
			console.log('writing data');
			for (let offset = 0; offset < data.byteLength; offset += demoBufferSize) {
				const arr = readDataIntoBuffer(data, offset);
				let base64 = btoa(arr.reduce((data, byte) => (data.push(String.fromCharCode(byte)), data), []).join(''))
				parser.write(base64);
			}
			console.log('closing pipe');
			parser.close();
		};
		reader.readAsArrayBuffer(document.getElementById('demofile').files[0])
	})
}

function readHeader() {

}

function getHeader() {

}

// class Parser {
// 	demofile = "";
// 	parsingStarted = false;
// 	headerRead = false;
// 	gameState = {};
//
// 	constructor(demofile) {
// 		this.demofile = demofile;
// 	}
//
// 	getHeader() {
//
// 	}
//
// 	readHeader() {
//
// 	}
//
// 	parseFirstFrame() {
//
// 	}
//
// 	parseNextFrame() {
//
// 	}
//
// 	get gameState() {
// 		return this.gameState
// 	}
//
// }


function readDataIntoBuffer(data, offset) {
	if (offset + demoBufferSize <= data.byteLength) {
		return new Uint8Array(data, offset, demoBufferSize);
	}
	return new Uint8Array(data, offset, data.byteLength-offset);
}

function state(state) {
	document.getElementById('state').innerText = state;
}

function displayStats(stats) {
	stats = stats.sort((a, b) => a.playerName.localeCompare(b.playerName)).sort((a, b) => b.Kills - a.Kills);
	const table = document.getElementById('stats');
	// stats.forEach(p => {
	// 	const row = document.createElement('tr');
	// 	row.appendChild(td(p.playerName));
	// 	row.appendChild(td(p.Kills));
	// 	row.appendChild(td(p.Deaths));
	// 	row.appendChild(td(p.Assists));
	// 	table.appendChild(row);
	// });
	stats.forEach(p => {
		const row = document.createElement('tr');
		Object.keys(p).forEach(function(key) {
			// console.table('Key : ' + key + ', Value : ' + data[key])
			row.appendChild(td(p[key]));
		});
		// row.appendChild(td(p.playerName));
		// row.appendChild(td(p.Kills));
		// row.appendChild(td(p.deaths));
		// row.appendChild(td(p.assists));
		table.appendChild(row);
	});
}

function displayHeader(header) {
	const table = document.getElementById('header');

	// header.forEach(p => {
	// 	const row = document.createElement('tr');
	// 	Object.keys(p).forEach(function(key) {
	// 		row.appendChild(td(p[key]));
	// 	});
	// 	table.appendChild(row);
	// });

	Object.keys(header).forEach(function(key) {
		const row = document.createElement('tr');
		row.appendChild(td(key));
		row.appendChild(td(header[key]));
		table.appendChild(row);
	});
}

function td(val) {
	const td = document.createElement('td');
	td.innerText = val;
	return td;
}