package core

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)

// WriteBufferToFile writes the textBuffer contents to the specified file
func (e *EditorCore) WriteBufferToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for i, line := range e.TextBuffer {
		lineStr := string(line)
		_, err := writer.WriteString(lineStr)
		if err != nil {
			return err
		}

		if i < len(e.TextBuffer)-1 {
			_, err := writer.WriteString("\n")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// SaveCurrentState saves the current textBuffer to the sourceFile
// If no sourceFile is set, it returns an empty string to indicate save-as is needed
func (e *EditorCore) SaveCurrentState() (string, error) {
	if e.SourceFile == "" {
		// No file name set, caller should handle save-as
		return "", fmt.Errorf("no filename set")
	} else {
		// Save to existing file
		err := e.WriteBufferToFile(e.SourceFile)
		if err != nil {
			// Display error message to user
			e.PrintMessage(0, e.Rows, tcell.ColorRed, tcell.ColorDefault,
				fmt.Sprintf("Error saving file: %s", err.Error()))
			e.Terminal.Show()
			e.Terminal.PollEvent()
			return e.SourceFile, err
		}
		return e.SourceFile, nil
	}
}

// OpenFile opens a specific file and reads it into a text buffer
func (e *EditorCore) OpenFile(filename string) ([][]rune, error) {
	textBuffer := [][]rune{}
	file, err := os.Open(filename)
	if err != nil {
		// Return empty buffer if file doesn't exist
		return append(textBuffer, []rune{}), err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		line := scanner.Text()
		textBuffer = append(textBuffer, []rune{})
		for _, ch := range line {
			textBuffer[lineNumber] = append(textBuffer[lineNumber], rune(ch))
		}
		lineNumber++
	}
	if lineNumber == 0 {
		textBuffer = append(textBuffer, []rune{})
	}
	e.SourceFile = filename
	return textBuffer, nil
}

// Note
// Returns both directories and files, something should probably be done so this is used properly
func (e *EditorCore) GetCurrentPathFiles() ([]string, error) {
	//
	currentWd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(currentWd)

	if err != nil {
		return nil, err
	}

	returnSlice := []string{}

	for _, e := range entries {
		//For now, dont handle directories in the working directory
		//TODO: Add that later
		if !e.IsDir() {
			returnSlice = append(returnSlice, e.Name())
		}
	}

	return returnSlice, nil

}
