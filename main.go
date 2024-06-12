package main

import (
	"BomboweStatki/client"
	"context"

	gui "github.com/s25867/warships-gui/v2"
)

func main() {
	ui := gui.NewGUI(false)
	ctx := context.Background()
	go client.MainMenu(ui)
	ui.Start(ctx, nil)
}
