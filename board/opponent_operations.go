package board

import (
	https_requests "BomboweStatki/https_requests"
	"context"
	"encoding/json"
	"fmt"

	gui "github.com/grupawp/warships-gui/v2"
)

func OpponentBoardOperations(playerToken string, opponentBoard *gui.Board, playerStates, opponentStates [10][10]gui.State, shotCoordinates map[string]bool, ui *gui.GUI) {
	for {
		println("dupa1")
		char := opponentBoard.Listen(context.TODO())
		println("dupa2")
		col := int(char[0] - 'A')
		println("dupa3")
		var row int
		println("dupa4")
		if len(char) == 3 {
			row = 9
		} else {
			row = int(char[1] - '1')
		}
		println("dupa6")
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
		println("dupa7")
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
			UpdateBoardStates(opponentBoard, boardInfo)
		}
	}
}
