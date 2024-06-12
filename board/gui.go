package board

import (
	"strconv"

	gui "github.com/s25867/warships-gui/v2"
)

func Config(shipCoords []string) (playerStates [10][10]gui.State, opponentStates [10][10]gui.State, shipStatus map[string]bool, err error) {
	playerStates = [10][10]gui.State{}
	opponentStates = [10][10]gui.State{}

	for i := range playerStates {
		for j := range playerStates[i] {
			playerStates[i][j] = gui.Empty
			opponentStates[i][j] = gui.Empty
		}
	}

	shipStatus = make(map[string]bool)

	for _, coord := range shipCoords {
		shipStatus[coord] = false
		col := int(coord[0] - 'A')
		row, _ := strconv.Atoi(coord[1:])
		playerStates[col][row-1] = gui.Ship
	}

	return playerStates, opponentStates, shipStatus, nil
}

func GuiInit(ui *gui.GUI, playerStates [10][10]gui.State, opponentStates [10][10]gui.State) (playerBoard *gui.Board, opponentBoard *gui.Board, btnArea *gui.HandleArea) {

	boardConfig := gui.NewBoardConfig()

	playerBoard = gui.NewBoard(1, 3, boardConfig)
	opponentBoard = gui.NewBoard(50, 3, boardConfig)

	exitButtonConfig := gui.NewButtonConfig()
	exitButtonConfig.Width = 0
	exitButtonConfig.Height = 0
	exitButtonConfig.Width = 15
	exitButtonConfig.BgColor = gui.Red
	exitButton := gui.NewButton(40, 25, "Exit", exitButtonConfig)

	btnMapping := map[string]gui.Spatial{
		"exitButton": exitButton,
	}
	btnArea = gui.NewHandleArea(btnMapping)

	ui.Draw(btnArea)
	ui.Draw(exitButton)

	playerBoard.SetStates(playerStates)

	ui.Draw(playerBoard)
	ui.Draw(opponentBoard)

	return playerBoard, opponentBoard, btnArea
}
