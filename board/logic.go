package board

import (
	"encoding/json"
	"fmt"

	gui "github.com/grupawp/warships-gui/v2"
)

func UpdateAndDisplayGameStatus(playerToken string, ui *gui.GUI, getGameStatus func(string) (string, error)) {
	gameStatus, err := getGameStatus(playerToken)
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

	gameStatusStr, gameStatusExists := statusMap["game_status"].(string)
	lastGameStatusStr, lastGameStatusExists := statusMap["last_game_status"].(string)

	ui.Draw(gui.NewText(1, 0, shouldFireText, nil))
	ui.Draw(gui.NewText(1, 1, timerText, nil))
	ui.Draw(gui.NewText(1, 2, oppShotsText, nil))
	if gameStatusExists && lastGameStatusExists && gameStatusStr == "ended" && lastGameStatusStr == "lose" {
		ui.Draw(gui.NewText(1, 0, " Unfortunately You Lose", nil))
	} else if gameStatusStr == "ended" && lastGameStatusStr == "win" {
		ui.Draw(gui.NewText(1, 0, " Congratulations You Win", nil))
	}
}
