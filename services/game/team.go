package game

import pbgame "github.com/ilovewangli1314/OnlineGame_Server_1/protos/game"

type (
	// Team is struct for team
	Team struct {
		data  *pbgame.Team
		heros []*Hero

		turnIdx int
	}
)

// NewTeam return a team
func NewTeam(data *pbgame.Team) *Team {
	team := &Team{data: data}
	team.heros = make([]*Hero, 0)
	for _, heroData := range data.Heros {
		hero := NewHero(heroData)
		team.addHero(hero)
	}

	return team
}

func (t *Team) addHero(hero *Hero) {
	t.heros = append(t.heros, hero)
	hero.belongTeam = t
}
func (t *Team) onHeroDie(hero *Hero) {
	idx := -1
	for i, v := range t.heros {
		if v == hero {
			idx = i
		}
	}
	t.heros = append(t.heros[:idx], t.heros[idx+1:]...)
}

func (t *Team) runAction(team *Team) *pbgame.Action {
	if t.isRoundEnd() {
		return nil
	}

	turnHero := t.getTurn()
	targetHero := team.getHead()
	if turnHero != nil && targetHero != nil {
		// create action protobuf data for all players
		pbAction := &pbgame.Action{
			SrcHeroId:    int32(turnHero.data.Id),
			TargetHeroId: int32(targetHero.data.Id),
		}

		t.heros[t.turnIdx].attack(targetHero)
		t.turnIdx++

		return pbAction
	}

	return nil
}

func (t *Team) getTurn() *Hero {
	turnIdx := t.turnIdx
	if turnIdx >= 0 && turnIdx < len(t.heros) {
		return t.heros[turnIdx]
	}
	return nil
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
