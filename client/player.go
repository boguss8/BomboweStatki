package client

import (
	"encoding/json"
	"fmt"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

func PlayerBoardOperations(playerToken string, playerBoard *gui.Board, opponentStates [10][10]gui.State, playerStates [10][10]gui.State, ui *gui.GUI, shipStatus map[string]bool, dataCoords []string) {
	for {
		ProcessOpponentShots(playerToken, playerBoard, playerStates, opponentStates, ui, shipStatus, playerBoard, dataCoords)

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
