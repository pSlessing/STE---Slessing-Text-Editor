package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

type Settings struct {
	BGColor          tcell.Color `json:"bg_color"`
	FGColor          tcell.Color `json:"fg_color"`
	StatusBGColor    tcell.Color `json:"status_bg_color"`
	StatusFGColor    tcell.Color `json:"status_fg_color"`
	MsgBGColor       tcell.Color `json:"msg_bg_color"`
	MsgFGColor       tcell.Color `json:"msg_fg_color"`
	LineCountBGColor tcell.Color `json:"line_count_bg_color"`
	LineCountFGColor tcell.Color `json:"line_count_fg_color"`
	ErrorBGColor     tcell.Color `json:"error_bg_color"`
	ErrorFGColor     tcell.Color `json:"error_fg_color"`
}

type StyleSet struct {
	Main      tcell.Style
	Status    tcell.Style
	Linecount tcell.Style
	Message   tcell.Style
	Error     tcell.Style
}

// GetDefaultSettings returns the default configuration
func GetDefaultSettings() Settings {
	return Settings{
		BGColor:          tcell.ColorBlack,
		FGColor:          tcell.ColorWhite,
		StatusBGColor:    tcell.ColorWhite,
		StatusFGColor:    tcell.ColorBlack,
		MsgBGColor:       tcell.ColorWhite,
		MsgFGColor:       tcell.ColorBlack,
		LineCountBGColor: tcell.ColorWhite,
		LineCountFGColor: tcell.ColorLightBlue,
		ErrorBGColor:     tcell.ColorBlack,
		ErrorFGColor:     tcell.ColorRed,
	}
}

// SaveSettings saves the current settings to a JSON file
func SaveSettings(settings Settings) error {
	// Get OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config directory: %w", err)
	}
	configDir = filepath.Join(configDir, "SlessingTextEditor")

	// Ensure the config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert settings to JSON string
	jsonStr, err := SettingsToJSON(settings)
	if err != nil {
		return fmt.Errorf("failed to convert settings to JSON: %w", err)
	}

	// Write to file using similar pattern to WriteBufferToFile
	configPath := filepath.Join(configDir, "config.json")
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	_, err = writer.WriteString(jsonStr)
	if err != nil {
		return fmt.Errorf("failed to write settings to file: %w", err)
	}

	return nil
}

// LoadSettings loads settings from a JSON file, creating default config if file doesn't exist
func LoadSettings() (Settings, error) {
	// Get OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Settings{}, fmt.Errorf("failed to get user config directory: %w", err)
	}
	configDir = filepath.Join(configDir, "SlessingTextEditor")

	configPath := filepath.Join(configDir, "config.json")

	// Try to open the file, similar to OpenFile pattern
	file, err := os.Open(configPath)
	if err != nil {
		// File doesn't exist, create default config
		defaultSettings := GetDefaultSettings()
		saveErr := SaveSettings(defaultSettings)
		if saveErr != nil {
			return Settings{}, fmt.Errorf("failed to create default config: %w", saveErr)
		}
		return defaultSettings, nil
	}
	defer file.Close()

	// Read the file content
	var jsonContent string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonContent += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		return Settings{}, fmt.Errorf("failed to read config file: %w", err)
	}

	// Convert JSON to settings
	settings, err := JSONToSettings(jsonContent)
	if err != nil {
		return Settings{}, fmt.Errorf("failed to parse settings JSON: %w", err)
	}

	return settings, nil
}

// SettingsToJSON converts a Settings struct to a JSON string
func SettingsToJSON(settings Settings) (string, error) {
	jsonData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings to JSON: %w", err)
	}
	return string(jsonData), nil
}

// JSONToSettings converts a JSON string to a Settings struct
func JSONToSettings(jsonStr string) (Settings, error) {
	var settings Settings
	if err := json.Unmarshal([]byte(jsonStr), &settings); err != nil {
		return Settings{}, fmt.Errorf("failed to unmarshal JSON to settings: %w", err)
	}
	return settings, nil
}

// ApplySettings applies the loaded settings to the global color variables
func (e *EditorCore) ApplySettings(settings Settings) {
	e.Styles.Main = tcell.StyleDefault.Background(settings.BGColor).Foreground(settings.FGColor)
	e.Styles.Status = tcell.StyleDefault.Background(settings.StatusBGColor).Foreground(settings.StatusFGColor)
	e.Styles.Message = tcell.StyleDefault.Background(settings.MsgBGColor).Foreground(settings.MsgFGColor)
	e.Styles.Linecount = tcell.StyleDefault.Background(settings.LineCountBGColor).Foreground(settings.LineCountFGColor)
	e.Styles.Error = tcell.StyleDefault.Background(settings.ErrorBGColor).Foreground(settings.ErrorFGColor)

}

