package game

import (
	"context"
	"sync"

	"github.com/topfreegames/pitaya/session"

	"github.com/google/uuid"
	pbcommon "github.com/ilovewangli1314/OnlineGame_Server_1/protos"
	pbgame "github.com/ilovewangli1314/OnlineGame_Server_1/protos/game"
	"github.com/topfreegames/pitaya"

	"github.com/topfreegames/pitaya/component"
	"github.com/topfreegames/pitaya/config"
	"github.com/topfreegames/pitaya/groups"
	"github.com/topfreegames/pitaya/timer"
)

type (
	// Entry is struct for players enter game
	Entry struct {
		component.Base
		timer        *timer.Timer
		waittingUids []string
		uniqueGameID int
		mutex        sync.Mutex

		Stats *Stats
	}

	// Stats exports the game status
	Stats struct {
		outboundBytes int
		inboundBytes  int
	}
)

// Outbound gets the outbound status
func (Stats *Stats) Outbound(ctx context.Context, in []byte) ([]byte, error) {
	Stats.outboundBytes += len(in)
	return in, nil
}

// Inbound gets the inbound status
func (Stats *Stats) Inbound(ctx context.Context, in []byte) ([]byte, error) {
	Stats.inboundBytes += len(in)
	return in, nil
}

// NewEntry returns a new entry
func NewEntry() *Entry {
	return &Entry{
		waittingUids: make([]string, 0),
		Stats:        &Stats{},
	}
}

// Init runs on service initialization
func (e *Entry) Init() {
	gsi := groups.NewMemoryGroupService(config.NewConfig())
	pitaya.InitGroups(gsi)
}

// Join handle user join game request
func (e *Entry) Join(ctx context.Context) (*pbcommon.Response, error) {
	fakeUID := uuid.New().String() // every join use a new uuid !!!
	s := pitaya.GetSessionFromCtx(ctx)
	err := s.Bind(ctx, fakeUID) // binding uid to current session
	if err != nil {
		return &pbcommon.Response{Code: 1}, err
	}

	e.mutex.Lock()
	e.waittingUids = append(e.waittingUids, fakeUID)
	if len(e.waittingUids) == 2 { // begin game when players is enough
		game := NewGame(ctx, e.uniqueGameID, e.waittingUids)
		e.uniqueGameID++

		for _, uid := range e.waittingUids {
			session.GetSessionByUID(uid).Set("game", game)
		}
		e.waittingUids = e.waittingUids[0:0]
	}
	e.mutex.Unlock()

	return &pbcommon.Response{Code: 0}, nil
}

// UseSkill handle player use skill request
func (e *Entry) UseSkill(ctx context.Context, msg *pbgame.UseSkill) (*pbcommon.Response, error) {
	s := pitaya.GetSessionFromCtx(ctx)
	game := s.Get("game").(*Game)
	if game != nil {
		return game.UseSkill(ctx, msg)
	}

	return &pbcommon.Response{Code: 0}, nil
}

// RestartGame will restart a new game
func (e *Entry) RestartGame(ctx context.Context) (*pbcommon.Response, error) {
	s := pitaya.GetSessionFromCtx(ctx)
	game := s.Get("game").(*Game)

	// create a new game instance
	e.mutex.Lock()
	newGame := NewGame(ctx, e.uniqueGameID, game.uids)
	e.uniqueGameID++
	e.mutex.Unlock()

	for _, uid := range newGame.uids {
		session.GetSessionByUID(uid).Set("game", newGame)
	}

	game.stopGame()
	return &pbcommon.Response{Code: 0}, nil
}
