main.go: Inits ui

client/menus.go: Display menus

client/elements.go: Defines what should be dipslayed in menus + global text configs

client/operations.go: Launching game and functions that run before boards launch

client/board.go: Board operations, from making ship layout to shooting and displaying

board/gui.go: Init board with config

client/bomBot.go: Custom bot functions, start bomBot game and bot shooting logic

client/helpers.go: Variety of functions used in multiple parts of the code

client/requests.go: Server requests

client/retry.go: Functions for retrying server requests on non 200 responses
