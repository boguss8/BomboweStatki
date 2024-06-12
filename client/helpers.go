package client

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

type Ship struct {
	Coords          []string
	SurroundingArea []string
	IsDestroyed     string
}

var shipsMutex sync.Mutex

func mapShips(coords []string) map[int]Ship {
	shipsMutex.Lock()
	defer shipsMutex.Unlock()

	ships := make(map[int]Ship)
	shipIndex := 0

	for i := 0; i < len(coords); i++ {
		// Check if a ship with the current coordinate already exists
		shipExists := false
		for _, ship := range ships {
			for _, coord := range ship.Coords {
				if coord == coords[i] {
					shipExists = true
					break
				}
			}
			if shipExists {
				break
			}
		}

		// If the ship doesn't exist, check if the coordinate is adjacent to an existing ship
		if !shipExists {
			adjacentShips := make([]int, 0)
			for idx, ship := range ships {
				isAdjacent, err := isAdjacentShip(coords[i], ship.Coords, 1)
				if err != nil {
					continue
				}
				if isAdjacent {
					adjacentShips = append(adjacentShips, idx)
				}
			}

			if len(adjacentShips) > 1 {
				mergedShip := ships[adjacentShips[0]]
				for _, adjIdx := range adjacentShips[1:] {
					mergedShip.Coords = append(mergedShip.Coords, ships[adjIdx].Coords...)
					delete(ships, adjIdx)
				}
				mergedShip.Coords = append(mergedShip.Coords, coords[i])
				ships[adjacentShips[0]] = mergedShip
			} else if len(adjacentShips) == 1 {
				// If the coordinate is adjacent to exactly one ship, add it to that ship
				ship := ships[adjacentShips[0]]
				ship.Coords = append(ship.Coords, coords[i])
				ships[adjacentShips[0]] = ship
			} else {
				// If the coordinate is not adjacent to any ship, create a new ship
				ship := Ship{Coords: []string{coords[i]}}
				surroundingArea := getSurroundingCoords(coords[i])
				ship.SurroundingArea = removeDuplicates(surroundingArea)
				ship.SurroundingArea = removeShipCoordsFromArea(ship.Coords, ship.SurroundingArea)
				ships[shipIndex] = ship
				shipIndex++
			}
		}
	}

	// Update surrounding area for all ships
	for idx, ship := range ships {
		surroundingArea := make([]string, 0)
		for _, coord := range ship.Coords {
			surroundingArea = append(surroundingArea, getSurroundingCoords(coord)...)
		}
		surroundingArea = removeDuplicates(surroundingArea)
		surroundingArea = removeShipCoordsFromArea(ship.Coords, surroundingArea)
		ship.SurroundingArea = surroundingArea
		ships[idx] = ship
	}

	return ships
}

func removeShipCoordsFromArea(shipCoords []string, area []string) []string {
	shipCoordsMap := make(map[string]bool)
	for _, coord := range shipCoords {
		shipCoordsMap[coord] = true
	}

	result := make([]string, 0)
	for _, coord := range area {
		if !shipCoordsMap[coord] {
			result = append(result, coord)
		}
	}

	return result
}

func getSurroundingCoords(coord string) []string {
	col := int(coord[0] - 'A')
	row, _ := strconv.Atoi(coord[1:])
	surroundingCoords := make([]string, 0)

	// Generate surrounding coordinates
	for i := col - 1; i <= col+1; i++ {
		for j := row - 1; j <= row+1; j++ {
			if i >= 0 && i <= 9 && j >= 1 && j <= 10 {
				surroundingCoords = append(surroundingCoords, fmt.Sprintf("%c%d", 'A'+i, j))
			}
		}
	}

	return surroundingCoords
}

func removeDuplicates(slice []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range slice {
		if !encountered[slice[v]] {
			encountered[slice[v]] = true
			result = append(result, slice[v])
		}
	}

	return result
}

func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	runeStr := []rune(s)

	for i := 0; i < len(runeStr); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runeStr) {
			nn = len(runeStr)
		}
		chunks = append(chunks, string(runeStr[i:nn]))
	}
	return chunks
}

// Helper function to find the index of a string in a slice
func findIndex(slice []string, str string) int {
	for i, item := range slice {
		if item == str {
			return i
		}
	}
	return -1
}

func isAdjacentShip(char string, ship []string, mode int) (bool, error) {
	if len(ship) == 0 {
		return false, nil
	}

	if len(char) == 0 {
		return false, errors.New("invalid coordinate: empty string")
	}

	x := int(char[0] - 'A')
	y, err := strconv.Atoi(char[1:])
	if err != nil {
		return false, fmt.Errorf("error converting row: %v", err)
	}

	for _, c := range ship {
		if len(c) < 2 {
			return false, fmt.Errorf("invalid ship coordinate: %s", c)
		}
		cx := int(c[0] - 'A')
		cy, err := strconv.Atoi(c[1:])
		if err != nil {
			return false, fmt.Errorf("error converting ship coordinate: %v", err)
		}

		// Checks if the coordinates are adjacent horizontally or vertically
		if mode == 1 && ((cx == x && (cy == y+1 || cy == y-1)) || (cy == y && (cx == x+1 || cx == x-1))) {
			return true, nil
		}
		// Checks if the coordinates are adjacent horizontally, vertically, or diagonally
		if mode == 2 && (abs(cx-x) <= 1 && abs(cy-y) <= 1) {
			return true, nil
		}
	}
	return false, nil
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
