package game

import gui "github.com/s25867/warships-gui/v2"

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
