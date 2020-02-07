package retrievalmarketconnector_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/piecestore"
	"github.com/filecoin-project/go-fil-markets/shared/tokenamount"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dss "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-hamt-ipld"
	bstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/filecoin-project/go-filecoin/internal/app/go-filecoin/retrieval_market_connector"
	tut "github.com/filecoin-project/go-filecoin/internal/app/go-filecoin/shared_testutils"
	"github.com/filecoin-project/go-filecoin/internal/pkg/block"
	"github.com/filecoin-project/go-filecoin/internal/pkg/chain"
	"github.com/filecoin-project/go-filecoin/internal/pkg/repo"
	th "github.com/filecoin-project/go-filecoin/internal/pkg/testhelpers"
	"github.com/filecoin-project/go-filecoin/internal/pkg/types"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/state"
)

func TestNewRetrievalClientNodeConnector(t *testing.T) {
	ctx := context.Background()

	bs := bstore.NewBlockstore(dss.MutexWrap(datastore.NewMapDatastore()))
	cs := requireNewChainStoreWithBlock(ctx, t)
	rmFake := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))

	ps := tut.RequireMakeTestPieceStore(t)
	rcnc := NewRetrievalClientNodeConnector(bs, cs, rmFake, rmFake, ps, rmFake, rmFake, rmFake, rmFake, rmFake)
	assert.NotNil(t, rcnc)
}

func TestRetrievalClientNodeConnector_GetOrCreatePaymentChannel(t *testing.T) {
	ctx := context.Background()

	bs, cs, ps, clientAddr, minerAddr, channelAmount := testSetup(ctx, t)

	t.Run("Errors if clientWallet get balance fails", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
		tapa.BalanceErr = errors.New("boom")

		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
		res, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
		assert.EqualError(t, err, "boom")
		assert.Equal(t, address.Undef, res)
		tapa.BalanceErr = nil
	})

	t.Run("if the payment channel does not exist", func(t *testing.T) {
		bs := bstore.NewBlockstore(dss.MutexWrap(datastore.NewMapDatastore()))
		cs := requireNewChainStoreWithBlock(ctx, t)

		ps := tut.RequireMakeTestPieceStore(t)

		t.Run("creates a new payment channel registry entry and posts createChannel message", func(t *testing.T) {
			tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
			tapa.StubMessageResponse(t, clientAddr, minerAddr, types.NewAttoFIL(channelAmount.Int))

			rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)

			nonce := tapa.Nonce
			res, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
			assert.NoError(t, err)
			assert.Equal(t, address.Undef, res)
			assert.NotNil(t, tapa.ExpectedPmtChans[minerAddr])
			assert.Equal(t, nonce+1, tapa.Nonce)
		})
		t.Run("Errors if there aren't enough funds in wallet", func(t *testing.T) {
			tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))

			rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)

			res, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, tokenamount.FromInt(2000))
			assert.EqualError(t, err, "not enough funds in wallet")
			assert.Equal(t, address.Undef, res)
		})

		t.Run("Errors if client or minerWallet addr is invalid", func(t *testing.T) {
			tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))

			rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
			_, err := rcnc.GetOrCreatePaymentChannel(ctx, address.Undef, minerAddr, channelAmount)
			assert.EqualError(t, err, "empty address")

			_, err = rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, address.Undef, channelAmount)
			assert.EqualError(t, err, "empty address")

		})
		t.Run("Errors if can't get block height/head tipset", func(t *testing.T) {
			_, _, _, localCs := requireNewEmptyChainStore(ctx, t)

			tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
			rcnc := NewRetrievalClientNodeConnector(bs, localCs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
			res, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
			assert.EqualError(t, err, "Key not found in tipindex")
			assert.Equal(t, address.Undef, res)
		})
	})

	t.Run("if payment channel exists", func(t *testing.T) {
		t.Run("Retrieves existing payment channel address", func(t *testing.T) {
			tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
			tapa.StubMessageResponse(t, clientAddr, minerAddr, types.NewAttoFIL(channelAmount.Int))

			rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
			expectedChID, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
			require.NoError(t, err)
			assert.Equal(t, address.Undef, expectedChID)

			actualChID, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
			require.NoError(t, err)
			assert.NotEqual(t, "", actualChID.String())
			assert.Len(t, tapa.ActualPmtChans, 1)
			assert.Equal(t, tapa.ActualPmtChans[clientAddr].ChannelID, tapa.ExpectedPmtChans[clientAddr].ChannelID)
			assert.Equal(t, tapa.ExpectedPmtChans[clientAddr].ChannelID.Bytes(), actualChID.Bytes())
		})
	})
}

