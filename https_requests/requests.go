package https_requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func InitGame() (map[string]interface{}, string, error) {
	data := map[string]interface{}{
		"coords": []string{"A2", "A4", "B9", "C7", "D1", "D2", "D3", "D4", "D7", "E7", "F1", "F2", "F3", "F5", "G5", "G8", "G9", "I4", "J4", "J8"},
		"desc":   "pierwszy raz",
		"nick":   "Jan_Niecny",
		"wpbot":  false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, "", err
	}

	resp, err := http.Post("https://go-pjatk-server.fly.dev/api/game", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", errors.New("status is not 200")

	}

	playerStatsResponse, err := GetPlayerStats("Jan_Niecny")
	if err != nil {
		return nil, "", err
	}

	var playerStats map[string]map[string]interface{}
	err = json.Unmarshal([]byte(playerStatsResponse), &playerStats)
	if err != nil {
		return nil, "", err
	}

	playerToken := resp.Header.Get("x-auth-token")

	return data, playerToken, nil
}

func GetLobbyInfo() (string, error) {
	resp, err := http.Get("https://go-pjatk-server.fly.dev/api/lobby")
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

func DisplayLobbyStatus() (string, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	displayedPlayers := make(map[string]bool)

	for range ticker.C {
		lobbyInfo, err := GetLobbyInfo()
		if err != nil {
			return "", err
		}

		var lobby []map[string]string
		err = json.Unmarshal([]byte(lobbyInfo), &lobby)
		if err != nil {
			var message map[string]string
			err = json.Unmarshal([]byte(lobbyInfo), &message)
			if err != nil {
				return "", err
			}

			if message["message"] == "A wild Chaos Monkey appeared!" {
				continue
			}
		}

		currentPlayers := make(map[string]bool)
		for _, player := range lobby {
			currentPlayers[player["nick"]] = true

			if !displayedPlayers[player["nick"]] {
				fmt.Println("Player:", player["nick"], "Status:", player["game_status"])
			}

			if player["nick"] == "Jan_Niecny" && player["game_status"] != "waiting" {
				fmt.Println("Player's game status is not waiting, breaking...")
				return player["nick"], nil
			}
		}

		displayedPlayers = currentPlayers
	}

	return "", nil
}

func GetPlayerStats(nick string) (string, error) {
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

func GetBoardInfoWithRetry(playerToken string) ([]string, error) {
	const maxRetries = 5
	const initialDelay = time.Second

	var boardInfo []string
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		boardInfo, err = GetBoardInfo(playerToken)
		if err == nil {
			return boardInfo, nil
		}

		if isRetryableError(err) {
			delay := initialDelay * time.Duration(retry+1)
			time.Sleep(delay)
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("exceeded maximum retries: %w", err)
}

func GetBoardInfo(playerToken string) ([]string, error) {
	const maxRetries = 10
	retryDelay := 1 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/board", nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("X-Auth-Token", playerToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error on attempt %d: %v\n", retry+1, err)
			time.Sleep(retryDelay)
			retryDelay *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error on attempt %d: status code %d\n", retry+1, resp.StatusCode)
			resp.Body.Close()
			time.Sleep(retryDelay)
			retryDelay *= 2
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var response struct {
			Message string   `json:"message"`
			Board   []string `json:"board"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}

		if response.Message != "" {
			return nil, errors.New(response.Message)
		}

		return response.Board, nil
	}

	return nil, errors.New("exceeded maximum number of retries")
}

func isRetryableError(err error) bool {
	//always retry
	return true
}

func GetGameStatus(playerToken string) (string, error) {
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

func FireAtEnemy(playerToken, coord string) (string, error) {
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

func GetGameDescription(playerToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/desc", nil)
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

	var gameDesc map[string]string
	err = json.Unmarshal(body, &gameDesc)
	if err != nil {
		return "", err
	}

	return gameDesc["opp_desc"], nil
}
