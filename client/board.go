package client

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	gui "github.com/s25867/warships-gui/v2"
)

func generateRandomBoard() []string {
	shipSizes := []int{4, 3, 3, 2, 2, 2, 1, 1, 1, 1}
	board := make([]string, 0, 20)
	allCoords := make([]string, 0, 20)

	// Initialize all possible coordinates
	for i := 'A'; i <= 'J'; i++ {
		for j := 1; j <= 10; j++ {
			coord := fmt.Sprintf("%c%d", i, j)
			allCoords = append(allCoords, coord)
		}
	}

	// Place ships on the board
	for _, size := range shipSizes {
		var shipCoords []string
		for len(shipCoords) < size {
			// Randomly select a starting coordinate
			start := allCoords[rand.Intn(len(allCoords))]
			if len(shipCoords) > 0 {
				// Check if the starting coordinate is adjacent to the last coordinate
				adjacent, err := isAdjacentShip(start, shipCoords, 1)
				if err != nil || !adjacent {
					continue
				}
			}
			// Add the starting coordinate to the ship
			shipCoords = append(shipCoords, start)
			index := findIndex(allCoords, start)
			allCoords = append(allCoords[:index], allCoords[index+1:]...)
		}
		board = append(board, shipCoords...)
		for _, shipCoord := range shipCoords { // Mark surrounding area as missed
			for _, coord := range getSurroundingCoords(shipCoord) {
				index := findIndex(allCoords, coord)
				if index != -1 {
					allCoords = append(allCoords[:index], allCoords[index+1:]...)
				}
			}
		}
	}
	return board
}

