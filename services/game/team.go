package game

type (
	// Team is struct for team
	Team struct {
		heros   []*Hero
		turnIdx int
	}
)

func (t *Team) runAction(team *Team) {
	targetHero := team.getHead()
	if !t.isRoundEnd() && targetHero != nil {
		t.heros[t.turnIdx].attack(targetHero)
	}
}

func (t *Team) getHead() *Hero {
	if len(t.heros) > 0 {
		return t.heros[0]
	}

	return nil
}

func (t *Team) isRoundEnd() bool {
	return t.turnIdx >= len(t.heros)
}
func (t *Team) beginRound() {
	t.turnIdx = 0
}

func (t *Team) isEmpty() bool {
	return len(t.heros) <= 0
}