func TestRetrievalClientNodeConnector_AllocateLane(t *testing.T) {
	ctx := context.Background()
	bs, cs, ps, clientAddr, minerAddr, channelAmount := testSetup(ctx, t)

	t.Run("Errors if payment channel does not exist", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)

		addr, err := address.NewIDAddress(12345)
		require.NoError(t, err)
		res, err := rcnc.AllocateLane(addr)
		assert.EqualError(t, err, "no such ChannelID")
		assert.Zero(t, res)
	})
	t.Run("Increments and returns lastLane val", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
		tapa.StubMessageResponse(t, clientAddr, minerAddr, types.NewAttoFIL(channelAmount.Int))

		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
		_, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
		require.NoError(t, err)

		pce, ok := tapa.ActualPmtChans[clientAddr]
		require.True(t, ok)

		gochid, err := address.NewFromBytes(pce.ChannelID.Bytes())
		require.NoError(t, err)
		lane, err := rcnc.AllocateLane(gochid)
		require.NoError(t, err)
		assert.Equal(t, lane, tapa.ExpectedLanes[pce.ChannelID])
	})
}

func TestRetrievalClientNodeConnector_CreatePaymentVoucher(t *testing.T) {
	ctx := context.Background()
	bs, cs, ps, clientAddr, minerAddr, channelAmount := testSetup(ctx, t)

	t.Run("Returns a voucher with a signature", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
		tapa.StubMessageResponse(t, clientAddr, minerAddr, types.NewAttoFIL(channelAmount.Int))

		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
		_, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
		require.NoError(t, err)

		chid, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
		require.NoError(t, err)
		lane, err := rcnc.AllocateLane(chid)
		require.NoError(t, err)

		expectedAmt := tokenamount.FromInt(100)

		currentNonce := tapa.Nonce
		voucher, err := rcnc.CreatePaymentVoucher(ctx, chid, expectedAmt, lane)
		require.NoError(t, err)
		assert.Equal(t, expectedAmt, voucher.Amount)
		assert.Equal(t, lane, voucher.Lane)
		assert.Equal(t, currentNonce+1, voucher.Nonce)
		assert.NotNil(t, voucher.Signature)
	})

	t.Run("Errors if payment channel does not exist", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
		tapa.StubMessageResponse(t, clientAddr, minerAddr, types.NewAttoFIL(channelAmount.Int))

		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
		_, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
		require.NoError(t, err)

		chid, err := address.NewIDAddress(rand.Uint64())
		require.NoError(t, err)
		voucher, err := rcnc.CreatePaymentVoucher(ctx, chid, tokenamount.FromInt(100), 1)
		assert.EqualError(t, err, "no such ChannelID")
		assert.Nil(t, voucher)
	})
	t.Run("Errors if there aren't enough funds in payment channel", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
		tapa.StubMessageResponse(t, clientAddr, minerAddr, types.ZeroAttoFIL)
		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
		_, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
		require.NoError(t, err)
		chid, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, tokenamount.FromInt(0))
		require.NoError(t, err)
		lane, err := rcnc.AllocateLane(chid)
		require.NoError(t, err)

		voucher, err := rcnc.CreatePaymentVoucher(ctx, chid, tokenamount.FromInt(201), lane)
		assert.EqualError(t, err, "not enough funds in payment channel")
		assert.Nil(t, voucher)
	})

	t.Run("Errors if payment channel lane doesn't exist", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))

		channelAmt := tokenamount.FromInt(200)
		tapa.StubMessageResponse(t, clientAddr, minerAddr, types.NewAttoFIL(channelAmt.Int))

		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
		_, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmt)
		require.NoError(t, err)
		chid, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, tokenamount.FromInt(0))
		require.NoError(t, err)

		voucher, err := rcnc.CreatePaymentVoucher(ctx, chid, tokenamount.FromInt(50), 0)
		assert.EqualError(t, err, "payment channel has no lanes allocated")
		assert.Nil(t, voucher)
	})

	t.Run("Errors if can't get nonce", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
		tapa.NonceErr = errors.New("no noncense")

		tapa.StubMessageResponse(t, clientAddr, minerAddr, types.NewAttoFIL(channelAmount.Int))
		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
		_, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
		require.NoError(t, err)
		chid, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, tokenamount.FromInt(0))
		require.NoError(t, err)
		lane, err := rcnc.AllocateLane(chid)
		require.NoError(t, err)

		voucher, err := rcnc.CreatePaymentVoucher(ctx, chid, tokenamount.FromInt(1), lane)
		assert.EqualError(t, err, "no noncense")
		assert.Nil(t, voucher)

	})

	t.Run("Errors if can't sign bytes", func(t *testing.T) {
		tapa := NewRetrievalMarketClientFakeAPI(t, tokenamount.FromInt(1000))
		tapa.SigErr = errors.New("signature failure")

		tapa.StubMessageResponse(t, clientAddr, minerAddr, types.NewAttoFIL(channelAmount.Int))
		rcnc := NewRetrievalClientNodeConnector(bs, cs, tapa, tapa, ps, tapa, tapa, tapa, tapa, tapa)
		_, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, channelAmount)
		require.NoError(t, err)
		chid, err := rcnc.GetOrCreatePaymentChannel(ctx, clientAddr, minerAddr, tokenamount.FromInt(0))
		require.NoError(t, err)
		lane, err := rcnc.AllocateLane(chid)
		require.NoError(t, err)

		voucher, err := rcnc.CreatePaymentVoucher(ctx, chid, tokenamount.FromInt(1), lane)
		assert.EqualError(t, err, "signature failure")
		assert.Nil(t, voucher)
	})
}

