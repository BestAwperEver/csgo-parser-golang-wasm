package demoparser

import (
  "bytes"
  "encoding/json"
  "github.com/markus-wa/demoinfocs-golang"
  "github.com/markus-wa/demoinfocs-golang/events"
  "github.com/markus-wa/demoinfocs-golang/metadata"
  "reflect"
  "syscall/js"
  "unsafe"
)

func (dp *DemoParser) setupNextFrame() {
  dp.nextFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    //args[0].Invoke(dp.parseNextFrame())
    return js.ValueOf(dp.parseNextFrame())
  })
}

func (dp *DemoParser) setupGetPositions() {
  dp.getPos = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    b, err := json.Marshal(dp.getPlayersPositions())
    dp.checkError(err)

    //args[0].Invoke(string(b))
    return js.ValueOf(string(b))
  })
}

type MapInfo metadata.Map

type InGameCoords struct {
  X, Y float64
}

func (m MapInfo) InverseTranslateScale(x, y float64) (X float64, Y float64) {
  X = x * m.Scale + m.PZero.X
  Y = m.PZero.Y - y * m.Scale
  return
}

func (dp *DemoParser) setupInverseTranslate() {
  dp.inverseTranslate = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    meta := MapInfo(metadata.MapNameToMap[dp.header.MapName])
    X, Y := meta.InverseTranslateScale(args[0].Float(), args[1].Float())
    coords := InGameCoords{X, Y,}

    // dp.log(fmt.Sprintf("%.2f, %.2f", coords.X, coords.Y))

    b, err := json.Marshal(coords)
    dp.checkError(err)

    return js.ValueOf(string(b))
  })
}

func (dp *DemoParser) setupTranslate() {
  dp.translate = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    meta := metadata.MapNameToMap[dp.header.MapName]
    X, Y := meta.TranslateScale(args[0].Float(), args[1].Float())
    coords := InGameCoords{X, Y,}

    b, err := json.Marshal(coords)
    dp.checkError(err)

    return js.ValueOf(string(b))
  })
}

func (dp *DemoParser) setupGetHeader() {
  dp.getHeader = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    b, err := json.Marshal(getMap(dp.header))
    dp.checkError(err)

    args[0].Invoke(string(b))
    return nil
  })
}

func (dp *DemoParser) setupShutdownCb() {
  dp.shutdownCb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    dp.done <- struct{}{}
    return nil
  })
}

func (dp *DemoParser) setupGetBombPosition() {
  dp.getBomb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    b, err := json.Marshal(dp.getBombPosition())
    dp.checkError(err)

    return js.ValueOf(string(b))
  })
}

func (dp *DemoParser) setupOnDemoLoadCb() {
  dp.onDemoLoadCb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    reader := bytes.NewReader(dp.inBuf)
    var err error
    dp.parser = demoinfocs.NewParser(reader)
    dp.header, err = dp.parser.ParseHeader()
    dp.checkError(err)
    dp.parser.RegisterEventHandler(func(e events.WeaponFire){
      dp.firing[e.Shooter.SteamID] = dp.parser.CurrentTime()
    })

    dp.log("Ready for operations")

    return nil
  })
}

func (dp *DemoParser) setupInitMemCb() {
  // the length of the array buffer is passed
  // then the buf slice is initialized to that length
  // and a pointer to that slice is passed back to the browser
  dp.initMemCb = js.FuncOf(func(this js.Value, i []js.Value) interface{} {
    length := i[0].Int()
    dp.console.Call("log", "length:", length)
    dp.inBuf = make([]uint8, length)
    hdr := (*reflect.SliceHeader)(unsafe.Pointer(&dp.inBuf))
    ptr := uintptr(unsafe.Pointer(hdr.Data))
    dp.console.Call("log", "ptr:", ptr)
    js.Global().Call("gotMem", ptr)
    return nil
  })
}