func editBoard(ui *gui.GUI, opponentBoard *gui.Board, opponentStates [10][10]gui.State, newShipLayout []string, shipTypes []int, buttonArea *gui.HandleArea) {
	go func() {
		// Listen for exit button click
		ctx := context.Background()
		for {
			if clicked := buttonArea.Listen(ctx); clicked == "exitButton" {
				ui.Draw(gui.NewText(30, 24, "Exiting to main menu in 2 seconds...", errorText))
				time.Sleep(2 * time.Second)
				go MainMenu(ui)
				return
			}
		}
	}()

	for i := 0; i < len(shipTypes); i++ {

		ui.Draw(gui.NewText(1, 2, "                                  ", errorText))
		ui.Draw(gui.NewText(1, 1, fmt.Sprintf("Placing %x/10 ship of size %d", i+1, shipTypes[i]), nil))
		//init map
		ships := mapShips(newShipLayout)
		//fix ship status
		for idx, ship := range ships {
			if len(ship.Coords) == shipTypes[idx] { //jesli rozmiar statku nie zgadza sie z wymaganym
				ship.IsDestroyed = "false"
			} else {
				ship.IsDestroyed = "true"
			}
			ships[idx] = ship
		}

		for j := 0; j < shipTypes[i]; j++ {

			char := opponentBoard.Listen(context.Background())
			// Check if char is already part of any ship in newShipLayout
			alreadyShot := false
			for _, coord := range newShipLayout {
				if coord == char {
					alreadyShot = true
					break
				}
			}
			if alreadyShot {
				ui.Draw(gui.NewText(1, 2, "Coordinate already part of a ship!", errorText))
				j--
				continue
			}

			// Check if char is in the surrounding area of any completed ship
			isInSurroundingArea := false
			for _, ship := range ships {

				if ship.IsDestroyed == "false" {
					for _, coord := range ship.SurroundingArea {
						if coord == char {
							isInSurroundingArea = true
							break
						}
					}
				}
				if isInSurroundingArea {
					break
				}
			}
			if isInSurroundingArea {
				ui.Draw(gui.NewText(1, 2, "Coordinate is in the surrounding area of a completed ship!", errorText))
				j--
				continue
			}
			// Check if char is adjacent to any destroyed ship
			if !alreadyShot && !isInSurroundingArea {
				ships := mapShips(newShipLayout)

				for idx, ship := range ships {
					if len(ship.Coords) == shipTypes[idx] {
						ship.IsDestroyed = "false"
					} else {
						ship.IsDestroyed = "true"
					}
					ships[idx] = ship
				}
				isAdjacentToDestroyedShip := false
				destroyedShipsExist := false
				for _, ship := range ships {
					if ship.IsDestroyed == "true" {
						destroyedShipsExist = true
						isAdjacent, err := isAdjacentShip(char, ship.Coords, 1)
						if err != nil {
							continue
						}
						if isAdjacent {
							isAdjacentToDestroyedShip = true
							break
						}
					}
				}
				if destroyedShipsExist && !isAdjacentToDestroyedShip {
					ui.Draw(gui.NewText(1, 2, "Coordinate is not adjacent the ship!", errorText))
					j--
					continue
				}

				newShipLayout = append(newShipLayout, char)
				// update map
				ships = mapShips(newShipLayout)

				for idx, ship := range ships {
					if len(ship.Coords) == shipTypes[idx] {
						ship.IsDestroyed = "false"
						// Mark surrounding area as missed
						for _, surrCoord := range ship.SurroundingArea {
							row := int(surrCoord[0] - 'A')
							col, err := strconv.Atoi(surrCoord[1:])
							if err == nil {
								opponentStates[row][col-1] = gui.Miss
							}
						}
					} else {
						ship.IsDestroyed = "true"
					}
					ships[idx] = ship
				}

				// Update opponentStates with the new ship layout
				row := int(char[0] - 'A')
				col, err := strconv.Atoi(char[1:])
				if err == nil {
					opponentStates[row][col-1] = gui.Ship
				}

			}

			// Update opponentStates with the new ship layout
			row := int(char[0] - 'A')
			col, err := strconv.Atoi(char[1:])
			if err == nil {
				opponentStates[row][col-1] = gui.Ship
			}

			opponentBoard.SetStates(opponentStates)

		}

		// Update all ships' IsDestroyed status after placing each ship
		ships = mapShips(newShipLayout)
		for idx, ship := range ships {
			if len(ship.Coords) == shipTypes[i] {
				ship.IsDestroyed = "false"
				// Mark surrounding area as missed
				for _, surrCoord := range ship.SurroundingArea {
					row := int(surrCoord[0] - 'A')
					col, err := strconv.Atoi(surrCoord[1:])
					if err == nil {
						opponentStates[row][col-1] = gui.Miss
					}
				}
			} else {
				ship.IsDestroyed = "true"
			}
			ships[idx] = ship

		}
		ui.Draw(gui.NewText(1, 1, fmt.Sprintf("Placing %x/10 ship of size %d", i+2, shipTypes[i]), nil))
	}

	if len(newShipLayout) < 20 {
		ui.Draw(gui.NewText(1, 0, "Layout invalid, using default", errorText))
	} else {
		ui.Draw(gui.NewText(1, 0, "New ship layout saved", defaultText))
		DefaultGameInitData.Coords = make([]string, len(newShipLayout))
		copy(DefaultGameInitData.Coords, newShipLayout)
	}
	time.Sleep(2 * time.Second)
	go MainMenu(ui)
}

