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

// MaxScore will return the maximum score across all players
func (players Players) MaxScore() int {
	maxScore := 0

	for _, p := range players {
		if p.GetPoints() > maxScore {
			maxScore = p.GetPoints()
		}
	}

	return maxScore
}
