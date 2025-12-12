// main.go //Blabla
package main

import (
	"fmt"
	"os"
	"slessingTextEditor/core"
)

func main() {
	editor, err := core.NewEditor()
	if err != nil {
		fmt.Printf("Failed to initialize editor: %v\n", err)
		os.Exit(1)
	}

	// Plugins are auto-loaded in Run()
	// Wow, i can comment in here
	editor.Run()
}