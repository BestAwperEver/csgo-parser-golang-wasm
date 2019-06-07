// Copyright [2019] [Mark Farnan]

//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//        http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

// NOTICE:  Much of this demo is a re-write of the 'Moving red Laser' demo by Martin Olsansky https://medium.freecodecamp.org/webassembly-with-golang-is-fun-b243c0e34f02
// It has been re-written to make use of the go-canvas library,  and avoid context calls for drawing.

package main

import (
	"github.com/markus-wa/demoinfocs-golang"
	"github.com/markus-wa/demoinfocs-golang/common"
	"image/color"
	"syscall/js"
	"time"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	"github.com/markfarnan/go-canvas/canvas"
)

type gameState struct{ laserX, laserY, directionX, directionY, laserSize float64 }

var done chan struct{}

var cvs *canvas.Canvas2d
var width float64
var height float64

var gs = gameState{laserSize: 35, directionX: 13.7, directionY: -13.7, laserX: 40, laserY: 40}

var renderDelay = 20 * time.Millisecond

func main() {

	FrameRate := time.Second / renderDelay
	println("Browser FPS:", FrameRate)
	cvs, _ = canvas.NewCanvas2d(true)

	height = float64(cvs.Height())
	width = float64(cvs.Width())

	registerDrawCallbacks()

	//go doEvery(renderDelay, Render) // Kick off the Render function as go routine as it never returns
	<-done
}

func registerDrawCallbacks() {
	js.Global().Set("drawReplay", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cvs.Start(30, Render)
		return nil
	}))
}

// Helper function which calls the required func (in this case 'render') every time.Duration,  Call as a go-routine to prevent blocking, as this never returns
func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

// Called from the 'requestAnnimationFrame' function.   It may also be called seperatly from a 'doEvery' function, if the user prefers drawing to be seperate from the annimationFrame callback
func Render(gc *draw2dimg.GraphicContext) bool {

	renderOneFrame(gc, p.parser)

	if gs.laserX+gs.directionX > width-gs.laserSize || gs.laserX+gs.directionX < gs.laserSize {
		gs.directionX = -gs.directionX
	}
	if gs.laserY+gs.directionY > height-gs.laserSize || gs.laserY+gs.directionY < gs.laserSize {
		gs.directionY = -gs.directionY
	}

	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	gc.Clear()
	// move red laser
	gs.laserX += gs.directionX
	gs.laserY += gs.directionY

	// draws red 🔴 laser
	gc.SetFillColor(color.RGBA{0xff, 0x00, 0x00, 0xff})
	gc.SetStrokeColor(color.RGBA{0xff, 0x00, 0x00, 0xff})

	gc.BeginPath()
	//gc.ArcTo(gs.laserX, gs.laserY, gs.laserSize, gs.laserSize, 0, math.Pi*2)
	draw2dkit.Circle(gc, gs.laserX, gs.laserY, gs.laserSize)
	gc.FillStroke()
	gc.Close()

	return true
}

func drawPlayer(gc *draw2dimg.GraphicContext, player *common.Player) {

	pos_x := player.Position.X / de_dust2_wight * width
	pos_y := player.Position.Y / de_dust2_height * height

	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	gc.Clear()

	// move red laser
	gs.laserX += gs.directionX
	gs.laserY += gs.directionY

	var clr color.RGBA

	if player.Team == common.TeamTerrorists {
		clr = color.RGBA{0xff, 0x00,0xff,0xff}
	} else if player.Team == common.TeamCounterTerrorists {
		clr = color.RGBA{0x00, 0x00,0xff,0xff}
	} else {
		clr = color.RGBA{0x80, 0x80, 0x80, 0xff}
	}

	gc.SetFillColor(clr)
	gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})

	gc.BeginPath()
	//gc.ArcTo(gs.laserX, gs.laserY, gs.laserSize, gs.laserSize, 0, math.Pi*2)
	draw2dkit.Circle(gc, pos_x, pos_y, 20)

	gc.FillStroke()
	gc.Close()
}

const (
	de_dust2_wight = 9664
	de_dust2_height = 7488
)

func renderOneFrame(gc *draw2dimg.GraphicContext, parser demoinfocs.Parser) bool {
	next, err := parser.ParseNextFrame()
	checkError(err)
	if !next {
		return false
	}

	for _, player := range parser.GameState().Participants().Playing() {
		drawPlayer(gc, player)
	}

	return true
}
