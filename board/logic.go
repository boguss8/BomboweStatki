package board

import (
	https_requests "BomboweStatki/https_requests"
	"encoding/json"
	"fmt"

	gui "github.com/grupawp/warships-gui/v2"
)

func UpdateBoardStates(board *gui.Board, boardInfo []string) {
	var newStates [10][10]gui.State

	for i, row := range boardInfo {
		if i >= 10 {
			break
		}

		for j, state := range row {
			if j >= 10 {
				break
			}

			switch state {
			case 'H':
				newStates[i][j] = gui.Hit
			case 'M':
				newStates[i][j] = gui.Miss
			case 'S':
				newStates[i][j] = gui.Ship
			default:
				newStates[i][j] = gui.Empty
			}
		}
	}

	board.SetStates(newStates)
}

func IsShipAtLocation(coords []string, col, row int) bool {
	for _, coord := range coords {
		if int(coord[0]-'A') == col && int(coord[1]-'1') == row {
			return true
		}
	}
	return false
}

func UpdateAndDisplayGameStatus(playerToken string, ui *gui.GUI) {
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

	shouldFireText := "Should fire: No!"
	shouldFire, shouldFireExists := statusMap["should_fire"].(bool)
	if shouldFireExists && shouldFire {
		shouldFireText = "Should fire: Yes"
	}

	var timerText string
	timerValue, timerExists := statusMap["timer"].(int)
	if timerExists {
		timerText = fmt.Sprintf("Timer: %d", timerValue)
	}

	var oppShotsText string
	oppShotsValue, oppShotsExists := statusMap["opp_shots"].([]interface{})
	if oppShotsExists {
		oppShotsText = fmt.Sprintf("Opponent shots: %v", oppShotsValue)
	}

	gameStatus, gameStatusExists := statusMap["game_status"].(string)
	lastGameStatus, lastGameStatusExists := statusMap["last_game_status"].(string)

	ui.Draw(gui.NewText(1, 0, shouldFireText, nil))
	ui.Draw(gui.NewText(1, 1, timerText, nil))
	ui.Draw(gui.NewText(1, 2, oppShotsText, nil))
	if gameStatusExists && lastGameStatusExists && gameStatus == "ended" && lastGameStatus == "lose" {
		ui.Draw(gui.NewText(1, 0, " Unfortunately You Lose", nil))
	} else if gameStatus == "ended" && lastGameStatus == "win" {
		ui.Draw(gui.NewText(1, 0, " Congratulations You Win", nil))
	}
}
