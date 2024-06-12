package client

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	gui "github.com/s25867/warships-gui/v2"
)

func bomBotInit(ui *gui.GUI, gameData GameInitData) {
	gameDataBot := GameInitData{
		Coords:     generateRandomBoard(),
		Desc:       "Zapewnia wybuchową rozgrywkę!",
		Nick:       "BomBot",
		TargetNick: gameData.Nick,
		Wpbot:      false,
	}

	playerToken, err := retryOnError(ui, func() (string, error) {
		return InitGame(gameData)
	})
	if err != nil {
		ui.Draw(gui.NewText(1, 29, "Error initializing game: "+err.Error()+". Retrying...", nil))
	}

	botToken, err := retryOnError(ui, func() (string, error) {
		return InitGame(gameDataBot)
	})
	if err != nil {
		ui.Draw(gui.NewText(1, 29, "Error initializing game: "+err.Error()+". Retrying...", nil))
	}

	ui.NewScreen("game" + playerToken)
	ui.SetScreen("game" + playerToken)
	err = waitForStart(gameData.Nick, ui)
	if err != nil {
		ui.Draw(gui.NewText(1, 30, "Error waiting for game to start: "+err.Error(), nil))
	}
	err = LaunchGameBoard(ui, playerToken, gameData)
	if err != nil {
		ui.Draw(gui.NewText(1, 31, err.Error(), nil))
	}
	go bomBotShots(ui, botToken)
}

func bomBotShots(ui *gui.GUI, botToken string) {
	// Initialize all possible coordinates
	var fireMapMutex = &sync.Mutex{}
	var statusMapMutex = &sync.Mutex{}
	allCoords := make([]string, 0, 100)
	for i := 'A'; i <= 'J'; i++ {
		for j := 1; j <= 10; j++ {
			allCoords = append(allCoords, fmt.Sprintf("%c%d", i, j))
		}
	}
	// Initialize the ship table
	var hitShots []string
	botTable := mapShips([]string{})

	for {
		// Check game status
		var gameStatus string
		var err error
		var statusMap map[string]interface{}
		for {
			gameStatus, err = retryOnError(ui, func() (string, error) {
				return GetGameStatus(botToken)
			})
			if err != nil {
				ui.Draw(gui.NewText(1, 28, "Error getting game status: "+err.Error(), nil))
				continue
			}

			statusMapMutex.Lock()
			err = json.Unmarshal([]byte(gameStatus), &statusMap)
			statusMapMutex.Unlock()
			if err != nil {
				ui.Draw(gui.NewText(0, 0, "Error parsing game status no.%s: "+err.Error(), nil))
				continue
			}

			break
		}
		shouldFire, ok := statusMap["should_fire"].(bool)
		if !ok || !shouldFire {
			continue
		}
		var randCoord string = ""

		//fire at ship or random
		for _, ship := range botTable {
			if ship.IsDestroyed == "false" && len(ship.SurroundingArea) > 0 {
				randIndex := rand.Intn(len(ship.SurroundingArea))
				randCoord = ship.SurroundingArea[randIndex]
			}
		}
		if randCoord == "" {
			if len(allCoords) > 0 {
				randIndex := rand.Intn(len(allCoords))
				randCoord = allCoords[randIndex]
				allCoords = append(allCoords[:randIndex], allCoords[randIndex+1:]...)
			} else {
				ui.Draw(gui.NewText(1, 28, "Error calculating possible ship locations. Surrendering game...", nil))
				ui.Draw(gui.NewText(40, 24, "Leaving game...", nil))
				_, err := retryOnError(ui, func() (string, error) {
					return AbandonGame(botToken)
				})
				if err != nil {
					ui.Draw(gui.NewText(25, 24, "Error leaving game: "+err.Error(), nil))
					continue
				}
			}
		}

		response, err := retryOnError(ui, func() (string, error) {
			return FireAtEnemy(botToken, randCoord)
		})
		if err != nil {
			ui.Draw(gui.NewText(1, 29, "Error firing at enemy: "+err.Error(), nil))
			continue
		}
		ui.Draw(gui.NewText(1, 28, "Got response...", nil))

		// Lock before accessing fireMap
		fireMapMutex.Lock()
		var fireMap map[string]interface{}
		err = json.Unmarshal([]byte(response), &fireMap)
		fireMapMutex.Unlock()

		if err != nil || fireMap == nil {
			continue
		}

		if result, ok := fireMap["result"].(string); ok {
			if result == "hit" {
				hitShots = append(hitShots, randCoord)
				botTable = mapShips(hitShots)

				for i, ship := range botTable {
					if ship.IsDestroyed != "true" {
						newSurroundingArea := []string{}
						for _, surrCoord := range ship.SurroundingArea {
							// Check if the surrounding coordinate is in the list of all coordinates
							if findIndex(allCoords, surrCoord) != -1 {
								if adjacent, err := isAdjacentShip(surrCoord, ship.Coords, 1); err != nil {
									ui.Draw(gui.NewText(1, 28, "Error: "+err.Error(), nil))
								} else if adjacent {
									// If the surrounding coordinate is adjacent to the ship, add it to the new surrounding area
									newSurroundingArea = append(newSurroundingArea, surrCoord)
								}
							}
						}
						ship.SurroundingArea = newSurroundingArea
					}
					// Since the ship was hit, mark it as not destroyed
					ship.IsDestroyed = "false"
					botTable[i] = ship
				}

			} else if result == "sunk" {
				hitShots = append(hitShots, randCoord)
				botTable = mapShips(hitShots)
				// Find a ship that has been sunk and mark it as destroyed
				for i, ship := range botTable {
					ship.IsDestroyed = "true"
					botTable[i] = ship

					// Remove all coordinates and surrounding coordinates of the sunk ship from allCoords
					for _, coord := range append(ship.Coords, ship.SurroundingArea...) {
						index := findIndex(allCoords, coord)
						if index != -1 {
							allCoords = append(allCoords[:index], allCoords[index+1:]...)
						}
					}
				}
			} else {
				// If the shot missed, check if it was on the surrounding area of a ship that is not destroyed
				for i, ship := range botTable {
					if ship.IsDestroyed != "true" {
						index := findIndex(ship.SurroundingArea, randCoord)
						if index != -1 {
							// Remove the coordinate from the surrounding area of the ship
							ship.SurroundingArea = append(ship.SurroundingArea[:index], ship.SurroundingArea[index+1:]...)
							botTable[i] = ship
						}
					}
				}
				// If the shot missed, add the coordinate to the list of all coordinates
				index := findIndex(allCoords, randCoord)
				if index != -1 {
					allCoords = append(allCoords[:index], allCoords[index+1:]...)
				}
			}
		}
	}
}
