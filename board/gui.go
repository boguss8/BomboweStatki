package board

import (
	https_requests "BomboweStatki/https_requests"

	gui "github.com/grupawp/warships-gui/v2"
)

func Config(playerToken string) (playerStates [10][10]gui.State, opponentStates [10][10]gui.State, shipStatus map[string]bool, err error) {
	playerStates = [10][10]gui.State{}
	opponentStates = [10][10]gui.State{}

	for i := range playerStates {
		for j := range playerStates[i] {
			playerStates[i][j] = gui.Empty
			opponentStates[i][j] = gui.Empty
		}
	}

	shipStatus = make(map[string]bool)

	boardInfo, err := https_requests.GetBoardInfo(playerToken)
	if err != nil {
		return playerStates, opponentStates, shipStatus, err
	}

	for _, coord := range boardInfo {
		shipStatus[coord] = false
	}

	return playerStates, opponentStates, shipStatus, nil
}

func GuiInit(playerStates [10][10]gui.State, opponentStates [10][10]gui.State) (ui *gui.GUI, playerBoard *gui.Board, opponentBoard *gui.Board) {
	ui = gui.NewGUI(true)
	boardConfig := gui.NewBoardConfig()
	boardConfig.HitColor = gui.NewColor(0, 0, 255)
	boardConfig.MissColor = gui.NewColor(0, 255, 0)
	boardConfig.ShipColor = gui.NewColor(0, 0, 255)

	playerBoard = gui.NewBoard(1, 3, boardConfig)
	opponentBoard = gui.NewBoard(50, 3, boardConfig)

	ui.Draw(playerBoard)
	ui.Draw(opponentBoard)
	ui.Draw(gui.NewText(1, 28, "Press Ctrl+C to exit", nil))

	playerBoard.SetStates(playerStates)
	opponentBoard.SetStates(opponentStates)

	return ui, playerBoard, opponentBoard
}
