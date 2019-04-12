package game

import (
	"context"
	"sync"

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
func (r *Entry) Init() {
	gsi := groups.NewMemoryGroupService(config.NewConfig())
	pitaya.InitGroups(gsi)
}

// Join handle user join game request
func (r *Entry) Join(ctx context.Context) (*pbcommon.Response, error) {
	fakeUID := uuid.New().String() // every join use a new uuid !!!
	s := pitaya.GetSessionFromCtx(ctx)
	err := s.Bind(ctx, fakeUID) // binding uid to current session
	if err != nil {
		return &pbcommon.Response{Code: 1}, err
	}

	r.mutex.Lock()
	r.waittingUids = append(r.waittingUids, fakeUID)
	if len(r.waittingUids) == 2 { // begin game when players is enough
		game := NewGame(ctx, r.waittingUids, r.uniqueGameID)
		s.Set("game", game)

		r.waittingUids = r.waittingUids[0:0]
	}
	r.mutex.Unlock()

	return &pbcommon.Response{Code: 0}, nil
}

// UseSkill handle player use skill request
func (r *Entry) UseSkill(ctx context.Context, msg *pbgame.UseSkill) (*pbcommon.Response, error) {
	s := pitaya.GetSessionFromCtx(ctx)
	game := s.Get("game").(*Game)
	if game != nil {
		return game.UseSkill(ctx, msg)
	}

	return &pbcommon.Response{Code: 0}, nil
}