func testSetup(ctx context.Context, t *testing.T) (bstore.Blockstore, *chain.Store, piecestore.PieceStore, address.Address, address.Address, tokenamount.TokenAmount) {
	bs := bstore.NewBlockstore(dss.MutexWrap(datastore.NewMapDatastore()))
	cs := requireNewChainStoreWithBlock(ctx, t)

	ps := tut.RequireMakeTestPieceStore(t)

	clientAddr, err := address.NewIDAddress(rand.Uint64())
	require.NoError(t, err)
	minerAddr, err := address.NewIDAddress(rand.Uint64())
	require.NoError(t, err)

	channelAmount := tokenamount.FromInt(500)
	return bs, cs, ps, clientAddr, minerAddr, channelAmount
}

func requireNewChainStoreWithBlock(ctx context.Context, t *testing.T) *chain.Store {
	root, builder, genTS, cs := requireNewEmptyChainStore(ctx, t)

	rootBlk := builder.AppendBlockOnBlocks()
	th.RequireNewTipSet(t, rootBlk)
	require.NoError(t, cs.SetHead(ctx, genTS))

	// add tipset and state to chainstore
	require.NoError(t, cs.PutTipSetMetadata(ctx, &chain.TipSetMetadata{
		TipSet:          genTS,
		TipSetStateRoot: root,
		TipSetReceipts:  types.EmptyReceiptsCID,
	}))
	return cs
}

func requireNewEmptyChainStore(ctx context.Context, t *testing.T) (cid.Cid, *chain.Builder, block.TipSet, *chain.Store) {
	cst := hamt.NewCborStore()

	// Cribbed from chain/store_test
	st1 := state.NewTree(cst)
	root, err := st1.Flush(ctx)
	require.NoError(t, err)

	// link testing state to test block
	builder := chain.NewBuilder(t, address.Undef)
	genTS := builder.NewGenesis()
	r := repo.NewInMemoryRepo()

	// setup chain store
	ds := r.Datastore()
	cs := chain.NewStore(ds, cst, state.NewTreeLoader(), chain.NewStatusReporter(), genTS.At(0).Cid())
	return root, builder, genTS, cs
}
