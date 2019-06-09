package demoparser

import (
  "bytes"
  "encoding/json"
  "github.com/markus-wa/demoinfocs-golang"
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

func (dp *DemoParser) setupOnDemoLoadCb() {
  dp.onDemoLoadCb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    reader := bytes.NewReader(dp.inBuf)
    var err error
    dp.parser = demoinfocs.NewParser(reader)
    dp.header, err = dp.parser.ParseHeader()
    dp.checkError(err)

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
