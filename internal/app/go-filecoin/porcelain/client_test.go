package porcelain_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/go-filecoin/internal/app/go-filecoin/porcelain"
	"github.com/filecoin-project/go-filecoin/internal/pkg/block"
	e "github.com/filecoin-project/go-filecoin/internal/pkg/enccid"
	"github.com/filecoin-project/go-filecoin/internal/pkg/encoding"
	"github.com/filecoin-project/go-filecoin/internal/pkg/types"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/actor"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/actor/builtin/miner"
	vmaddr "github.com/filecoin-project/go-filecoin/internal/pkg/vm/address"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/state"

	tf "github.com/filecoin-project/go-filecoin/internal/pkg/testhelpers/testflags"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type claPlumbing struct {
	actorFail   bool
	actorChFail bool
	messageFail bool

	MinerAddress address.Address
}

func (cla *claPlumbing) ActorLs(ctx context.Context) (<-chan state.GetAllActorsResult, error) {
	out := make(chan state.GetAllActorsResult)

	if cla.actorFail {
		return nil, errors.New("ACTOR FAILURE")
	}

	go func() {
		defer close(out)
		for i := 0; i < 42; i++ {
			if cla.actorChFail {
				out <- state.GetAllActorsResult{
					Error: errors.New("ACTOR CHANNEL FAILURE"),
				}
			} else {
				cla.MinerAddress = vmaddr.NewForTestGetter()()
				actor := actor.Actor{Code: e.NewCid(types.MinerActorCodeCid)}
				out <- state.GetAllActorsResult{
					Address: cla.MinerAddress.String(),
					Actor:   &actor,
				}
			}
		}
	}()

	return out, nil
}

func (cla *claPlumbing) ChainHeadKey() block.TipSetKey {
	return block.NewTipSetKey()
}

func (cla *claPlumbing) MessageQuery(ctx context.Context, optFrom, to address.Address, method types.MethodID, _ block.TipSetKey, params ...interface{}) ([][]byte, error) {
	if cla.messageFail {
		return nil, errors.New("MESSAGE FAILURE")
	}

	if method == miner.GetAsks {
		askIDs, _ := encoding.Encode([]types.Uint64{0})
		return [][]byte{askIDs}, nil
	}

	ask := miner.Ask{
		Expiry: types.NewBlockHeight(1),
		ID:     big.NewInt(2),
		Price:  types.NewAttoFILFromFIL(3),
	}
	askBytes, _ := encoding.Encode(ask)
	return [][]byte{askBytes}, nil
}

func TestClientListAsks(t *testing.T) {
	tf.UnitTest(t)

	t.Run("success", func(t *testing.T) {
		t.Skip("Depends on internal vm datastructure (miner.Ask) fixing is a bridge too far")
		ctx := context.Background()
		plumbing := &claPlumbing{}

		results := porcelain.ClientListAsks(ctx, plumbing)
		result := <-results

		expectedResult := porcelain.Ask{
			Expiry: types.NewBlockHeight(1),
			ID:     uint64(2),
			Miner:  plumbing.MinerAddress,
			Price:  types.NewAttoFILFromFIL(3),
		}

		assert.Equal(t, expectedResult, result)
	})

	t.Run("failed actor ls", func(t *testing.T) {
		ctx := context.Background()
		plumbing := &claPlumbing{
			actorFail: true,
		}

		results := porcelain.ClientListAsks(ctx, plumbing)
		result := <-results

		assert.Error(t, result.Error, "ACTOR FAILURE")
	})

	t.Run("failed actor ls via channel", func(t *testing.T) {
		ctx := context.Background()
		plumbing := &claPlumbing{
			actorChFail: true,
		}

		results := porcelain.ClientListAsks(ctx, plumbing)
		result := <-results

		assert.Error(t, result.Error, "ACTOR CHANNEL FAILURE")
	})

	t.Run("failed message query", func(t *testing.T) {
		ctx := context.Background()
		plumbing := &claPlumbing{
			messageFail: true,
		}

		results := porcelain.ClientListAsks(ctx, plumbing)
		result := <-results

		assert.Error(t, result.Error, "MESSAGE FAILURE")
	})
}
