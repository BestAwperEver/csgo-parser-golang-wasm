package demoparser

import (
  "WebAssembly2/go/util/bitarray"
  "WebAssembly2/go/util/elias"
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/markus-wa/demoinfocs-golang"
  "github.com/markus-wa/demoinfocs-golang/events"
  "github.com/markus-wa/demoinfocs-golang/metadata"
  "path"
  "reflect"
  "syscall/js"
  "unsafe"
)

func (dp *DemoParser) setupNextFrame() {
  dp.nextFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    // args[0].Invoke(dp.parseNextFrame())
    return js.ValueOf(dp.parseNextFrame())
  })
}

func (dp *DemoParser) setupGetPositions() {
  dp.getPos = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    b, err := json.Marshal(dp.getPlayersPositions())
    dp.checkError(err)

    // args[0].Invoke(string(b))
    return js.ValueOf(string(b))
  })
}

func (dp *DemoParser) setupGetChickenPositions() {
  dp.getChickenPos = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    b, err := json.Marshal(dp.getChickensPositions())
    dp.checkError(err)

    return js.ValueOf(string(b))
  })
}

func (dp *DemoParser) setupGetWeapons() {
  dp.getWeaponPos = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    b, err := json.Marshal(dp.getWeaponPositions())
    dp.checkError(err)

    return js.ValueOf(string(b))
  })
}

type MapInfo metadata.Map

type InGameCoords struct {
  X, Y float64
}

type BombState struct {
  X, Y    float64
  Defused bool
}

func (m MapInfo) InverseTranslateScale(x, y float64) (X float64, Y float64) {
  X = x*m.Scale + m.PZero.X
  Y = m.PZero.Y - y*m.Scale
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

func (dp *DemoParser) setupGetBombState() {
  dp.getBomb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    b, err := json.Marshal(dp.getBombState())
    dp.checkError(err)

    return js.ValueOf(string(b))
  })
}

func (dp *DemoParser) setupDecodePositions() {
  dp.decodePositions = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    // ba := bitarray.NewDenseBitArray(uint64(len(args) * 64))
    // for i, v := range args {
    //   ba.BlocksArray[i] = args[i].
    // }
    // m := make(map[string]interface{})
    // err := bson.Unmarshal(args[0], m)
    // dp.console.Call("log", "args:", args[0].(bson.Raw))
    // println(args)
    // reader := bytes.NewReader(dp.inBuf)
    // _, err := bson.NewFromIOReader(reader)
    // dp.checkError(err)
    // var a []int64
    // var out struct{ Data bson.Raw }
    // err := bson.Unmarshal(dp.inBuf, &out)
    // dp.checkError(err)
    // dp.console.Call("log", "args:", out.Data.String())
    // dp.log(fmt.Sprintf("%d", args[0].Int()))
    // fmt.Println((int64(15734808) << 32) | int64(662167551))    // should be 6758048643100671906719
    // fmt.Println((int64(-290537564) << 32) | int64(1160063650)) // should be -1247849334479443294
    // fmt.Println((int64(2127879606) << 32) | int64(-425272387)) // should be 9139173321465060285
    // var d int64 = 9139173321465060285
    // fmt.Println(d >> 32)
    // fmt.Println(((d-((d>>32)<<32))))
    // l := -425272387
    // h := 2127879606
    // fmt.Println(-(^int(d-((d>>32)<<32)) + 1))
    // fmt.Println(int64(uint(^(-l - 1))))
    // fmt.Println((int64(h) << 32) | int64(uint(^(-l - 1)))) // should be 9139173321465060285
    // h := -18456577
    // l := -360
    // dp.log(fmt.Sprintf("%d", (int64(h) << 32) | int64(uint(^(-l - 1)))))
    length := args[0].Int()
    longArray := make([]bitarray.Block, length)
    for i := 0; i < length; i++ {
      longArray[i] = bitarray.Block(args[i+1].Int())<<32 | bitarray.Block(uint32(args[i+1+length].Int()))
      // if args[i+1+length].Int() > 0 {
      //   longArray[i] = (int64(args[i+1].Int()) << 32) | (int64(args[i+1+length].Int()))
      // } else {
      //   // dp.log("asdf")
      //   longArray[i] = (int64(args[i+1].Int()) << 32) | (int64(uint32(^(-(args[i+1+length].Int())-1))))
      // }
      dp.log(fmt.Sprintf("%d, %d, %d, %d", i, int64(args[i+1].Int()), int64(args[i+1+length].Int()), longArray[i]))
      // dp.log(fmt.Sprintf("%d", int64(longArray[i])))
    }
    ba := bitarray.NewDenseBitArray(uint64(64 * length))
    ba.BlocksArray = longArray
    bawl := elias.BitArrayWithLength{
      Length:   uint64(64 * length),
      BitArray: ba,
    }
    // b, err := json.Marshal(elias.ArrayToCumulative(elias.EliasGammaDecode(bawl, true)))
    // dp.checkError(err)

    return js.TypedArrayOf(elias.ArrayToCumulative(elias.EliasGammaDecode(bawl, true)))
  })
}

func (dp *DemoParser) setupOnDemoLoadCb() {
  dp.onDemoLoadCb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    reader := bytes.NewReader(dp.inBuf)
    var err error
    dp.parser = demoinfocs.NewParser(reader)
    dp.header, err = dp.parser.ParseHeader()
    dp.checkError(err)
    dp.header.MapName = path.Base(dp.header.MapName)
    dp.parser.RegisterEventHandler(func(e events.WeaponFire) {
      dp.firing[e.Shooter.SteamID] = dp.parser.CurrentTime()
    })
    dp.parser.RegisterEventHandler(func(e events.BombDefused) {
      dp.bombDefused = true
    })
    dp.parser.RegisterEventHandler(func(e events.RoundStart) {
      dp.bombDefused = false
    })
    dp.parser.RegisterEventHandler(dp.handleDataTablesParsed)
    // dp.parser.RegisterEventHandler(dp.weaponEntityHandler)

    dp.log("Ready for operations")

    return nil
  })
}

func (dp *DemoParser) setupInitMemCb() {
  // the length of the array buffer is passed
  // then the buf slice is initialized to that length
  // and a pointer to that slice is passed back to the browser
  dp.initMemCb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    length := args[0].Int()
    dp.console.Call("log", "length:", length)
    dp.inBuf = make([]uint8, length)
    hdr := (*reflect.SliceHeader)(unsafe.Pointer(&dp.inBuf))
    ptr := uintptr(unsafe.Pointer(hdr.Data))
    dp.console.Call("log", "ptr:", ptr)
    // args[1].Invoke(ptr)
    js.Global().Call("gotMem", ptr)
    return nil
  })
}
