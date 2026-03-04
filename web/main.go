package main

import (
	"syscall/js"

	"github.com/doug-benn/Jacques/internal/processor"
)

func main() {
	js.Global().Set("cleanSchema", js.FuncOf(cleanSchema))
	js.Global().Set("ready", js.FuncOf(ready))
	select {}
}

func cleanSchema(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return ""
	}

	sql := args[0].String()
	experimental := args[1].Bool()

	opts := &processor.Options{ExperimentalFolding: experimental}
	return processor.Process(sql, opts)
}

func ready(this js.Value, args []js.Value) interface{} {
	return true
}
