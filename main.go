package main

import (
	"BomboweStatki/board"
	"BomboweStatki/https_requests"
	"context"
	"encoding/json"
	"fmt"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

func main() {
	tries := 0
	var playerToken string
	for {
		if tries >= 10 {
			fmt.Println("Exceeded maximum number of tries. Exiting.")
			return
		}

		var err error
		_, playerToken, err = https_requests.InitGame()
		if err != nil {
			fmt.Printf("Error on attempt %d: %v\n", tries+1, err)
			tries++
			time.Sleep(time.Second)
			continue
		}

		if playerToken == "" {
			fmt.Println("X-Auth-Token not provided by the server.")
			tries++
			continue
		}

		break
	}
	playerName, err := https_requests.DisplayLobbyStatus()
	if err != nil {
		fmt.Println("Error displaying lobby status:", err)
		return
	}

	fmt.Println("Player's name:", playerName)

	https_requests.GetLobbyInfo()

	playerStates, opponentStates, shipStatus, err := board.Config(playerToken)
	if err != nil {
		fmt.Println(err)
		return
	}

	ui, playerBoard, opponentBoard := board.GuiInit(playerStates, opponentStates)

	var shotCoordinates = make(map[string]bool)

	dataCoords, err := https_requests.GetBoardInfo(playerToken)
	if err != nil {
		fmt.Println("Error getting board info:", err)
		return
	}

	go opponentBoardOperations(playerToken, opponentBoard, playerStates, opponentStates, shotCoordinates, ui)

	go playerBoardOperations(playerToken, playerBoard, opponentStates, playerStates, ui, shipStatus, dataCoords)

	ui.Start(context.TODO(), nil)
}

func opponentBoardOperations(playerToken string, opponentBoard *gui.Board, playerStates, opponentStates [10][10]gui.State, shotCoordinates map[string]bool, ui *gui.GUI) {
	for {
		char := opponentBoard.Listen(context.TODO())
		col := int(char[0] - 'A')
		var row int
		if len(char) == 3 {
			row = 9
		} else {
			row = int(char[1] - '1')
		}

		gameStatus, err := https_requests.GetGameStatus(playerToken)
		if err != nil {
			ui.Draw(gui.NewText(1, 1, "Error getting game status: "+err.Error(), nil))
			continue
		}

		var statusMap map[string]interface{}
		err = json.Unmarshal([]byte(gameStatus), &statusMap)
		if err != nil {
			ui.Draw(gui.NewText(1, 1, "Error parsing game status: "+err.Error(), nil))
			continue
		}

		shouldFire, ok := statusMap["should_fire"].(bool)
		if ok && shouldFire {
			coord := fmt.Sprintf("%c%d", 'A'+col, row+1)
			if _, alreadyShot := shotCoordinates[coord]; alreadyShot {
				continue
			}

			shotCoordinates[coord] = true
			fireResponse, err := https_requests.FireAtEnemy(playerToken, coord)
			if err != nil {
				ui.Draw(gui.NewText(1, 1, "Error firing at enemy: "+err.Error(), nil))
				continue
			}

			var fireMap map[string]interface{}
			err = json.Unmarshal([]byte(fireResponse), &fireMap)
			if err != nil {
				ui.Draw(gui.NewText(1, 1, "Error parsing fire response: "+err.Error(), nil))
				continue
			}

			result, resultExists := fireMap["result"].(string)
			if resultExists {
				if result == "hit" || result == "sunk" {
					opponentStates[col][row] = gui.Hit
				} else if result == "miss" {
					opponentStates[col][row] = gui.Miss
				}
				fireResponseText := fmt.Sprintf("Fire response: %s", result)
				ui.Draw(gui.NewText(1, 25, fireResponseText, nil))
			} else {
				opponentStates[col][row] = gui.Empty
			}

			opponentBoard.SetStates(opponentStates)
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
			board.UpdateBoardStates(opponentBoard, boardInfo)
		}
	}
}

func playerBoardOperations(playerToken string, playerBoard *gui.Board, opponentStates [10][10]gui.State, playerStates [10][10]gui.State, ui *gui.GUI, shipStatus map[string]bool, dataCoords []string) {
	for {
		processOpponentShots(playerToken, opponentStates, playerStates, ui, shipStatus, playerBoard, dataCoords)

		displayGameStatus(playerToken, ui)

		extraTurn := checkExtraTurn(playerToken)
		if extraTurn {
			continue
		}

		time.Sleep(time.Second)
	}
}

func checkExtraTurn(playerToken string) bool {
	gameStatus, err := https_requests.GetGameStatus(playerToken)
	if err != nil {
		fmt.Println("Error getting game status:", err)
		return false
	}

	var statusMap map[string]interface{}
	err = json.Unmarshal([]byte(gameStatus), &statusMap)
	if err != nil {
		fmt.Println("Error parsing game status:", err)
		return false
	}

	extraTurn, ok := statusMap["extra_turn"].(bool)
	if ok && extraTurn {
		return true
	}

	return false
}

func processOpponentShots(playerToken string, opponentStates [10][10]gui.State, playerStates [10][10]gui.State, ui *gui.GUI, shipStatus map[string]bool, playerBoard *gui.Board, dataCoords []string) {
	for _, coord := range dataCoords {
		col := int(coord[0] - 'A')
		var row int
		if len(coord) == 3 {
			row = 9
		} else {
			row = int(coord[1] - '1')
		}
		playerStates[col][row] = gui.Ship
	}
	gameStatus, err := https_requests.GetGameStatus(playerToken)
	if err != nil {
		ui.Draw(gui.NewText(1, 1, "Error getting game status: "+err.Error(), nil))
		return
	}

	var statusMap map[string]interface{}
	err = json.Unmarshal([]byte(gameStatus), &statusMap)
	if err != nil {
		ui.Draw(gui.NewText(1, 1, "Error parsing game status: "+err.Error(), nil))
		return
	}

	oppShots, ok := statusMap["opp_shots"].([]interface{})
	if ok {
		for _, shot := range oppShots {
			coord := shot.(string)
			col := int(coord[0] - 'A')
			var row int
			if len(coord) == 3 {
				row = 9
			} else {
				row = int(coord[1] - '1')
			}

			isHit := false
			for _, staticCoord := range dataCoords {
				if staticCoord == coord {
					isHit = true
					break
				}
			}

			if isHit {
				playerStates[col][row] = gui.Hit
				allPartsHit := true
				for _, state := range playerStates[col] {
					if state == gui.Ship {
						allPartsHit = false
						break
					}
				}
				if allPartsHit {
					shipStatus[coord] = true
				}
			} else {
				playerStates[col][row] = gui.Miss
			}
		}
	}

	boardInfo := make([]string, 10)
	for i, row := range playerStates {
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
	board.UpdateBoardStates(playerBoard, boardInfo)
}

func displayGameStatus(playerToken string, ui *gui.GUI) {
	gameStatus, err := https_requests.GetGameStatus(playerToken)
	if err != nil {
		ui.Draw(gui.NewText(1, 1, "Error getting game status: "+err.Error(), nil))
		return
	}

	var statusMap map[string]interface{}
	err = json.Unmarshal([]byte(gameStatus), &statusMap)
	if err != nil {
		ui.Draw(gui.NewText(1, 1, "Error parsing game status: "+err.Error(), nil))
		return
	}

	gameStatus, gameStatusExists := statusMap["game_status"].(string)
	lastGameStatus, lastGameStatusExists := statusMap["last_game_status"].(string)

	if gameStatusExists && gameStatus != "ended" {
		shouldFireText := "Should fire: No!"
		shouldFire, shouldFireExists := statusMap["should_fire"].(bool)
		if shouldFireExists && shouldFire {
			shouldFireText = "Should fire: Yes"
		}
		ui.Draw(gui.NewText(1, 0, shouldFireText, nil))
	}

	oppDescValue, err := https_requests.GetGameDescription(playerToken)
	if err != nil {
		ui.Draw(gui.NewText(1, 1, "Error getting game description: "+err.Error(), nil))
		return
	}
	oppDescText := fmt.Sprintf("Opponent description: %s", oppDescValue)
	ui.Draw(gui.NewText(1, 1, oppDescText, nil))

	var oppShotsText string
	oppShotsValue, oppShotsExists := statusMap["opp_shots"].([]interface{})
	if oppShotsExists {
		oppShotsText = fmt.Sprintf("Opponent shots: %v", oppShotsValue)
	}
	ui.Draw(gui.NewText(1, 2, oppShotsText, nil))

	if gameStatusExists && lastGameStatusExists && gameStatus == "ended" && lastGameStatus == "lose" {
		ui.Draw(gui.NewText(1, 0, " Unfortunately You Lose", nil))
	} else if gameStatus == "ended" && lastGameStatus == "win" {
		ui.Draw(gui.NewText(1, 0, " Congratulations You Win", nil))
	}
}
