package client

import (
	board "BomboweStatki/board"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
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

func waitForChallenger(ui *gui.GUI, playerToken string, gameData GameInitData, cancel context.CancelFunc) {
	isWaitingForChallenger = true
	gameStarted := false
	userInLobby := false
	for !gameStarted || !userInLobby {
		lobbyInfo, _, err := retryOnErrorWithPlayers(ui, func() ([]Player, string, error) {
			return GetLobbyInfo()
		})
		if err != nil {
			ui.Draw(gui.NewText(2, 0, "Error getting lobby info: "+err.Error(), errorText))
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
				ui.Draw(gui.NewText(2, 0, "Error getting game status: "+err.Error(), errorText))
			}

			var gameStatus GameStatusResponse
			err = json.Unmarshal([]byte(gameStatusResponse), &gameStatus)
			if err != nil {
				ui.Draw(gui.NewText(2, 0, "Error unmarshaling response: "+err.Error(), errorText))
			}

			if gameStatus.GameStatus == "game_in_progress" {
				cancel()
				ui.NewScreen("game" + playerToken)
				ui.SetScreen("game" + playerToken)
				waitForStart(gameData.Nick, ui)
				LaunchGameBoard(ui, playerToken, gameData)
				isWaitingForChallenger = false
				return
			} else if gameStatus.GameStatus == "" {
				cancel()
				go pvpMenu(ui, "", nil, nil, make(chan bool))
				isWaitingForChallenger = false
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	isWaitingForChallenger = false
}

func changeShipLayout(ui *gui.GUI) error {
	ui.NewScreen("game")
	ui.SetScreen("game")

	playerShipCoordinates := DefaultGameInitData.Coords

	playerStates, opponentStates, _, err := board.Config(playerShipCoordinates)
	if err != nil {
		ui.Draw(gui.NewText(1, 28, "Error launching the board: "+err.Error(), errorText))
		return err
	}

	_, opponentBoard, buttonArea := board.GuiInit(ui, playerStates, opponentStates)

	newShipLayout := []string{}
	shipTypes := []int{4, 3, 3, 2, 2, 2, 1, 1, 1, 1}
	go editBoard(ui, opponentBoard, opponentStates, newShipLayout, shipTypes, buttonArea)

	return errors.New("finished editing board")
}

// printTopPlayers prints the top 10 players on the UI, split into two columns
func printTopPlayers(ui *gui.GUI, x, y int) {
	// Sort players by points (descending order)
	var players []PlayerStats
	var err error
	for i := 0; i < 10; i++ {
		players, err = GetStats()
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err != nil {
		ui.Draw(gui.NewText(0, 0, "Error: "+err.Error(), errorText))
		return
	}

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
		textConfig.FgColor = gui.Yellow
		textConfig.BgColor = gui.Black
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

func StartGame(ui *gui.GUI, gameData GameInitData) error {

	playerToken, err := retryOnError(ui, func() (string, error) {
		return InitGame(gameData)
	})
	if err != nil {
		ui.Draw(gui.NewText(1, 29, "Error initializing game: "+err.Error()+". Retrying in 2 seconds...", errorText))
	}

	ui.NewScreen("game" + playerToken)
	ui.SetScreen("game" + playerToken)
	err = waitForStart(gameData.Nick, ui)
	if err != nil {
		ui.Draw(gui.NewText(1, 28, "Error waiting for game to start: "+err.Error(), errorText))
	}

	err = LaunchGameBoard(ui, playerToken, gameData)
	if err != nil {
		ui.Draw(gui.NewText(1, 28, err.Error(), errorText))
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
			ui.Draw(gui.NewText(2, 0, "Error getting lobby info: "+err.Error(), errorText))
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

func printPlayerStats(ui *gui.GUI, nick string, x, y int) {
	playerStats, err := retryOnErrorWithPlayerStats(ui, func() (PlayerStats, error) {
		return GetPlayerStats(nick)
	})
	if err != nil {
		ui.Draw(gui.NewText(2, 0, "Error getting player stats: "+err.Error(), errorText))
		return
	}

	ui.Draw(gui.NewText(x, y, "Current Player Stats:", defaultText)) // Header
	ui.Draw(gui.NewText(x, y+1, "Player Nick: "+playerStats.Nick, defaultText))
	ui.Draw(gui.NewText(x, y+2, "Games Played: "+strconv.Itoa(playerStats.Games), defaultText))
	ui.Draw(gui.NewText(x, y+3, "Points: "+strconv.Itoa(playerStats.Points), defaultText))
	ui.Draw(gui.NewText(x, y+4, "Rank: "+strconv.Itoa(playerStats.Rank), defaultText))
	ui.Draw(gui.NewText(x, y+5, "Wins: "+strconv.Itoa(playerStats.Wins), defaultText))
}

func drawNameField(ui *gui.GUI, ctx context.Context) string {
	buttonConfig := gui.NewButtonConfig()
	buttonConfig.Height = 3
	buttonConfig.Width = 9
	buttonConfig.BgColor = gui.Green
	nametext := gui.NewText(2, 2, "Enter your new username here", defaultText)
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
	desctext := gui.NewText(2, 2, "Enter your new description here", defaultText)
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

var remaining int
var mu sync.Mutex

func lobbyTimer(ctx context.Context, reset chan bool, ui *gui.GUI) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	remaining = 15

	for {
		select {
		case <-ctx.Done():
			return
		case <-reset:
			remaining = 15
		case <-ticker.C:
			mu.Lock()
			remaining--
			mu.Unlock()
			ui.Draw(gui.NewText(2, 4, fmt.Sprintf("Time remaining: %d seconds", remaining), defaultText))

			if remaining == 0 {
				return
			}
		}
	}
}