// GetCurrentSettings creates a Settings struct from the current global variables
func (e *EditorCore) GetCurrentSettings() Settings {
	mainfg, mainbg, _ := e.Styles.Main.Decompose()
	statusfg, statusbg, _ := e.Styles.Status.Decompose()
	msgfg, msgbg, _ := e.Styles.Message.Decompose()
	linecountfg, linecountbg, _ := e.Styles.Linecount.Decompose()
	errorfg, errorbg, _ := e.Styles.Error.Decompose()
	return Settings{
		BGColor:          mainbg,
		FGColor:          mainfg,
		StatusBGColor:    statusbg,
		StatusFGColor:    statusfg,
		MsgBGColor:       msgbg,
		MsgFGColor:       msgfg,
		LineCountBGColor: linecountbg,
		LineCountFGColor: linecountfg,
		ErrorBGColor:     errorbg,
		ErrorFGColor:     errorfg,
	}
}

// Example usage:
/*
func main() {
	// Load settings on startup (creates default if doesn't exist)
	settings, err := LoadSettings()
	if err != nil {
		fmt.Printf("Error loading settings: %v\n", err)
		return
	}
	ApplySettings(settings)

	// Your application code here...

	// Save settings when modified
	currentSettings := GetCurrentSettings()
	err = SaveSettings(currentSettings)
	if err != nil {
		fmt.Printf("Error saving settings: %v\n", err)
	}
}
*/

// colorSetting points at one foreground/background slot of a StyleSet style,
// so the settings loop can index into it instead of switching on position.
type colorSetting struct {
	style        *tcell.Style
	isBackground bool
}

// colorSettingsTable is the single source of truth for what "setting N"
// means: its order defines SettingsLength and is shared by every position ->
// style lookup (getCurrentColorPos, updateStylesHelper).
func (e *EditorCore) colorSettingsTable() []colorSetting {
	return []colorSetting{
		{&e.Styles.Main, false},
		{&e.Styles.Main, true},
		{&e.Styles.Status, false},
		{&e.Styles.Status, true},
		{&e.Styles.Message, false},
		{&e.Styles.Message, true},
		{&e.Styles.Linecount, false},
		{&e.Styles.Linecount, true},
		{&e.Styles.Error, false},
		{&e.Styles.Error, true},
	}
}

// Helper function to get the current color position based on the selected setting
func (e *EditorCore) getCurrentColorPos(currentPos int, colorNames []string) int {
	settings := e.colorSettingsTable()
	if currentPos < 0 || currentPos >= len(settings) {
		return 0
	}

	setting := settings[currentPos]
	fg, bg, _ := setting.style.Decompose()
	currentColor := fg
	if setting.isBackground {
		currentColor = bg
	}

	for i, colorName := range colorNames {
		if e.getColorFromName(colorName) == currentColor {
			return i
		}
	}

	// The stored color isn't one of the named presets (e.g. an arbitrary RGB
	// value) - signal "no match" instead of silently claiming it's colorNames[0].
	return -1
}

// Helper function to convert color name to tcell.Color
func (e *EditorCore) getColorFromName(colorName string) tcell.Color {
	colors := map[string]tcell.Color{
		"Black":    tcell.ColorBlack,
		"Red":      tcell.ColorRed,
		"Green":    tcell.ColorGreen,
		"Yellow":   tcell.ColorYellow,
		"Blue":     tcell.ColorBlue,
		"Magenta":  tcell.ColorDarkMagenta,
		"Cyan":     tcell.ColorDarkCyan,
		"White":    tcell.ColorWhite,
		"Gray":     tcell.ColorGray,
		"DarkGray": tcell.ColorDarkGray,
		"Silver":   tcell.ColorSilver,
		"Maroon":   tcell.ColorMaroon,
		"Olive":    tcell.ColorOlive,
		"Lime":     tcell.ColorLime,
		"Aqua":     tcell.ColorAqua,
		"Teal":     tcell.ColorTeal,
		"Navy":     tcell.ColorNavy,
		"Fuchsia":  tcell.ColorFuchsia,
		"Purple":   tcell.ColorPurple,
		"Orange":   tcell.ColorOrange,
		"Default":  tcell.ColorDefault,
	}

	if color, exists := colors[colorName]; exists {
		return color
	}
	return tcell.ColorDefault
}

func (e *EditorCore) updateStylesHelper(currentPos int, selectedColor tcell.Color) {
	settings := e.colorSettingsTable()
	if currentPos < 0 || currentPos >= len(settings) {
		return
	}

	setting := settings[currentPos]
	if setting.isBackground {
		*setting.style = setting.style.Background(selectedColor)
	} else {
		*setting.style = setting.style.Foreground(selectedColor)
	}
}

