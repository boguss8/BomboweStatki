package client

import (
	"time"

	gui "github.com/s25867/warships-gui/v2"
)

type ServerRequestWithPlayerStats func() (PlayerStats, error)

type ServerRequestWithPlayers func() ([]Player, string, error)

type ServerRequest func() (string, error)

func retryOnErrorWithPlayerStats(ui *gui.GUI, serverRequest ServerRequestWithPlayerStats) (PlayerStats, error) {
	var playerStats PlayerStats
	var err error

	for i := 0; i < 20; i++ {
		playerStats, err = serverRequest()
		if err == nil {
			return playerStats, nil
		}
		ui.Draw(gui.NewText(1, 28, "Error: "+err.Error()+". Retrying in 2 seconds...", errorText))
		time.Sleep(100 * time.Millisecond)
	}

	return playerStats, err
}

func retryOnErrorWithPlayers(ui *gui.GUI, serverRequest ServerRequestWithPlayers) ([]Player, string, error) {
	var players []Player
	var result string
	var err error

	for i := 0; i < 20; i++ {
		players, result, err = serverRequest()
		if err == nil {
			return players, result, nil
		}
		ui.Draw(gui.NewText(1, 28, "Error: "+err.Error()+". Retrying in 2 seconds...", errorText))
		time.Sleep(100 * time.Millisecond)
	}

	return players, result, err
}

func retryOnError(ui *gui.GUI, serverRequest ServerRequest) (string, error) {
	var result string
	var err error

	for i := 0; i < 20; i++ {
		result, err = serverRequest()
		if err == nil {
			return result, nil
		}
		ui.Draw(gui.NewText(1, 28, "Error: "+err.Error()+". Retrying...", errorText))
		time.Sleep(100 * time.Millisecond)
	}

	return result, err
}
