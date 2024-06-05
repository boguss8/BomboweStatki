package main

import (
	"BomboweStatki/client"
	"fmt"
	"os"
	"os/exec"
)

func DisplayOptions() {
	fmt.Println("Choose an option:")
	fmt.Println("1. Add yourself to the lobby")
	fmt.Println("2. Challenge an opponent")
	fmt.Println("3. Change your ship layout")
	fmt.Println("4. Show top 10 players")
	fmt.Println("5. Search for a player")
}

func main() {
	for {
		DisplayOptions()

		var choice int
		for {
			fmt.Print("Enter your choice: ")
			_, err := fmt.Scanln(&choice)
			if err != nil || choice < 1 || choice > 6 {
				DisplayOptions()
				continue
			}
			break
		}

		switch choice {
		case 1:
			client.AddToLobby()
			DisplayOptions()
		case 2:
			client.ChallengeOpponent()
			DisplayOptions()
		case 3:
			clearScreen()
			client.ChangeShipLayout()
		case 4:
			stats, err := client.GetStats()
			if err != nil {
				fmt.Println(err)
				continue
			}
			client.DisplayStats(stats)
		case 5:
			fmt.Print("Enter the player's nickname: ")
			var nick string
			_, err := fmt.Scanln(&nick)
			if err != nil {
				fmt.Println("Error reading nickname:", err)
				continue
			}
		}
	}
}

func DisplayPlayerStats(player client.PlayerStats) {
	fmt.Printf("Nick: %s, Points: %d, Wins: %d, Games: %d, Rank: %d\n", player.Nick, player.Points, player.Wins, player.Games, player.Rank)
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
