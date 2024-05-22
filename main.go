package main

import (
	"BomboweStatki/client"
	"encoding/json"
	"fmt"
)

func main() {
	for {
		fmt.Println("Choose an option:")
		fmt.Println("1. Add yourself to the lobby")
		fmt.Println("2. Challenge an opponent")
		fmt.Println("3. Show top 10 players")
		fmt.Println("4. Search for a player")

		var choice int
		for {
			fmt.Print("Enter your choice: ")
			_, err := fmt.Scanln(&choice)
			if err != nil {
				fmt.Println("Please enter 1, 2, 3 or 4.")
				continue
			}
			if choice != 1 && choice != 2 && choice != 3 && choice != 4 {
				fmt.Println("Invalid choice. Please enter 1, 2, 3 or 4.")
				continue
			}
			break
		}

		switch choice {
		case 1:
			client.AddToLobby()
		case 2:
			client.ChallengeOpponent()
		case 3:
			stats, err := client.GetStats()
			if err != nil {
				fmt.Println(err)
				break
			}
			client.DisplayStats(stats)
		case 4:
			fmt.Print("Enter the player's nickname: ")
			var nick string
			_, err := fmt.Scanln(&nick)
			if err != nil {
				fmt.Println("Error reading nickname:", err)
				break
			}
			playerStats, err := client.GetPlayerStats(nick)
			if err != nil {
				fmt.Println(err)
				break
			}
			var statsMap map[string]client.PlayerStats
			err = json.Unmarshal([]byte(playerStats), &statsMap)
			if err != nil {
				fmt.Println("Error parsing player stats:", err)
				break
			}
			player := statsMap["stats"]
			DisplayPlayerStats(player)
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func DisplayPlayerStats(player client.PlayerStats) {
	fmt.Printf("Nick: %s, Points: %d, Wins: %d, Games: %d, Rank: %d\n", player.Nick, player.Points, player.Wins, player.Games, player.Rank)
}
