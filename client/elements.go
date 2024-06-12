package client

import (
	board "BomboweStatki/board"

	gui "github.com/s25867/warships-gui/v2"
)

// Global text configuration
var defaultText = gui.NewTextConfig()

func init() {
	defaultText.FgColor = gui.White
	defaultText.BgColor = gui.Black
}

var errorText = gui.NewTextConfig()

func init() {
	errorText.FgColor = gui.Red
	errorText.BgColor = gui.Black
}

func MainMenuElements(ui *gui.GUI) *MenuUI {
	sectionText := gui.NewText(2, 1, "Bombowe Statki", defaultText)
	textConfig := gui.NewTextConfig()
	textConfig.FgColor = gui.Yellow
	textConfig.BgColor = gui.Black
	hofText := gui.NewText(30, 1, "Hall of Fame", textConfig)
	// Action Buttons
	buttonConfig := gui.NewButtonConfig()
	buttonConfig.Height = 3
	buttonConfig.Width = 9
	buttonConfig.BgColor = gui.Green
	buttonConfig.FgColor = gui.Black
	pvpButtton := gui.NewButton(2, 5, "PvP", buttonConfig)
	buttonConfig.BgColor = gui.Blue
	botButtton := gui.NewButton(2, 9, "Bot", buttonConfig)
	buttonConfig.BgColor = gui.Red
	profileButton := gui.NewButton(2, 13, "Profile", buttonConfig)

	// Handle Area for buttons
	buttonMapping := map[string]gui.Spatial{
		"botButtton":    botButtton,
		"pvpButtton":    pvpButtton,
		"profileButton": profileButton,
	}
	buttonArea := gui.NewHandleArea(buttonMapping)

	// Draw all objects
	drawables := []gui.Drawable{
		sectionText,
		hofText,
		buttonArea,
		pvpButtton,
		botButtton,
		profileButton,
	}
	for _, drawable := range drawables {
		ui.Draw(drawable)
	}

	return &MenuUI{
		Ui:            ui,
		pvpButtton:    pvpButtton,
		botButton:     botButtton,
		refreshButton: profileButton,
		ButtonArea:    buttonArea,
	}
}

type botMenuUI struct {
	Ui         *gui.GUI
	ButtonArea *gui.HandleArea
	Drawable   []gui.Drawable
}
type ProfileUI struct {
	Ui               *gui.GUI
	returnButton     *gui.Button
	ButtonArea       *gui.HandleArea
	UsernameField    *gui.TextInput
	DescriptionField *gui.TextInput
	EditNameButton   *gui.Button
	EditDescButton   *gui.Button
}

func ProfileElements(ui *gui.GUI) *ProfileUI {
	sectionText := gui.NewText(2, 1, "Edit your profile", defaultText)
	currentName := gui.NewText(80, 5, "Current username: "+DefaultGameInitData.Nick, defaultText)
	currentDesc := gui.NewText(80, 6, "Current description: "+DefaultGameInitData.Desc, defaultText)

	// Action Buttons
	buttonConfig := gui.NewButtonConfig()
	buttonConfig.Height = 3
	buttonConfig.Width = 20
	buttonConfig.FgColor = gui.Black
	buttonConfig.BgColor = gui.Red
	returnButton := gui.NewButton(2, 9, "Return", buttonConfig)
	_, h := returnButton.Size()
	buttonConfig.BgColor = gui.Green
	editNameButton := gui.NewButton(2, h+10, "Edit Name", buttonConfig)
	_, h = editNameButton.Size()
	editDescButton := gui.NewButton(2, 2*h+11, "Edit Description", buttonConfig)
	_, h = editDescButton.Size()
	buttonConfig.BgColor = gui.Blue
	editBoardButton := gui.NewButton(2, 3*+h+12, "Edit Board Layout", buttonConfig)
	_, h = editBoardButton.Size()
	randomBoardButton := gui.NewButton(2, 4*h+13, "Get Random Board", buttonConfig)

	//board
	boardText := gui.NewText(32, 3, "Your current board layout", defaultText)
	boardStates, _, _, _ := board.Config(DefaultGameInitData.Coords)
	boardConfig := gui.NewBoardConfig()
	boardLayout := gui.NewBoard(28, 5, boardConfig)
	boardLayout.SetStates(boardStates)

	// Handle Area for buttons
	buttonMapping := map[string]gui.Spatial{
		"returnButton":      returnButton,
		"boardButton":       editBoardButton,
		"editNameButton":    editNameButton,
		"editDescButton":    editDescButton,
		"randomBoardButton": randomBoardButton,
	}
	buttonArea := gui.NewHandleArea(buttonMapping)

	// Draw all objects
	drawables := []gui.Drawable{
		boardText,
		sectionText,
		currentName,
		currentDesc,
		buttonArea,
		editBoardButton,
		returnButton,
		editNameButton,
		editDescButton,
		randomBoardButton,
		boardLayout,
	}
	for _, drawable := range drawables {
		ui.Draw(drawable)
	}

	return &ProfileUI{
		Ui:             ui,
		returnButton:   returnButton,
		ButtonArea:     buttonArea,
		EditNameButton: editNameButton,
		EditDescButton: editDescButton,
	}
}

