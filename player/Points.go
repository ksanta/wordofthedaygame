package player

type Points struct {
	points int
}

func (p *Points) AddPoints(points int) {
	p.points += points
}

func (p *Points) GetPoints() int {
	return p.points
}