func (e *EditorCore) loopChangeSettings() {

	colors := map[string]tcell.Color{
		// Basic colors
		"Black":   tcell.ColorBlack,
		"Red":     tcell.ColorRed,
		"Green":   tcell.ColorGreen,
		"Yellow":  tcell.ColorYellow,
		"Blue":    tcell.ColorBlue,
		"Magenta": tcell.ColorDarkMagenta,
		"Cyan":    tcell.ColorDarkCyan,
		"White":   tcell.ColorWhite,

		// Extended colors
		"Gray":     tcell.ColorGray,
		"DarkGray": tcell.ColorDarkGray,
		"Silver":   tcell.ColorSilver,
		"Maroon":   tcell.ColorMaroon,
		"Olive":    tcell.ColorOlive,
		"Lime":     tcell.ColorLime,
		"Aqua":     tcell.ColorAqua,
		"Teal":     tcell.ColorTeal,
		"Navy":     tcell.ColorNavy,
		"Fuchsia":  tcell.ColorFuchsia,
		"Purple":   tcell.ColorPurple,
		"Orange":   tcell.ColorOrange,
		"Default":  tcell.ColorDefault,
	}

	colorNames := []string{
		"Black", "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White",
		"Gray", "DarkGray", "Silver", "Maroon", "Olive", "Lime", "Aqua", "Teal",
		"Navy", "Fuchsia", "Purple", "Orange", "Default",
	}

	exampleOffset := 5
	e.Terminal.Clear()
	currentPos := 0
	colorPos := 0

	// Initialize colorPos to match the current setting's color
	colorPos = e.getCurrentColorPos(currentPos, colorNames)

	e.DisplaySettingsLoop(currentPos)
	e.DisplayColorsLoop(exampleOffset)
	e.Terminal.Show()

	for {
		event := e.Terminal.PollEvent()
		switch ev := event.(type) {

		case *tcell.EventResize:
			e.Terminal.Sync()

		case *tcell.EventKey:
			mod, key := ev.Modifiers(), ev.Key()
			if mod == tcell.ModNone {
				switch key {
				case tcell.KeyEnter:
					{
						// Apply selected color to current style setting.
						// colorPos is -1 when the stored color didn't match any
						// preset and hasn't been changed via Left/Right - leave
						// it untouched instead of overwriting it with colorNames[0].
						if currentPos < e.SettingsLength {
							if colorPos >= 0 {
								selectedColor := colors[colorNames[colorPos]]
								e.updateStylesHelper(currentPos, selectedColor)
							}
							currentSettings := e.GetCurrentSettings()
							err := SaveSettings(currentSettings)
							if err != nil {
								e.PrintMessageStyle(e.Cols/2, e.Rows/2, e.Styles.Error, "An error happened when saving the settings, please try again")
							}

						}
						e.Terminal.Clear()
						return
					}
				case tcell.KeyEsc:
					{
						return
					}
				case tcell.KeyUp:
					{
						currentPos--
						if currentPos < 0 {
							currentPos = 0
						}

						// Update colorPos to match the current setting's color
						if currentPos < e.SettingsLength {
							colorPos = e.getCurrentColorPos(currentPos, colorNames)
						}
					}
				case tcell.KeyDown:
					{
						currentPos++
						if currentPos == e.SettingsLength {
							currentPos = e.SettingsLength - 1
						}

						// Update colorPos to match the current setting's color
						if currentPos < e.SettingsLength {
							colorPos = e.getCurrentColorPos(currentPos, colorNames)
						}
					}
				case tcell.KeyLeft:
					{
						// Navigate through colors
						colorPos--
						if colorPos < 0 {
							colorPos = len(colorNames) - 1
						}

						// Apply selected color immediately for preview
						if currentPos < e.SettingsLength {
							selectedColor := colors[colorNames[colorPos]]
							e.updateStylesHelper(currentPos, selectedColor)
						}
					}
				case tcell.KeyRight:
					{
						// Navigate through colors
						colorPos++
						if colorPos >= len(colorNames) {
							colorPos = 0
						}

						// Apply selected color immediately for preview
						if currentPos < e.SettingsLength {
							selectedColor := colors[colorNames[colorPos]]
							e.updateStylesHelper(currentPos, selectedColor)
						}
					}
				default:
				}
			} else if mod == tcell.ModCtrl {

			} else if mod == tcell.ModAlt {
			}

		}

		// Recompute Cols/Rows every iteration so a terminal resize while in
		// settings mode takes effect immediately instead of only after
		// returning to command mode.
		e.updateDimensions()

		e.Terminal.Clear()

		// Save the updated settings
		currentSettings := e.GetCurrentSettings()
		err := SaveSettings(currentSettings)
		if err != nil {
			// Show error message
			e.PrintMessageStyle(0, e.Rows-1, e.Styles.Error, "Error saving settings")
			e.Terminal.Show()
			e.Terminal.PollEvent() // Wait for user input
		}

		// Pass current color selection to display functions for live preview
		e.DisplaySettingsLoop(currentPos)
		e.DisplayColorsLoop(exampleOffset)

		e.Terminal.Show()
	}

}
