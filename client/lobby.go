package client

import (
	"BomboweStatki/board"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

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

	newShipLayout := []string{}
	shipTypes := []int{4, 3, 3, 2, 2, 2, 1, 1, 1, 1}

	go editBoard(ui, opponentBoard, opponentStates, newShipLayout, shipTypes)

	ui.Start(context.TODO(), nil)

}

func IsAdjacentShip(char string, ship []string, mode int) (bool, error) {
	if len(ship) == 0 {
		return false, nil
	}

	if len(char) == 0 {
		return false, errors.New("invalid coordinate: empty string")
	}

	x := int(char[0] - 'A')
	y, err := strconv.Atoi(char[1:])
	if err != nil {
		return false, fmt.Errorf("error converting row: %v", err)
	}

	for _, c := range ship {
		if len(c) < 2 {
			return false, fmt.Errorf("invalid ship coordinate: %s", c)
		}
		cx := int(c[0] - 'A')
		cy, err := strconv.Atoi(c[1:])
		if err != nil {
			return false, fmt.Errorf("error converting ship coordinate: %v", err)
		}

		if mode == 1 && ((cx == x && (cy == y+1 || cy == y-1)) || (cy == y && (cx == x+1 || cx == x-1))) {
			return true, nil
		}
		if mode == 2 && (abs(cx-x) <= 1 && abs(cy-y) <= 1) {
			return true, nil
		}
	}
	return false, nil
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func DisplayStats(stats []PlayerStats) {
	for i, player := range stats {
		if i >= 10 {
			break
		}
		fmt.Printf("Rank: %d, Nick: %s, Points: %d, Wins: %d, Games: %d\n", player.Rank, player.Nick, player.Points, player.Wins, player.Games)
	}
}
