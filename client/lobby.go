package client

import (
	board "BomboweStatki/board"
	game "BomboweStatki/game"
	"context"
	"fmt"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

// lobby
func AddToLobby() {
	var username string
	for {
		fmt.Print("Enter your username: ")
		_, err := fmt.Scanln(&username)
		if err != nil {
			fmt.Println("Invalid input. Please enter a valid username.")
			continue
		}
		break
	}

	var desc string
	for {
		fmt.Print("Enter your game description: ")
		_, err := fmt.Scanln(&desc)
		if err != nil {
			fmt.Println("Invalid input. Please enter a valid game description.")
			continue
		}
		break
	}

	var bot bool
	for {
		fmt.Print("Choose an option:\n1. Wait for a player\n2. Fight with a bot\nEnter your choice: ")
		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil || (choice != 1 && choice != 2) {
			fmt.Println("Invalid input. Please enter 1 to wait for a player or 2 to fight with a bot.")
			continue
		}
		if choice == 1 {
			bot = false
		} else {
			bot = true
		}
		break
	}
	gameData := GameInitData{
		Nick:       username,
		Desc:       desc,
		TargetNick: "",
		Wpbot:      bot,
	}
	playerToken, shipCoords, err := InitGame(gameData)
	if err != nil {
		fmt.Println("Error initializing game:", err)
		return
	}

	fmt.Println("Player added to lobby, waiting for an opponent...")

	gameStarted := false
	for !gameStarted {
		lobbyInfo, _, err := GetLobbyInfo()
		if err != nil {
			fmt.Println("Error getting lobby info:", err)
			return
		}

		userInLobby := false
		for _, player := range lobbyInfo {
			if player.Nick == username {
				userInLobby = true
				if player.GameStatus == "waiting" {
					fmt.Println("Waiting for an opponent...")
					gameStarted = false
					refreshLobbyLoop(playerToken)
					time.Sleep(1 * time.Second)
					break
				}
			}
		}

		if !userInLobby {
			fmt.Println("User not in lobby")
			gameStarted = true
			break
		}
	}

	if !gameStarted {
		return
	}

	// Launch game board
	playerStates, opponentStates, shipStatus, err := board.Config(playerToken, shipCoords)
	if err != nil {
		fmt.Println(err)
		return
	}

	ui, playerBoard, opponentBoard := board.GuiInit(playerStates, opponentStates)

	var shotCoordinates = make(map[string]bool)

	dataCoords, err := GetBoardInfo(playerToken)
	if err != nil {
		fmt.Println("Error getting board info:", err)
		return
	}

	go opponentBoardOperations(playerToken, opponentBoard, playerStates, opponentStates, shotCoordinates, ui)

	go playerBoardOperations(playerToken, playerBoard, opponentStates, playerStates, ui, shipStatus, dataCoords)

	ui.Start(context.TODO(), nil)

	fmt.Scanln()
}

func refreshLobbyLoop(playerToken string) {
	err := RefreshLobby(playerToken)
	if err != nil {
		fmt.Println("Error refreshing lobby:", err)
		return
	}
	fmt.Println("Refreshed lobby")

	time.Sleep(2 * time.Second)
}

func ChallengeOpponent() {
	for {
		fmt.Println("Choose an option:")
		fmt.Println("1. Refresh lobby (display current players)")
		fmt.Println("2. Challenge opponent by username")

		var choice int
		for {
			fmt.Print("Enter your choice: ")
			_, err := fmt.Scanln(&choice)
			if err != nil {
				fmt.Println("Error reading choice:", err)
				fmt.Println("Please enter 1 or 2.")
				continue
			}
			if choice != 1 && choice != 2 {
				fmt.Println("Invalid choice. Please enter 1 or 2.")
				continue
			}
			break
		}

		switch choice {
		case 1:
			isEmpty, err := DisplayLobbyStatus()
			if err != nil {
				fmt.Println("Error refreshing lobby:", err)
				return
			}
			fmt.Println("Lobby is empty:", isEmpty)

		case 2:
			fmt.Print("Enter opponent's username: ")
			var opponentUsername string
			_, err := fmt.Scanln(&opponentUsername)
			if err != nil {
				fmt.Println("Error reading opponent's username:", err)
				return
			}

			startGameWithOpponent(opponentUsername)
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func DisplayLobbyStatus() (bool, error) {
	lobbyInfo, rawResponse, err := GetLobbyInfo()
	if err != nil {
		return false, err
	}

	fmt.Println("Raw response:", rawResponse)

	if len(lobbyInfo) == 0 {
		return true, nil
	}

	for _, player := range lobbyInfo {
		fmt.Printf("Game Status: %s, Nick: %s\n", player.GameStatus, player.Nick)
	}

	return false, nil
}

func startGameWithOpponent(opponentUsername string) {
	fmt.Print("Enter your username: ")
	var username string
	_, err := fmt.Scanln(&username)
	if err != nil {
		fmt.Println("Error reading username:", err)
		return
	}

	fmt.Print("Enter your game description: ")
	var desc string
	_, err = fmt.Scanln(&desc)
	if err != nil {
		fmt.Println("Error reading game description:", err)
		return
	}
	gameData := GameInitData{
		Nick:       username,
		Desc:       desc,
		TargetNick: opponentUsername,
		Wpbot:      false,
	}
	playerToken, shipCoords, err := InitGame(gameData)
	if err != nil {
		fmt.Println("Error initializing game:", err)
		return
	}

	playerStates, opponentStates, shipStatus, err := board.Config(playerToken, shipCoords)
	if err != nil {
		fmt.Println(err)
		return
	}

	ui, playerBoard, opponentBoard := board.GuiInit(playerStates, opponentStates)

	var shotCoordinates = make(map[string]bool)

	dataCoords, err := GetBoardInfo(playerToken)
	if err != nil {
		fmt.Println("Error getting board info:", err)
		return
	}

	go opponentBoardOperations(playerToken, opponentBoard, playerStates, opponentStates, shotCoordinates, ui)

	go playerBoardOperations(playerToken, playerBoard, opponentStates, playerStates, ui, shipStatus, dataCoords)

	ui.Start(context.TODO(), nil)
}

func ChangeShipLayout() {
	playerShipCoordinates := DefaultGameInitData.Coords
	playerStates, _, _, _ := board.Config("", playerShipCoordinates)
	opponentStates := [10][10]gui.State{}
	ui, _, opponentBoard := board.GuiInit(playerStates, opponentStates)

	shipCoordinates := map[string]bool{}
	shipsPlaced := 0
	maxShips := 20

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ui.Start(ctx, nil)
	}()

	for shipsPlaced < maxShips {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		char := opponentBoard.Listen(ctx)
		if len(char) == 0 {
			fmt.Println("Error: received empty input")
			break
		}
		col := int(char[0] - 'A')
		var row int
		if len(char) == 3 {
			row = 9
		} else {
			row = int(char[1] - '1')
		}

		coord := fmt.Sprintf("%c%d", 'A'+col, row+1)

		if _, exists := shipCoordinates[coord]; exists {
			ui.Draw(gui.NewText(1, 1, "Position already occupied. Choose a different spot.", nil))
			continue
		}

		shipCoordinates[coord] = true
		opponentStates[col][row] = gui.Ship
		opponentBoard.SetStates(opponentStates)
		shipsPlaced++

		boardInfo := make([]string, 10)
		for i, row := range opponentStates {
			for _, state := range row {
				switch state {
				case gui.Hit:
					boardInfo[i] += "H"
				case gui.Miss:
					boardInfo[i] += "M"
				case gui.Ship:
					boardInfo[i] += "S"
				default:
					boardInfo[i] += " "
				}
			}
		}
		game.UpdateBoardStates(opponentBoard, boardInfo)

		ui.Draw(gui.NewText(1, 1, fmt.Sprintf("Ships placed: %d/%d", shipsPlaced, maxShips), nil))
	}
	ui.Draw(gui.NewText(1, 1, "New ship layout saved.", nil))

	DefaultGameInitData.Coords = make([]string, 0, len(shipCoordinates))
	for coord := range shipCoordinates {
		DefaultGameInitData.Coords = append(DefaultGameInitData.Coords, coord)
	}
}

func DisplayStats(stats []PlayerStats) {
	for i, player := range stats {
		if i >= 10 {
			break
		}
		fmt.Printf("Rank: %d, Nick: %s, Points: %d, Wins: %d, Games: %d\n", player.Rank, player.Nick, player.Points, player.Wins, player.Games)
	}
}
