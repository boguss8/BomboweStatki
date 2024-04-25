package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

func getGameStatus(playerToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Auth-Token", playerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func getPlayerStats(nick string) (string, error) {
	url := fmt.Sprintf("https://go-pjatk-server.fly.dev/api/stats/%s", nick)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func getBoardInfo(playerToken string) ([]string, error) {
	req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/board", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Token", playerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var boardInfo map[string][]string
	err = json.Unmarshal(body, &boardInfo)
	if err != nil {
		return nil, err
	}

	return boardInfo["board"], nil
}

func fireAtEnemy(playerToken, coord string) (string, error) {
	data := map[string]string{
		"coord": coord,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://go-pjatk-server.fly.dev/api/game/fire", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Auth-Token", playerToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func updateBoardStates(board *gui.Board, boardInfo []string) {
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
				newStates[i][j] = gui.State("Hit")
			case 'M':
				newStates[i][j] = gui.State("Miss")
			case 'S':
				newStates[i][j] = gui.State("Ship")
			default:
				newStates[i][j] = gui.State("Empty")
			}
		}
	}

	board.SetStates(newStates)
}

func main() {
	data := map[string]interface{}{
		"coords": []string{"A2", "A4", "B9", "C7", "D1", "D2", "D3", "D4", "D7", "E7", "F1", "F2", "F3", "F5", "G5", "G8", "G9", "I4", "J4", "J8"},
		"desc":   "pierwszy raz",
		"nick":   "Jan_Niecny",
		"wpbot":  true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := http.Post("https://go-pjatk-server.fly.dev/api/game", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	playerStatsResponse, err := getPlayerStats("Jan_Niecny")
	if err != nil {
		fmt.Println(err)
		return
	}

	var playerStats map[string]map[string]interface{}
	err = json.Unmarshal([]byte(playerStatsResponse), &playerStats)
	if err != nil {
		fmt.Println(err)
		return
	}

	stats := playerStats["stats"]

	gamesValue, gamesExists := stats["games"]
	winsValue, winsExists := stats["wins"]

	if !gamesExists || !winsExists {
		fmt.Println("Games or wins data not found")
		return
	}

	games := int(gamesValue.(float64))
	wins := int(winsValue.(float64))

	fmt.Println("Games:", games)
	fmt.Println("Wins:", wins)

	fmt.Println("response Status:", resp.Status)
	playerToken := resp.Header.Get("x-auth-token")
	fmt.Println("x-auth-token:", playerToken)

	ui := gui.NewGUI(true)
	txt := gui.NewText(1, 1, "Press on any coordinate to log it.", nil)
	ui.Draw(txt)
	ui.Draw(gui.NewText(1, 2, "Press Ctrl+C to exit", nil))

	board := gui.NewBoard(1, 4, nil)
	ui.Draw(board)

	go func() {
		for {
			coord := board.Listen(context.TODO())
			txt.SetText(fmt.Sprintf("Coordinate: %s", coord))
			ui.Log("Coordinate: %s", coord)

			fireResponse, err := fireAtEnemy(playerToken, coord)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("Fire response:", fireResponse)

			boardInfo, err := getBoardInfo(playerToken)
			if err != nil {
				fmt.Println(err)
				continue
			}

			updateBoardStates(board, boardInfo)

			ui.Remove(txt)
			txt = gui.NewText(1, 1, fmt.Sprintf("Coordinate: %s", coord), nil)
			ui.Draw(txt)
		}
	}()

Loop:
	for {
		gameStatusResponse, err := getGameStatus(playerToken)
		if err != nil {
			fmt.Println(err)
			return
		}
		var gameStatusObj map[string]interface{}
		err = json.Unmarshal([]byte(gameStatusResponse), &gameStatusObj)
		if err != nil {
			fmt.Println(err)
			return
		}
		playerStatsResponse, err := getPlayerStats("Jan_Niecny")
		if err != nil {
			fmt.Println(err)
			return
		}

		var playerStats map[string]map[string]interface{}
		err = json.Unmarshal([]byte(playerStatsResponse), &playerStats)
		if err != nil {
			fmt.Println(err)
			return
		}

		stats := playerStats["stats"]

		gamesValue, gamesExists := stats["games"]
		winsValue, winsExists := stats["wins"]

		if !gamesExists || !winsExists {
			fmt.Println("Games or wins data not found")
			return
		}

		games2 := int(gamesValue.(float64))
		wins2 := int(winsValue.(float64))

		fmt.Println("Games:", games2)
		fmt.Println("Wins:", wins2)
		gameStatus := gameStatusObj["game_status"].(string)

		switch gameStatus {
		case "waiting":
			fmt.Println("Waiting for the opponent")
		case "waiting_wpbot":
			fmt.Println("Waiting for the WPBot")
		case "game_in_progress":
			playerToken := resp.Header.Get("x-auth-token")
			fmt.Println("x-auth-token:", playerToken)

			boardInfo, err := getBoardInfo(playerToken)
			if err != nil {
				fmt.Println(err)
				return
			}

			updateBoardStates(board, boardInfo)

			fmt.Println("Board info:", boardInfo)

			ui.Start(context.TODO(), nil)

			shouldFire, exists := gameStatusObj["should_fire"].(bool)
			if exists && shouldFire {
				fmt.Print("Enter coordinate to fire at: ")

				coordCh := make(chan string)
				go func() {
					var coord string
					fmt.Scanln(&coord)
					coordCh <- coord
				}()

				select {
				case coord := <-coordCh:
					fireResponse, err := fireAtEnemy(playerToken, coord)
					if err != nil {
						fmt.Println(err)
						return
					}

					fmt.Println("Fire response:", fireResponse)
				case <-time.After(time.Duration(gameStatusObj["timer"].(float64)) * time.Second):
					fmt.Println("Time's up!")
				}
			}

			fmt.Println("Game in progress")
		case "ended":
			fmt.Println("Game ended")
			if wins2 == wins {
				fmt.Println("You lost the game")
			} else {
				fmt.Println("You won the game")
			}
			break Loop
		default:
			fmt.Println("Unknown status")
		}

		fmt.Println("Game status object:", gameStatusObj)
		time.Sleep(time.Second)
	}
}
