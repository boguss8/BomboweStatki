package client

import (
	board "BomboweStatki/board"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	gui "github.com/s25867/warships-gui/v2"
)

type GameStatusResponse struct {
	GameStatus     string   `json:"game_status"`
	LastGameStatus string   `json:"last_game_status"`
	Nick           string   `json:"nick"`
	OppShots       []string `json:"opp_shots"`
	Opponent       string   `json:"opponent"`
	ShouldFire     bool     `json:"should_fire"`
}

func waitForChallenger(ui *gui.GUI, playerToken string, gameData GameInitData) {
	gameStarted := false
	userInLobby := false
	for !gameStarted || !userInLobby {
		lobbyInfo, _, err := retryOnErrorWithPlayers(ui, func() ([]Player, string, error) {
			return GetLobbyInfo()
		})
		if err != nil {
			ui.Draw(gui.NewText(2, 0, "Error getting lobby info: "+err.Error(), nil))
		}

		userInLobby = false
		for _, player := range lobbyInfo {
			if player.Nick == DefaultGameInitData.Nick {
				userInLobby = true
				if player.GameStatus == "waiting" {
					gameStarted = false
					time.Sleep(200 * time.Millisecond)
					break
				}
			}
		}

		if !userInLobby {
			gameStatusResponse, err := retryOnError(ui, func() (string, error) {
				return GetGameStatus(playerToken)
			})
			if err != nil {
				ui.Draw(gui.NewText(2, 0, "Error getting game status: "+err.Error(), nil))
				return
			}

			var gameStatus GameStatusResponse
			err = json.Unmarshal([]byte(gameStatusResponse), &gameStatus)
			if err != nil {
				ui.Draw(gui.NewText(2, 0, "Error unmarshaling response: "+err.Error(), nil))

				return
			}

			if gameStatus.GameStatus == "game_in_progress" {
				ui.NewScreen("game" + playerToken)
				ui.SetScreen("game" + playerToken)
				waitForStart(gameData.Nick, ui)
				LaunchGameBoard(ui, playerToken, gameData)
				return
			} else {
				lobbyInfo, _, err := retryOnErrorWithPlayers(ui, func() ([]Player, string, error) {
					return GetLobbyInfo()
				})
				if err != nil {
					ui.Draw(gui.NewText(2, 0, "Error getting lobby info: "+err.Error(), nil))
				}

				LobbyElements(ui, lobbyInfo)
				go pvpLobby(ui, "")
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func changeShipLayout(ui *gui.GUI) error {
	ui.NewScreen("game")
	ui.SetScreen("game")

	playerShipCoordinates := DefaultGameInitData.Coords

	playerStates, opponentStates, _, err := board.Config(playerShipCoordinates)
	if err != nil {
		ui.Draw(gui.NewText(1, 28, "Error launching the board: "+err.Error(), nil))
		return err
	}

	_, opponentBoard, buttonArea := board.GuiInit(ui, playerStates, opponentStates)

	newShipLayout := []string{}
	shipTypes := []int{4, 3, 3, 2, 2, 2, 1, 1, 1, 1}
	go editBoard(ui, opponentBoard, opponentStates, newShipLayout, shipTypes, buttonArea)

	return errors.New("finished editing board")
}

type MenuUI struct {
	Ui            *gui.GUI
	pvpButtton    *gui.Button
	botButton     *gui.Button
	refreshButton *gui.Button
	ButtonArea    *gui.HandleArea
}

// printTopPlayers prints the top 10 players on the UI, split into two columns
func printTopPlayers(ui *gui.GUI, players []PlayerStats, x, y int) {
	// Sort players by points (descending order)
	sort.Slice(players, func(i, j int) bool {
		return players[i].Points > players[j].Points
	})

	// Calculate the maximum name length
	maxNameLength := 0
	for _, player := range players {
		if len(player.Nick) > maxNameLength {
			maxNameLength = len(player.Nick)
		}
	}

	// Calculate the width of each column based on the maximum name length
	columnWidth := maxNameLength + 15 // Adjust as needed for additional spacing

	// Print players in two columns
	for i, player := range players {
		textConfig := gui.NewTextConfig()
		textConfig.FgColor = gui.White
		textConfig.BgColor = gui.Yellow
		column := i % 2
		row := i / 2
		columnX := x + (column * (columnWidth + 5)) // Adjust spacing between columns
		ui.Draw(gui.NewText(columnX, y+row*6, fmt.Sprintf("Rank: %d", player.Rank), textConfig))
		ui.Draw(gui.NewText(columnX, y+1+row*6, "Player Nick: "+player.Nick, textConfig))
		ui.Draw(gui.NewText(columnX, y+2+row*6, "Games Played: "+strconv.Itoa(player.Games), textConfig))
		ui.Draw(gui.NewText(columnX, y+3+row*6, "Points: "+strconv.Itoa(player.Points), textConfig))
		ui.Draw(gui.NewText(columnX, y+4+row*6, "Wins: "+strconv.Itoa(player.Wins), textConfig))
	}
}

func MainMenu(ui *gui.GUI) {
	ui.NewScreen("menu")
	ui.SetScreen("menu")

	menuUi := MainMenuElements(ui)

	ctx := context.Background()
	var players []PlayerStats
	var err error
	for i := 0; i < 10; i++ {
		players, err = GetStats()
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		ui.Draw(gui.NewText(0, 0, "Error: "+err.Error(), nil))
		return
	}

	printTopPlayers(ui, players, 30, 3)

	// Handle button clicks
	for {
		clicked := menuUi.ButtonArea.Listen(ctx)
		switch clicked {
		case "pvpButtton":
			go pvpLobby(ui, "")
			return
		case "botButtton":
			// DefaultGameInitData.Wpbot = true
			// StartGame(ui, DefaultGameInitData)
			go botMenu(ui)
			return
		case "profileButton":
			go Profile(ui)
		}
	}
}

func MainMenuElements(ui *gui.GUI) *MenuUI {
	sectionText := gui.NewText(2, 1, "Bombowe Statki", nil)
	textConfig := gui.NewTextConfig()
	textConfig.FgColor = gui.White
	textConfig.BgColor = gui.Yellow
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

func StartGame(ui *gui.GUI, gameData GameInitData) error {

	playerToken, err := retryOnError(ui, func() (string, error) {
		return InitGame(gameData)
	})
	if err != nil {
		ui.Draw(gui.NewText(1, 29, "Error initializing game: "+err.Error()+". Retrying in 2 seconds...", nil))
	}

	ui.NewScreen("game" + playerToken)
	ui.SetScreen("game" + playerToken)
	err = waitForStart(gameData.Nick, ui)
	if err != nil {
		ui.Draw(gui.NewText(1, 28, "Error waiting for game to start: "+err.Error(), nil))
	}

	err = LaunchGameBoard(ui, playerToken, gameData)
	if err != nil {
		ui.Draw(gui.NewText(1, 28, err.Error(), nil))
	}
	return errors.New("game ended")
}

func LaunchGameBoard(ui *gui.GUI, playerToken string, gameData GameInitData) error {
	// Configure the board
	playerStates, opponentStates, shipStatus, err := board.Config(DefaultGameInitData.Coords)

	if err != nil {
		return fmt.Errorf("error launching the board: %v", err)
	}

	// Initialize the GUI for the board
	playerBoard, opponentBoard, buttonArea := board.GuiInit(ui, playerStates, opponentStates)

	dataCoords, err := GetBoardInfoWithRetry(playerToken)
	if err != nil {
		return fmt.Errorf("error getting board info: %v", err)
	}

	// Start operations on the player and opponent boards
	ctx, cancel := context.WithCancel(context.Background())
	go displayGameStatus(ctx, playerToken, ui, cancel)
	go opponentBoardOperations(ctx, playerToken, opponentBoard, opponentStates, ui, buttonArea)

	go playerBoardOperations(ctx, playerToken, playerBoard, playerStates, ui, shipStatus, dataCoords)

	return nil
}

func waitForStart(username string, ui *gui.GUI) error {
	gameStarted := false
	for !gameStarted {
		lobbyInfo, _, err := retryOnErrorWithPlayers(ui, func() ([]Player, string, error) {
			return GetLobbyInfo()
		})
		if err != nil {
			ui.Draw(gui.NewText(2, 0, "Error getting lobby info: "+err.Error(), nil))
		}

		userInLobby := false
		for _, player := range lobbyInfo {
			if player.Nick == username {
				userInLobby = true
				if player.GameStatus == "waiting" {
					gameStarted = false
					time.Sleep(1 * time.Second)
					break
				}
			}
		}

		if !userInLobby {
			return nil
		}
	}

	if !gameStarted {
		return nil
	}

	return nil
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

func printPlayerStats(ui *gui.GUI, nick string, x, y int) {
	playerStats, err := retryOnErrorWithPlayerStats(ui, func() (PlayerStats, error) {
		return GetPlayerStats(nick)
	})
	if err != nil {
		ui.Draw(gui.NewText(2, 0, "Error getting player stats: "+err.Error(), nil))
		return
	}

	ui.Draw(gui.NewText(x, y, "Current Player Stats:", nil)) // Header
	ui.Draw(gui.NewText(x, y+1, "Player Nick: "+playerStats.Nick, nil))
	ui.Draw(gui.NewText(x, y+2, "Games Played: "+strconv.Itoa(playerStats.Games), nil))
	ui.Draw(gui.NewText(x, y+3, "Points: "+strconv.Itoa(playerStats.Points), nil))
	ui.Draw(gui.NewText(x, y+4, "Rank: "+strconv.Itoa(playerStats.Rank), nil))
	ui.Draw(gui.NewText(x, y+5, "Wins: "+strconv.Itoa(playerStats.Wins), nil))
}

func Profile(ui *gui.GUI) {
	ui.NewScreen("profile")
	ui.SetScreen("profile")

	profileUi := ProfileElements(ui)

	ctx := context.Background()

	printPlayerStats(ui, DefaultGameInitData.Nick, 80, 11)

	// Handle button clicks
	for {
		clicked := profileUi.ButtonArea.Listen(ctx)
		switch clicked {
		case "returnButton":

			go MainMenu(ui)
			return
		case "boardButton":
			go func() {
				changeShipLayout(ui)
			}()
			return
		case "editNameButton":
			name := drawNameField(ui, ctx)
			if name == "" {
				go Profile(ui)
				continue
			}
			DefaultGameInitData.Nick = name
			go Profile(ui)

		case "editDescButton":
			desc := drawDescField(ui, ctx)
			if desc == "" {
				go Profile(ui)
				continue
			}
			DefaultGameInitData.Desc = desc
			go Profile(ui)
		case "randomBoardButton":
			DefaultGameInitData.Coords = generateRandomBoard()
			go Profile(ui)
		}
	}
}

func ProfileElements(ui *gui.GUI) *ProfileUI {
	sectionText := gui.NewText(2, 1, "Edit your profile", nil)
	currentName := gui.NewText(80, 5, "Current username: "+DefaultGameInitData.Nick, nil)
	currentDesc := gui.NewText(80, 6, "Current description: "+DefaultGameInitData.Desc, nil)

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
	boardText := gui.NewText(32, 3, "Your current board layout", nil)
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

func drawNameField(ui *gui.GUI, ctx context.Context) string {
	buttonConfig := gui.NewButtonConfig()
	buttonConfig.Height = 3
	buttonConfig.Width = 9
	buttonConfig.BgColor = gui.Green
	nametext := gui.NewText(2, 3, "Enter your new username here", nil)
	usernameField := gui.NewTextInput(2, 4, 20)
	saveButton := gui.NewButton(2, 5, "Save", buttonConfig)
	x, _ := saveButton.Position()
	w, _ := saveButton.Size()
	buttonConfig.BgColor = gui.Red
	cancelButton := gui.NewButton(x+w+2, 5, "Cancel", buttonConfig)
	buttonMapping := map[string]gui.Spatial{
		"saveButton":   saveButton,
		"cancelButton": cancelButton,
	}
	buttonArea := gui.NewHandleArea(buttonMapping)

	// Draw all objects
	drawables := []gui.Drawable{
		nametext,
		usernameField,
		buttonArea,
		cancelButton,
		saveButton,
	}
	for _, drawable := range drawables {
		ui.Draw(drawable)
	}

	for {
		clicked := buttonArea.Listen(ctx)
		switch clicked {
		case "saveButton":
			return usernameField.GetContent()
		case "cancelButton":
			return ""
		}
	}
}

func drawDescField(ui *gui.GUI, ctx context.Context) string {
	buttonConfig := gui.NewButtonConfig()
	buttonConfig.Height = 3
	buttonConfig.Width = 9
	buttonConfig.BgColor = gui.Green
	desctext := gui.NewText(2, 3, "Enter your new description here", nil)
	descriptionField := gui.NewTextInput(2, 4, 20)
	saveButton := gui.NewButton(2, 5, "Save", buttonConfig)
	x, _ := saveButton.Position()
	w, _ := saveButton.Size()
	buttonConfig.BgColor = gui.Red
	cancelButton := gui.NewButton(x+w+2, 5, "Cancel", buttonConfig)
	buttonMapping := map[string]gui.Spatial{
		"saveButton":   saveButton,
		"cancelButton": cancelButton,
	}
	buttonArea := gui.NewHandleArea(buttonMapping)

	// Draw all objects
	drawables := []gui.Drawable{
		desctext,
		descriptionField,
		buttonArea,
		cancelButton,
		saveButton,
	}
	for _, drawable := range drawables {
		ui.Draw(drawable)
	}

	for {
		clicked := buttonArea.Listen(ctx)
		switch clicked {
		case "saveButton":
			return descriptionField.GetContent()
		case "cancelButton":
			return ""
		}
	}
}

type LobbyUI struct {
	Ui         *gui.GUI
	ButtonArea *gui.HandleArea
	Drawable   []gui.Drawable
}

func pvpLobby(ui *gui.GUI, playerToken string) {
	ui.NewScreen("lobby")
	ui.SetScreen("lobby")

	lobbyInfo, _, err := retryOnErrorWithPlayers(ui, func() ([]Player, string, error) {
		return GetLobbyInfo()
	})
	if err != nil {
		ui.Draw(gui.NewText(2, 0, "Error getting lobby info: "+err.Error(), nil))
	}

	lobbyUi := LobbyElements(ui, lobbyInfo)

	ctx := context.Background()
	ui.Draw(gui.NewText(2, 9, "Click on an opponent to challenge him into a duel!", nil))
	for {
		clicked := lobbyUi.ButtonArea.Listen(ctx)
		switch clicked {
		case "resetLobbyTimerButton":
			if playerToken == "" {
				ui.Draw(gui.NewText(0, 0, "You are not in lobby", nil))
			} else {
				RefreshLobby(playerToken)
				ui.Draw(gui.NewText(2, 3, "Lobby timer reset", nil))
			}
		case "addYourselfButton":
			gameData := GameInitData{
				TargetNick: "",
				Wpbot:      false,
			}
			playerToken, err := retryOnError(ui, func() (string, error) {
				return InitGame(gameData)
			})
			if err != nil {
				ui.Draw(gui.NewText(1, 29, "Error initializing game: "+err.Error()+". Retrying in 2 seconds...", nil))
			}
			go waitForChallenger(ui, playerToken, gameData)
			go pvpLobby(ui, playerToken)

		case "returnButton":
			go MainMenu(ui)
			return
		case "refreshButton":
			lobbyInfo, _, err := retryOnErrorWithPlayers(ui, func() ([]Player, string, error) {
				return GetLobbyInfo()
			})
			if err != nil {
				ui.Draw(gui.NewText(2, 0, "Error getting lobby info: "+err.Error(), nil))
			}
			lobbyUi = LobbyElements(ui, lobbyInfo)
			go pvpLobby(ui, "")
		default:
			for _, player := range lobbyInfo {
				if player.Nick == clicked {
					if DefaultGameInitData.Nick == player.Nick {
						ui.Draw(gui.NewText(2, 2, "You can't duel yourself...", nil))
						continue
					}
					go func() {
						DefaultGameInitData.TargetNick = player.Nick
						DefaultGameInitData.Wpbot = false
						StartGame(ui, DefaultGameInitData)
					}()
					break
				}
			}
		}
	}
}
func LobbyElements(ui *gui.GUI, lobbyInfo []Player) *LobbyUI {
	sectionText := gui.NewText(2, 1, "Lobby", nil)

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
		ui.Draw(gui.NewText(2, 11, "Lobby is empty", nil))
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
			nickText := gui.NewText(x, y, player.Nick, nil)
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

type botMenuUI struct {
	Ui         *gui.GUI
	ButtonArea *gui.HandleArea
	Drawable   []gui.Drawable
}

func botMenu(ui *gui.GUI) {
	ui.NewScreen("lobby")
	ui.SetScreen("lobby")

	botUi := BotElements(ui)

	ctx := context.Background()

	for {
		clicked := botUi.ButtonArea.Listen(ctx)
		switch clicked {
		case "returnButton":
			go MainMenu(ui)
			return
		case "wpBotButton":
			gameData := GameInitData{
				TargetNick: "",
				Wpbot:      true,
			}
			StartGame(ui, gameData)
		case "bomBotButton":
			bomBotInit(ui, DefaultGameInitData)
			return
		}
	}
}

func BotElements(ui *gui.GUI) *botMenuUI {

	sectionText := gui.NewText(2, 1, "Bot Lobby", nil)

	buttonConfig := gui.NewButtonConfig()
	buttonConfig.Height = 3
	buttonConfig.Width = 10
	buttonConfig.FgColor = gui.Black
	buttonConfig.BgColor = gui.Green
	wpBotButton := gui.NewButton(2, 5, "wpBot", buttonConfig)
	buttonConfig.BgColor = gui.Blue
	bomBotButton := gui.NewButton(2, 9, "bomBot", buttonConfig)
	buttonConfig.BgColor = gui.Red
	returnButton := gui.NewButton(2, 13, "Return", buttonConfig)

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