func opponentBoardOperations(ctx context.Context, playerToken string, opponentBoard *gui.Board, opponentStates [10][10]gui.State, ui *gui.GUI, btnArea *gui.HandleArea) {
	var totalShots int
	var successfulShots int
	var shotCoordinates []string
	var fireMapMutex sync.Mutex
	shipsShot := []string{}
	errorTextConfig := gui.NewTextConfig()
	errorTextConfig.FgColor = gui.Red
	errorTextConfig.BgColor = gui.Black
	go func() {
		ctx := context.Background()

		for {
			if clicked := btnArea.Listen(ctx); clicked == "exitButton" {
				ui.Draw(gui.NewText(40, 24, "Leaving game...", errorTextConfig))
				_, err := retryOnError(ui, func() (string, error) {
					return AbandonGame(playerToken)
				})
				if err != nil {
					ui.Draw(gui.NewText(25, 24, "Error leaving game: "+err.Error(), errorTextConfig))
					continue
				}

				return
			}
		}
	}()
	var statusMapMutex sync.Mutex
	for {
		select {
		case <-ctx.Done(): // cancel context when the game ends
			return
		default:
			// Get game status
			var gameStatus string
			var err error
			var statusMap map[string]interface{}

			for {
				time.Sleep(200 * time.Millisecond)
				// Get game status
				gameStatus, err = retryOnError(ui, func() (string, error) {
					return GetGameStatus(playerToken)
				})
				if err != nil {
					ui.Draw(gui.NewText(1, 28, "Error getting game status: "+err.Error(), errorTextConfig))
					continue
				}

				// Lock before accessing statusMap
				statusMapMutex.Lock()
				err = json.Unmarshal([]byte(gameStatus), &statusMap)
				statusMapMutex.Unlock()
				if err != nil {
					ui.Draw(gui.NewText(0, 0, "Error parsing game status no.%s: "+err.Error(), errorTextConfig))
					continue
				}

				break
			}

			shouldFire, ok := statusMap["should_fire"].(bool)
			if !ok || !shouldFire {
				continue
			}

			// Listen for input
			listenCtx, cancel := context.WithCancel(context.Background())
			char := opponentBoard.Listen(listenCtx)
			cancel()
			// make row and col from char
			col := int(char[0] - 'A')
			var row int
			if len(char) == 3 {
				row = 9
			} else {
				row = int(char[1] - '1')
			}
			// check if the shot was already made
			found := false
			for _, coordinate := range shotCoordinates {
				if coordinate == char {
					found = true
					break
				}
			}
			if found {
				ui.Draw(gui.NewText(24, 2, "You have already fired at this coordinate", errorTextConfig))
				continue
			} else {
				ui.Draw(gui.NewText(24, 2, "                                         ", errorTextConfig))
			}
			// get fire response
			fireResponse, err := retryOnError(ui, func() (string, error) {
				return FireAtEnemy(playerToken, char)
			})
			if err != nil {
				ui.Draw(gui.NewText(1, 29, "Error firing at enemy: "+err.Error(), errorTextConfig))
				continue
			}
			totalShots++

			// Lock before accessing fireMap
			fireMapMutex.Lock()
			var fireMap map[string]interface{}
			err = json.Unmarshal([]byte(fireResponse), &fireMap)
			fireMapMutex.Unlock()

			if err != nil || fireMap == nil {
				continue
			}

			if result, ok := fireMap["result"].(string); ok {
				// Update board states based on fire response
				switch result {
				case "hit":
					shipsShot = append(shipsShot, char)
					opponentStates[col][row] = gui.Hit
					successfulShots++
				case "sunk":
					shipsShot = append(shipsShot, char)
					shipsShotMap := mapShips(shipsShot)
					for _, ship := range shipsShotMap {
						for _, coord := range ship.Coords {
							if coord == char {
								// Mark all coordinates of the ship as sunk
								for _, shipCoord := range ship.Coords {
									col := int(shipCoord[0] - 'A')
									row, _ := strconv.Atoi(shipCoord[1:])
									opponentStates[col][row-1] = gui.Sunk
								}
								// Mark the surrounding area of the ship as misses
								for _, shipCoord := range ship.Coords {
									col := int(shipCoord[0] - 'A')
									row, _ := strconv.Atoi(shipCoord[1:])
									// Check the surrounding cells
									for dCol := -1; dCol <= 1; dCol++ {
										for dRow := -1; dRow <= 1; dRow++ {
											newCol := col + dCol
											newRow := row - 1 + dRow

											if newCol >= 0 && newCol < 10 && newRow >= 0 && newRow < 10 && opponentStates[newCol][newRow] != gui.Sunk {
												opponentStates[newCol][newRow] = gui.Miss
												// Convert the coordinates back to the string format
												missCoord := fmt.Sprintf("%c%d", 'A'+newCol, newRow+1)
												// Add the coordinates to the shotCoordinates slice
												shotCoordinates = append(shotCoordinates, missCoord)
											}
										}
									}
								}
							}
						}
					}
					successfulShots++

				case "miss":
					opponentStates[col][row] = gui.Miss
				}
			} else {
				ui.Draw(gui.NewText(1, 29, "Error parsing fire response", errorTextConfig))
				continue
			}
			// Display fire accuracy
			var shotAccuracyText = "Shot accuracy: N/A"
			if totalShots > 0 {
				shotAccuracy := (float64(successfulShots) / float64(totalShots)) * 100
				shotAccuracyText = fmt.Sprintf("Shot accuracy: %.2f%%", shotAccuracy)
			}
			ui.Draw(gui.NewText(1, 26, shotAccuracyText, defaultText))
			shotCoordinates = append(shotCoordinates, char)
			opponentBoard.SetStates(opponentStates)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func playerBoardOperations(ctx context.Context, playerToken string, playerBoard *gui.Board, playerStates [10][10]gui.State, ui *gui.GUI, shipStatus map[string]bool, dataCoords []string) {
	for {
		select {
		case <-ctx.Done(): // cancel context when the game ends
			return
		default:
			processOpponentShots(playerToken, playerStates, ui, shipStatus, playerBoard, dataCoords)

		}
	}
}

func processOpponentShots(playerToken string, playerStates [10][10]gui.State, ui *gui.GUI, shipStatus map[string]bool, playerBoard *gui.Board, dataCoords []string) {
	for {
		time.Sleep(200 * time.Millisecond)

		gameStatus, err := retryOnError(ui, func() (string, error) {
			return GetGameStatus(playerToken)
		})
		if err != nil {
			ui.Draw(gui.NewText(1, 28, "Error getting game status: "+err.Error(), errorText))
			return
		}

		var statusMap map[string]interface{}

		err = json.Unmarshal([]byte(gameStatus), &statusMap)
		if err != nil {
			ui.Draw(gui.NewText(1, 28, "Error parsing game status: "+err.Error(), errorText))
			return
		}

		oppShots, ok := statusMap["opp_shots"].([]interface{})
		if !ok || len(oppShots) == 0 {
			break // No more shots to process
		}

		ships := mapShips(dataCoords)

		for _, shot := range oppShots {
			if coord, isString := shot.(string); isString {
				col := int(coord[0] - 'A')
				var row int
				if len(coord) == 3 {
					row = 9
				} else {
					row = int(coord[1] - '1')
				}

				isHit := false
				for _, staticCoord := range dataCoords { // Check if the shot is a hit
					if staticCoord == coord {
						isHit = true
						break
					}
				}

				if isHit {
					isSinglePieceShip := false
					hitShip := Ship{}
					for _, ship := range ships {

						for _, shipCoord := range ship.Coords {
							if shipCoord == coord {
								hitShip = ship
								if len(ship.Coords) == 1 {
									isSinglePieceShip = true
								}
								break
							}
						}
					}

					if isSinglePieceShip {
						playerStates[col][row] = gui.Sunk
					} else {
						playerStates[col][row] = gui.Hit
						allPartsHit := true
						for _, shipCoord := range hitShip.Coords {
							shipCol := int(shipCoord[0] - 'A')
							shipRow := int(shipCoord[1] - '1')
							if playerStates[shipCol][shipRow] != gui.Hit {
								allPartsHit = false
								break
							}
						}
						if allPartsHit {
							// Iterate through ship coordinates to mark them as sunk
							for _, shipCoord := range hitShip.Coords {
								shipCol := int(shipCoord[0] - 'A')
								shipRow := int(shipCoord[1] - '1')
								playerStates[shipCol][shipRow] = gui.Sunk
								shipStatus[shipCoord] = true
							}
						}
					}
				} else {
					playerStates[col][row] = gui.Miss
				}
			}
		}
		// Update the player board with the new states
		playerBoard.SetStates(playerStates)
		time.Sleep(100 * time.Millisecond)
	}
}

func displayGameStatus(ctx context.Context, playerToken string, ui *gui.GUI, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done(): // cancel context when the game ends
			return
		default:
			time.Sleep(200 * time.Millisecond)

			gameStatus, err := retryOnError(ui, func() (string, error) {
				return GetGameStatus(playerToken)
			})
			if err != nil {
				ui.Draw(gui.NewText(1, 28, "Error getting game status: "+err.Error(), errorText))
				return
			}

			var statusMap map[string]interface{}
			err = json.Unmarshal([]byte(gameStatus), &statusMap)
			if err != nil {
				ui.Draw(gui.NewText(1, 28, "Error getting game status: "+err.Error(), errorText))
				return
			}

			// statuses for determining the end of the game
			gameStatusStr, gameStatusExists := statusMap["game_status"].(string)
			lastGameStatus, lastGameStatusExists := statusMap["last_game_status"].(string)

			// timer
			timerValue, timerExists := statusMap["timer"].(float64)
			if timerExists {
				timerText := fmt.Sprintf("Timer: %.0f", timerValue)
				ui.Draw(gui.NewText(43, 1, timerText, defaultText))
			}

			// should fire text
			if gameStatusExists && gameStatusStr != "ended" {
				shouldFireText := "Should fire: No!"
				shouldFire, shouldFireExists := statusMap["should_fire"].(bool)
				if shouldFireExists && shouldFire {
					shouldFireText = "Should fire: Yes"
				}
				ui.Draw(gui.NewText(40, 0, shouldFireText, defaultText))
			}

			// Display user details
			userNick := DefaultGameInitData.Nick
			ui.Draw(gui.NewText(2, 27, "User Nick: "+userNick, defaultText))
			userDesc := DefaultGameInitData.Desc
			userDescChunks := splitIntoChunks(userDesc, 25)
			for i, chunk := range userDescChunks {
				ui.Draw(gui.NewText(2, 28+i, chunk, defaultText))
			}

			// Display opponent details
			opponent, opponentExists := statusMap["opponent"].(string)
			if opponentExists {
				ui.Draw(gui.NewText(60, 27, "Opponent Nick: "+opponent, defaultText))
			}

			oppDescValue, err := retryOnError(ui, func() (string, error) {
				return GetGameDescription(playerToken)
			})
			if err != nil {
				ui.Draw(gui.NewText(1, 28, "Error getting game status: "+err.Error(), errorText))
				return
			}
			// display opp desc as chunks
			oppDescChunks := splitIntoChunks(oppDescValue, 25)
			for i, chunk := range oppDescChunks {
				ui.Draw(gui.NewText(60, 28+i, chunk, defaultText))
			}

			// Display end game status, cancel goroutines and return to main menu
			if gameStatusExists && lastGameStatusExists && gameStatusStr == "ended" && lastGameStatus == "lose" {
				ui.Draw(gui.NewText(3, 1, "Unfortunately You Lose", errorText))
				cancel()
				time.Sleep(5 * time.Second)
				go MainMenu(ui)

			} else if gameStatusStr == "ended" && lastGameStatus == "win" {
				win := gui.NewTextConfig()
				win.FgColor = gui.Green
				win.BgColor = gui.Black
				ui.Draw(gui.NewText(3, 1, "Congratulations You Win", win))
				cancel()
				time.Sleep(5 * time.Second)
				go MainMenu(ui)
			}

		}
		time.Sleep(100 * time.Millisecond)
	}
}
