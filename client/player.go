package client

import (
	board "BomboweStatki/board"
	game "BomboweStatki/game"
	"encoding/json"
	"fmt"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

func PlayerBoardOperations(playerToken string, playerBoard *gui.Board, opponentStates [10][10]gui.State, playerStates [10][10]gui.State, ui *gui.GUI, shipStatus map[string]bool, dataCoords []string) {
	for {
		ProcessOpponentShots(playerToken, opponentStates, playerStates, ui, shipStatus, playerBoard, dataCoords)

		board.UpdateAndDisplayGameStatus(playerToken, ui, GetGameStatus)

		extraTurn := CheckExtraTurn(playerToken, GetGameStatus)
		if extraTurn {
			continue
		}

		time.Sleep(time.Second)
	}
}

func CheckExtraTurn(playerToken string, getGameStatus func(string) (string, error)) bool {
	gameStatus, err := getGameStatus(playerToken)
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

func ProcessOpponentShots(playerToken string, opponentStates [10][10]gui.State, playerStates [10][10]gui.State, ui *gui.GUI, shipStatus map[string]bool, playerBoard *gui.Board, dataCoords []string) {
	gameStatus, err := GetGameStatus(playerToken)
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
	game.UpdateBoardStates(playerBoard, boardInfo)
}
