package game

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/ilovewangli1314/OnlineGame_Server_1/services/common"

	pbgame "github.com/ilovewangli1314/OnlineGame_Server_1/protos/game"

	pbcommon "github.com/ilovewangli1314/OnlineGame_Server_1/protos"
	"github.com/topfreegames/pitaya"
	"github.com/topfreegames/pitaya/timer"
)

type (
	// Game represents a component that contains a bundle of game related handler
	// like Join/Message
	Game struct {
		gameID    int
		groupName string

		teams       []*Team
		round       int
		turnTeamIdx int
		bgCtx       context.Context

		seededRandom *common.SeededRandom
		timer        *timer.Timer
	}
)

// NewGame returns a new game
func NewGame(ctx context.Context, uids []string, gameID int) *Game {
	// for _, uid := range uids {
	// 	GetSessionByUID(uid)
	// }
	game := &Game{
		gameID:    gameID,
		groupName: fmt.Sprintf("game/game/%d", gameID),
	}
	teams := make([]*Team, 0)
	teams = append(teams, game.createTeam(0))
	teams = append(teams, game.createTeam(10000))
	game.teams = teams

	game.bgCtx = context.Background()
	// create group for game
	pitaya.GroupCreate(game.bgCtx, game.groupName)
	for _, uid := range uids {
		pitaya.GroupAddMember(ctx, game.groupName, uid)
	}
	// push game begin to all players in group
	game.seededRandom = &common.SeededRandom{RandomSeed: int32(time.Now().Unix())}
	pbScene := &pbgame.Scene{
		RandomSeed: game.seededRandom.RandomSeed,
		Teams:      game.getPbTeams(),
	}
	pitaya.GroupBroadcast(game.bgCtx, "game", game.groupName, "onGameBegin", pbScene)

	// begin timer for game actions
	game.timer = pitaya.NewTimer(time.Second, func() {
		game.runAction()
	})

	return game
}

func (g *Game) createTeam(heroBaseIdx int) *Team {
	pbTeam := &pbgame.Team{}
	pbTeam.Heros = make([]*pbgame.Hero, 0)
	for idx := 0; idx < 6; idx++ {
		pbHero := &pbgame.Hero{
			Id:      int32(heroBaseIdx + idx),
			Hp:      int32(30 + math.Round(5*rand.Float64())),
			Mp:      int32(100 + math.Round(5*rand.Float64())),
			Attack:  int32(10 + math.Round(5*rand.Float64())),
			Defense: int32(3 + math.Round(2*rand.Float64())),
		}
		pbTeam.Heros = append(pbTeam.Heros, pbHero)
	}

	return NewTeam(pbTeam)
}

func (g *Game) getPbTeams() []*pbgame.Team {
	pbTeams := make([]*pbgame.Team, 0)
	for _, team := range g.teams {
		pbTeams = append(pbTeams, team.data)
	}

	return pbTeams
}

func reply(code int32, msg string) *pbcommon.Response {
	return &pbcommon.Response{
		Code: code,
		Msg:  msg,
	}
}

// // Entry is the entrypoint
// func (g *Game) Entry(ctx context.Context) (*pbcommon.Response, error) {
// 	fakeUID := uuid.New().String() // just use s.ID as uid !!!
// 	s := pitaya.GetSessionFromCtx(ctx)
// 	err := s.Bind(ctx, fakeUID) // binding session uid
// 	if err != nil {
// 		return nil, pitaya.Error(err, "ENT-000")
// 	}
// 	return reply(200, "ok"), nil
// }

// // Join game
// func (g *Game) Join(ctx context.Context) (*pbcommon.Response, error) {
// 	fakeUID := uuid.New().String() // just use s.ID as uid !!!
// 	s := pitaya.GetSessionFromCtx(ctx)
// 	err := s.Bind(ctx, fakeUID) // binding session uid
// 	if err != nil {
// 		return nil, pitaya.Error(err, "ENT-000")
// 	}

// 	// members, err := pitaya.GroupMembers(ctx, "game")
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// s.Push("onMembers", &protos.AllMembers{Members: members})
// 	// pitaya.GroupBroadcast(ctx, "connector", "game", "onNewUser", &protos.NewUser{Content: fmt.Sprintf("New user: %d", s.ID())})
// 	// pitaya.GroupBroadcast(ctx, "game", "game", "onNewUser", &protos.NewUser{Content: fmt.Sprintf("New user: %d", s.ID())})
// 	pitaya.GroupAddMember(ctx, "game", s.UID())
// 	s.OnClose(func() {
// 		pitaya.GroupRemoveMember(ctx, "game", s.UID())

// 		// Fixme: 临时处理
// 		// members, err := pitaya.GroupMembers(ctx, "game")
// 		// if err == nil {
// 		// 	pitaya.GroupBroadcast(ctx, "game", "game", "onMembers", &protos.AllMembers{Members: members})
// 		// }
// 	})
// 	return &pbcommon.Response{Code: 0}, nil
// }

// // Message sync last message to all members
// func (g *Game) Message(ctx context.Context, msg *protos.UserMessage) {
// 	err := pitaya.GroupBroadcast(ctx, "connector", "game", "onMessage", msg)
// 	if err != nil {
// 		fmt.Println("error broadcasting message", err)
// 	}
// }

func (g *Game) runAction() {
	teams := g.teams
	teamCnt := len(teams)

	isRoundEnd := teams[g.turnTeamIdx].isRoundEnd()
	roundBegin := false
	if isRoundEnd {
		// 队伍数组中后出手的队伍放在后面，据此判断后出手队伍结束动作后一个回合结束
		if g.turnTeamIdx == len(teams)-1 {
			g.round++
			roundBegin = true
		}

		g.turnTeamIdx = (g.turnTeamIdx + 1) % teamCnt
	}

	srcTeam := teams[g.turnTeamIdx]
	if srcTeam.isEmpty() {
		g.timer.Stop()
		return
	}

	targetTeam := teams[(g.turnTeamIdx+1)%teamCnt]
	if targetTeam.isEmpty() {
		g.timer.Stop()
		return
	}

	// 回合开始时重置双方队伍为回合开始状态
	if roundBegin {
		srcTeam.beginRound()
		targetTeam.beginRound()
	}

	pbAction := srcTeam.runAction(targetTeam)
	// push action info to all players
	pitaya.GroupBroadcast(g.bgCtx, "game", g.groupName, "onRunAction", pbAction)
}

// UseSkill can use a skill to someone
func (g *Game) UseSkill(ctx context.Context) (*pbcommon.Response, error) {
	// s := pitaya.GetSessionFromCtx(ctx)
	// s.Get("game").Us

	return &pbcommon.Response{Code: 0}, nil
}

// // SendRPC sends rpc
// func (g *Game) SendRPC(ctx context.Context, msg []byte) (*pbcommon.Response, error) {
// 	ret := pbcommon.Response{}
// 	err := pitaya.RPC(ctx, "connector.connectorremote.remotefunc", &ret, &pbcommon.RPCMsg{})
// 	if err != nil {
// 		return nil, pitaya.Error(err, "RPC-000")
// 	}
// 	return reply(200, ret.Msg), nil
// }
