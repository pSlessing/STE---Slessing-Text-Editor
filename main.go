package main

import (
	"fmt"
	"os"
	"slessingTextEditor/core"
)

func main() {
	fileToOpen := ""
	if len(os.Args) > 1 {
		fileToOpen = os.Args[1]
	}
	editor, err := core.NewEditor()
	if err != nil {
		fmt.Printf("Failed to initialize editor: %v\n", err)
		os.Exit(1)
	}
	editor.Run(fileToOpen)
}
