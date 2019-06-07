package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/require"
)

const demoBufferSize = 1024 * 2048 // 2 MB

func TestNewParser(t *testing.T) {
	wg := new(sync.WaitGroup)

	callback := func(this js.Value, args []js.Value) interface{} {
		handleResult := func(this js.Value, result []js.Value) interface{} {
			fmt.Println(result[0])
			fmt.Println("Finished")
			wg.Done()
			return nil
		}

		parser := args[0]
		parser.Call("parseFinalStats", js.ValueOf(js.FuncOf(handleResult)))

		f, err := os.Open("../default.dem")
		require.Nil(t, err)

		b := make([]byte, demoBufferSize)
		n := demoBufferSize
		for n == demoBufferSize {
			n, err = f.Read(b)
			require.Nil(t, err)
			parser.Call("write", js.ValueOf(base64.StdEncoding.EncodeToString(b)))
		}

		parser.Call("close")
		return nil
	}

	wg.Add(1)

	newParser(js.ValueOf(nil), []js.Value{js.ValueOf(js.FuncOf(callback))})

	wg.Wait()
}
