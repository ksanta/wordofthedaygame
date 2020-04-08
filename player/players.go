package player

type Players []*Player

func (players Players) NumActivePlayers() int {
	activePlayers := 0

	for _, p := range players {
		if p.Active {
			activePlayers++
		}
	}
	return activePlayers
}

// ForActivePlayers will repeat the given function for active players only
func (players Players) ForActivePlayers(funcToDo func(p *Player)) {
	for _, p := range players {
		if p.Active {
			funcToDo(p)
		}
	}
}

// PlayerWithHighestPoints returns the player with the maximum points. They may not have actually won yet.
func (players Players) PlayerWithHighestPoints() *Player {
	maxScore := 0
	var winner *Player

	for _, p := range players {
		if p.GetPoints() > maxScore {
			maxScore = p.GetPoints()
			winner = p
		}
	}

	return winner
}
