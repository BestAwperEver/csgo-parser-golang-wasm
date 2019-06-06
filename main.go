package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/markus-wa/demoinfocs-golang/events"
	"io"
	"log"
	"syscall/js"

	dem "github.com/markus-wa/demoinfocs-golang"
	"github.com/markus-wa/demoinfocs-golang/common"
)

const (
	// WASM doesn't enjoy the big buffer sizes allocated by default
	msgQueueBufferSize = 128 * 1024
)

func main() {
	c := make(chan struct{}, 0)

	dem.DefaultParserConfig = dem.ParserConfig{
		MsgQueueBufferSize: msgQueueBufferSize,
	}

	registerCallbacks()

	fmt.Println("WASM Go Initialized")

	<-c
}

func registerCallbacks() {
	js.Global().Set("newParser", js.FuncOf(newParser))
}

// TODO: buffer reader/writer?

type parser struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func md5hex(b []byte) string {
	x := md5.Sum(b)
	return hex.EncodeToString(x[:])
}

func (p *parser) write(b64 string) {
	b, err := base64.StdEncoding.DecodeString(b64)
	checkError(err)

	n, err := p.writer.Write(b)
	// It's fine if there's no reader and we can't write
	if n < len(b) && err != io.ErrClosedPipe {
		checkError(err)
	}
}

func (p *parser) parse(callback js.Value) {
	defer p.reader.Close()
	parser := dem.NewParser(p.reader)
	////


	header, err := parser.ParseHeader()
	checkError(err)
	// TODO: report headerpointer error
	//fmt.Println("Header:", header)
	fmt.Println("Map: " + header.MapName)

	//var players []*common.Player
	var stats []playerStats

	//parser.RegisterEventHandler(func(e events.MatchStartedChanged) {
	//	if e.NewIsStarted {
	//		players = make([]*common.Player, len(parser.GameState().Participants().Playing()))
	//		copy(players, parser.GameState().Participants().Playing())
	//	}
	//})

	parser.RegisterEventHandler(func(e events.RoundEnd) {
		if parser.GameState().TeamTerrorists().Score == 15 && e.Winner == common.TeamTerrorists ||
			parser.GameState().TeamCounterTerrorists().Score == 15 && e.Winner == common.TeamCounterTerrorists ||
			parser.GameState().TotalRoundsPlayed() == 29 {
			//game is over (works for mm)
			for _, p := range parser.GameState().Participants().Playing() {
				stats = append(stats, statsFor(p))
			}
		}
	})

	err = parser.ParseToEnd()
	checkError(err)

	fmt.Println("Parsed")


	////
	b, err := json.Marshal(stats)
	checkError(err)

	// Return result to JS
	callback.Invoke(string(b))
}

type playerStats struct {
	Name    string `json:"name"`
	Kills   int    `json:"kills"`
	Deaths  int    `json:"deaths"`
	Assists int    `json:"assists"`
}

func statsFor(p *common.Player) playerStats {
	return playerStats{
		Name:    p.Name,
		Kills:   p.AdditionalPlayerInformation.Kills,
		Deaths:  p.AdditionalPlayerInformation.Deaths,
		Assists: p.AdditionalPlayerInformation.Assists,
	}
}

func newParser(this js.Value, args []js.Value) interface{} {
	r, w := io.Pipe()
	p := &parser{
		reader: r,
		writer: w,
	}

	m := map[string]interface{}{
		"write": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			p.write(args[0].String())
			return nil
		}),
		"close": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			w.Close()
			return nil
		}),
		"parse": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			go p.parse(args[0])
			return nil
		}),
	}

	// Callback to signal that creation finished, ready to receive data
	args[0].Invoke(m)
	return nil
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
