package https_requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func InitGame() (map[string]interface{}, string, error) {
	data := map[string]interface{}{
		"coords": []string{"A2", "A4", "B9", "C7", "D1", "D2", "D3", "D4", "D7", "E7", "F1", "F2", "F3", "F5", "G5", "G8", "G9", "I4", "J4", "J8"},
		"desc":   "pierwszy raz",
		"nick":   "Jan_Niecny",
		"wpbot":  true,
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

func GetBoardInfo(playerToken string) ([]string, error) {
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
