package game

import pbgame "github.com/ilovewangli1314/OnlineGame_Server_1/protos/game"

type (
	// Hero is struct for hero
	Hero struct {
		data       *pbgame.Hero
		belongTeam *Team

		isAttacking bool
	}
)

func (h *Hero) die() {

}

func (h *Hero) attack(target *Hero) {
	target.beAttacked(h)
}

func (h *Hero) beAttacked(src *Hero) {
	reduceHp := src.data.Attack - h.data.Defense
	h.data.Hp -= reduceHp

	if h.data.Hp <= 0 {
		h.die()
	}
}

func (h *Hero) isAlive() bool {
	return h.data.Hp > 0
}
