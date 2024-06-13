package client

import (
	"context"

	gui "github.com/s25867/warships-gui/v2"
)

func MainMenu(ui *gui.GUI) {
	ui.NewScreen("menu")
	ui.SetScreen("menu")

	menuUi := MainMenuElements(ui)

	ctx := context.Background()

	go printTopPlayers(ui, 30, 3)

	// Handle button clicks
	for {
		clicked := menuUi.ButtonArea.Listen(ctx)
		switch clicked {
		case "pvpButtton":
			timerContext, cancelTimer := context.WithCancel(context.Background())
			reset := make(chan bool)
			pvpMenu(ui, "", timerContext, cancelTimer, reset)
			return
		case "botButtton":

			go botMenu(ui)
			return
		case "profileButton":
			go profileMenu(ui)
		}
	}
}
func profileMenu(ui *gui.GUI) {
	ui.NewScreen("profile")
	ui.SetScreen("profile")

	profileUi := ProfileElements(ui)

	ctx := context.Background()

	printPlayerStats(ui, DefaultGameInitData.Nick, 80, 11)

	// Handle button clicks
	for {
		clicked := profileUi.ButtonArea.Listen(ctx)
		switch clicked {
		case "returnButton":

			go MainMenu(ui)
			return
		case "boardButton":
			go func() {
				changeShipLayout(ui)
			}()
			return
		case "editNameButton":
			name := drawNameField(ui, ctx)
			if name == "" {
				go profileMenu(ui)
				continue
			}
			DefaultGameInitData.Nick = name
			go profileMenu(ui)

		case "editDescButton":
			desc := drawDescField(ui, ctx)
			if desc == "" {
				go profileMenu(ui)
				continue
			}
			DefaultGameInitData.Desc = desc
			go profileMenu(ui)
		case "randomBoardButton":
			DefaultGameInitData.Coords = generateRandomBoard()
			go profileMenu(ui)
		}
	}
}
func botMenu(ui *gui.GUI) {
	botUi := BotElements(ui)

	ctx := context.Background()

	for {
		clicked := botUi.ButtonArea.Listen(ctx)
		switch clicked {
		case "returnButton":
			go MainMenu(ui)
			return
		case "wpBotButton":
			gameData := GameInitData{
				TargetNick: "",
				Wpbot:      true,
			}
			StartGame(ui, gameData)
		case "bomBotButton":
			bomBotInit(ui, DefaultGameInitData)
			return
		}
	}
}

var isWaitingForChallenger bool

func pvpMenu(ui *gui.GUI, playerToken string, timerContext context.Context, cancelTimer context.CancelFunc, reset chan bool) {
	ui.NewScreen("lobby")
	ui.SetScreen("lobby")
	// get lobby info
	lobbyInfo, _, err := retryOnErrorWithPlayers(ui, func() ([]Player, string, error) {
		return GetLobbyInfo()
	})
	if err != nil {
		ui.Draw(gui.NewText(2, 0, "Error getting lobby info: "+err.Error(), errorText))
	}

	lobbyUi := LobbyElements(ui, lobbyInfo)

	ctx := context.Background()
	ui.Draw(gui.NewText(2, 9, "Click on an opponent to challenge him into a duel!", defaultText))
	for {
		//Listen what button was clicked
		clicked := lobbyUi.ButtonArea.Listen(ctx)
		switch clicked {
		case "addYourselfButton":
			// Adds player to lobby and starts the timer
			gameData := GameInitData{
				TargetNick: "",
				Wpbot:      false,
			}
			playerToken, err = retryOnError(ui, func() (string, error) {
				return InitGame(gameData)
			})
			if err != nil {
				ui.Draw(gui.NewText(1, 29, "Error initializing game: "+err.Error()+". Retrying...", errorText))
			}
			timerContext, cancelTimer := context.WithCancel(context.Background())

			go lobbyTimer(timerContext, reset, ui)
			go waitForStart(ui, playerToken, gameData, cancelTimer)
			go pvpMenu(ui, playerToken, timerContext, cancelTimer, reset)
		case "resetLobbyTimerButton":
			// If user is in lobby reset the timer
			if playerToken == "" {
				go pvpMenu(ui, playerToken, timerContext, cancelTimer, reset)
				ui.Draw(gui.NewText(0, 0, "You are not in lobby", errorText))
			} else {
				reset <- true
				RefreshLobby(playerToken)
				go pvpMenu(ui, playerToken, timerContext, cancelTimer, reset)
				ui.Draw(gui.NewText(2, 3, "Lobby timer reset", defaultText))
			}

		case "returnButton":
			// Stops the timer and returns to main menu
			if cancelTimer != nil {
				cancelTimer()
			}

			go MainMenu(ui)
			return
		case "refreshButton":
			// Refreshes the lobby
			go pvpMenu(ui, playerToken, timerContext, cancelTimer, reset)
		default:
			// If player is not in lobby and clicked on a player, challenge him
			if !isWaitingForChallenger {
				for _, player := range lobbyInfo {
					if player.Nick == clicked {
						if DefaultGameInitData.Nick == player.Nick {
							ui.Draw(gui.NewText(2, 2, "You can't duel yourself...", errorText))
							continue
						}

						DefaultGameInitData.TargetNick = player.Nick
						DefaultGameInitData.Wpbot = false
						StartGame(ui, DefaultGameInitData)

						break
					}
				}
			} else {
				ui.Draw(gui.NewText(2, 2, "You are already waiting for a challenger...", errorText))
			}
		}
	}
}
