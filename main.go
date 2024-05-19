package main

import (
	"BomboweStatki/client"
	"fmt"
)

func main() {
	for {
		fmt.Println("Choose an option:")
		fmt.Println("1. Add yourself to the lobby")
		fmt.Println("2. Challenge an opponent")

		var choice int
		for {
			fmt.Print("Enter your choice: ")
			_, err := fmt.Scanln(&choice)
			if err != nil {
				fmt.Println("Error reading choice:", err)
				fmt.Println("Please enter 1 or 2.")
				continue
			}
			if choice != 1 && choice != 2 {
				fmt.Println("Invalid choice. Please enter 1 or 2.")
				continue
			}
			break
		}

		switch choice {
		case 1:
			client.AddToLobby()
		case 2:
			client.ChallengeOpponent()
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