type LobbyUI struct {
	Ui         *gui.GUI
	ButtonArea *gui.HandleArea
	Drawable   []gui.Drawable
}

func LobbyElements(ui *gui.GUI, lobbyInfo []Player) *LobbyUI {
	sectionText := gui.NewText(2, 1, "Lobby", defaultText)

	// Action Buttons
	buttonConfig := gui.NewButtonConfig()
	buttonConfig.Height = 3
	buttonConfig.Width = 9
	buttonConfig.FgColor = gui.Black
	buttonConfig.BgColor = gui.Green
	refreshButton := gui.NewButton(2, 5, "Refresh", buttonConfig)
	x, _ := refreshButton.Position()
	w, _ := refreshButton.Size()
	buttonConfig.BgColor = gui.Red
	returnButton := gui.NewButton(x+w+2, 5, "Return", buttonConfig)
	x, _ = returnButton.Position()
	w, _ = returnButton.Size()
	buttonConfig.BgColor = gui.Blue
	buttonConfig.Width = 12
	addYourselfButton := gui.NewButton(x+w+2, 5, "Join lobby", buttonConfig)
	buttonConfig.BgColor = gui.Green
	x, _ = addYourselfButton.Position()
	w, _ = addYourselfButton.Size()
	resetLobbyTimerButton := gui.NewButton(x+w+2, 5, "Reset Timer", buttonConfig)

	// Check if lobby is empty
	if len(lobbyInfo) == 0 {
		ui.Draw(gui.NewText(2, 11, "Lobby is empty", defaultText))
	}

	// Handle Area for buttons
	buttonMapping := map[string]gui.Spatial{
		"refreshButton":         refreshButton,
		"returnButton":          returnButton,
		"addYourselfButton":     addYourselfButton,
		"resetLobbyTimerButton": resetLobbyTimerButton,
	}

	x, y := 2, 11
	for _, player := range lobbyInfo {
		if player.GameStatus == "waiting" {
			buttonConfig.Width = max(len(player.Nick)+2, 9)
			opponentButton := gui.NewButton(x, y, player.Nick, buttonConfig)

			// Draw the player's nick
			nickText := gui.NewText(x, y, player.Nick, defaultText)
			ui.Draw(nickText)

			// Add the opponent button to the buttonMapping
			buttonMapping[player.Nick] = opponentButton

			// Update x and y for the next button
			x += buttonConfig.Width + 2
			if (x + buttonConfig.Width) > 80 {
				x = 2
				y += 4
			}
		}
	}

	// Update Handle Area to include all buttons
	buttonArea := gui.NewHandleArea(buttonMapping)

	// Draw all objects
	drawables := []gui.Drawable{
		sectionText,
		buttonArea,
		refreshButton,
		returnButton,
		addYourselfButton,
		resetLobbyTimerButton,
	}

	for _, player := range lobbyInfo {
		if player.GameStatus == "waiting" {
			opponentButton := buttonMapping[player.Nick].(gui.Drawable)
			drawables = append(drawables, opponentButton)
		}
	}

	for _, drawable := range drawables {
		ui.Draw(drawable)
	}

	return &LobbyUI{
		Drawable:   drawables,
		Ui:         ui,
		ButtonArea: buttonArea,
	}
}

type MenuUI struct {
	Ui            *gui.GUI
	pvpButtton    *gui.Button
	botButton     *gui.Button
	refreshButton *gui.Button
	ButtonArea    *gui.HandleArea
}

func BotElements(ui *gui.GUI) *botMenuUI {

	sectionText := gui.NewText(2, 2, "Pick a bot you want to figth against!", defaultText)

	buttonConfig := gui.NewButtonConfig()
	buttonConfig.Height = 3
	buttonConfig.Width = 10
	buttonConfig.FgColor = gui.Black
	buttonConfig.BgColor = gui.Green
	buttonConfig.WithBorder = true
	wpBotButton := gui.NewButton(14, 5, "wpBot", buttonConfig)
	buttonConfig.BgColor = gui.Blue
	bomBotButton := gui.NewButton(14, 9, "bomBot", buttonConfig)
	buttonConfig.BgColor = gui.Red
	returnButton := gui.NewButton(14, 13, "Return", buttonConfig)

	buttonMapping := map[string]gui.Spatial{
		"wpBotButton":  wpBotButton,
		"bomBotButton": bomBotButton,
		"returnButton": returnButton,
	}

	buttonArea := gui.NewHandleArea(buttonMapping)

	drawables := []gui.Drawable{
		sectionText,
		buttonArea,
		wpBotButton,
		bomBotButton,
		returnButton,
	}

	for _, drawable := range drawables {
		ui.Draw(drawable)
	}

	return &botMenuUI{
		Drawable:   drawables,
		Ui:         ui,
		ButtonArea: buttonArea,
	}
}
