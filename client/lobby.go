package client

import (
	board "BomboweStatki/board"
	"context"
	"fmt"
	"time"
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

	playerToken, shipCoords, err := InitGame(username, desc, "")
	if err != nil {
		fmt.Println("Error initializing game:", err)
		return
	}

	fmt.Println("Player added to lobby, waiting for an opponent...")
	fmt.Println("Press Enter to exit...")

	gameStarted := false
	for !gameStarted {
		lobbyInfo, _, err := GetLobbyInfo()
		if err != nil {
			// fmt.Println("Error getting lobby info:", err)
			return
		}

		userInLobby := false
		for _, player := range lobbyInfo {
			if player.Nick == username {
				userInLobby = true
				if player.GameStatus != "waiting" {
					fmt.Println("Game started!")
					gameStarted = true
					break
				}
			}
		}

		if !userInLobby {
			fmt.Println("User not in lobby, refreshing...")
			break
		}
		refreshLobbyLoop(playerToken)
		time.Sleep(1 * time.Second)
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
	for {
		err := RefreshLobby(playerToken)
		if err != nil {
			fmt.Println("Error refreshing lobby:", err)
			return
		}
		fmt.Println("Refreshed lobby")

		time.Sleep(2 * time.Second)
	}
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

	playerToken, shipCoords, err := InitGame(username, desc, opponentUsername)
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

func DisplayStats(stats []PlayerStats) {
	for i, player := range stats {
		if i >= 10 {
			break
		}
		fmt.Printf("Rank: %d, Nick: %s, Points: %d, Wins: %d, Games: %d\n", player.Rank, player.Nick, player.Points, player.Wins, player.Games)
	}
}
