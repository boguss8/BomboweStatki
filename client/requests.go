package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type PlayerStats struct {
	Games  int    `json:"games"`
	Nick   string `json:"nick"`
	Points int    `json:"points"`
	Rank   int    `json:"rank"`
	Wins   int    `json:"wins"`
}

type StatsResponse struct {
	Stats []PlayerStats `json:"stats"`
}

type Player struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}

func InitGame(username string, desc string, opponentName string) (string, []string, error) {
	data := map[string]interface{}{
		"coords":      []string{"A2", "A4", "B9", "C7", "D1", "D2", "D3", "D4", "D7", "E7", "F1", "F2", "F3", "F5", "G5", "G8", "G9", "I4", "J4", "J8"},
		"desc":        desc,
		"nick":        username,
		"target_nick": opponentName,
		"wpbot":       true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", nil, err
	}

	resp, err := http.Post("https://go-pjatk-server.fly.dev/api/game", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	playerToken := resp.Header.Get("x-auth-token")

	return playerToken, data["coords"].([]string), nil
}

func GetLobbyInfo() ([]Player, string, error) {
	resp, err := http.Get("https://go-pjatk-server.fly.dev/api/lobby")
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var result interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, "", err
	}

	var lobbyInfo []Player
	switch result := result.(type) {
	case []interface{}:
		err = json.Unmarshal(body, &lobbyInfo)
		if err != nil {
			return nil, "", err
		}

	case map[string]interface{}:
		var singlePlayer Player
		err = json.Unmarshal(body, &singlePlayer)
		if err != nil {
			return nil, "", err
		}
		lobbyInfo = append(lobbyInfo, singlePlayer)
	default:
		return nil, "", fmt.Errorf("unexpected type %T", result)
	}

	return lobbyInfo, string(body), nil
}

func RefreshLobby(authToken string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/refresh", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("X-Auth-Token", authToken)

	for i := 0; i < 3; i++ { // Retry up to 3 times
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			return nil
		} else if resp.StatusCode == http.StatusServiceUnavailable {
			time.Sleep(5 * time.Second) // Wait for 5 seconds before retrying
		} else {
			return fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
		}
	}

	return fmt.Errorf("received non-OK status code: 503 after 3 retries")
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

	var js json.RawMessage
	err = json.Unmarshal(body, &js)
	if err != nil {
		return string(body), nil
	}

	var playerStats map[string]map[string]interface{}
	err = json.Unmarshal(body, &playerStats)
	if err != nil {
		return "", err
	}

	playerStatsStr, err := json.Marshal(playerStats)
	if err != nil {
		return "", err
	}

	return string(playerStatsStr), nil
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

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
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
			Message string   `jsonxdd:"message"`
			Board   []string `jsonxdd:"board"`
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

func GetStats() ([]PlayerStats, error) {
	resp, err := http.Get("https://go-pjatk-server.fly.dev/api/stats")
	if err != nil {
		return nil, fmt.Errorf("error getting stats: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var statsResponse StatsResponse
	err = json.Unmarshal(body, &statsResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return statsResponse.Stats, nil
}

func AbandonGame(playerToken string) error {
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", "https://go-pjatk-server.fly.dev/api/game/abandon", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("X-Auth-Token", playerToken)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	return nil
}
