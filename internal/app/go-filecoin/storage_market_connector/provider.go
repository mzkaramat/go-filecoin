package storagemarketconnector

import (
	"context"
	"io"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/shared/tokenamount"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	spasm "github.com/filecoin-project/specs-actors/actors/builtin/market"
	spaminer "github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-filecoin/internal/app/go-filecoin/plumbing/msg"
	"github.com/filecoin-project/go-filecoin/internal/pkg/message"
	"github.com/filecoin-project/go-filecoin/internal/pkg/piecemanager"
	"github.com/filecoin-project/go-filecoin/internal/pkg/types"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/abi"
	fcsm "github.com/filecoin-project/go-filecoin/internal/pkg/vm/actor/builtin/storagemarket"
	vmaddr "github.com/filecoin-project/go-filecoin/internal/pkg/vm/address"
	"github.com/filecoin-project/go-filecoin/internal/pkg/wallet"
)

// StorageProviderNodeConnector adapts the node to provide an interface for the storage provider
type StorageProviderNodeConnector struct {
	connectorCommon

	minerAddr    address.Address
	chainStore   chainReader
	outbox       *message.Outbox
	pieceManager piecemanager.PieceManager
}

var _ storagemarket.StorageProviderNode = &StorageProviderNodeConnector{}

// NewStorageProviderNodeConnector creates a new connector
func NewStorageProviderNodeConnector(ma address.Address,
	cs chainReader,
	ob *message.Outbox,
	w *msg.Waiter,
	pm piecemanager.PieceManager,
	wg WorkerGetter,
	wlt *wallet.Wallet,
) *StorageProviderNodeConnector {
	return &StorageProviderNodeConnector{
		connectorCommon: connectorCommon{cs, w, wlt, ob, wg},
		chainStore:      cs,
		minerAddr:       ma,
		outbox:          ob,
		pieceManager:    pm,
	}
}

// AddFunds sends a message to add storage market collateral for the given address
func (s *StorageProviderNodeConnector) AddFunds(ctx context.Context, addr address.Address, amount tokenamount.TokenAmount) error {
	workerAddr, err := s.GetMinerWorker(ctx, s.minerAddr)
	if err != nil {
		return err
	}

	return s.addFunds(ctx, workerAddr, addr, amount)
}

// EnsureFunds checks the balance for an account and adds funds to the given amount if the balance is insufficient
func (s *StorageProviderNodeConnector) EnsureFunds(ctx context.Context, addr address.Address, amount tokenamount.TokenAmount) error {
	balance, err := s.GetBalance(ctx, addr)
	if err != nil {
		return err
	}

	if !balance.Available.LessThan(amount) {
		return nil
	}

	return s.AddFunds(ctx, addr, tokenamount.Sub(amount, balance.Available))
}

// PublishDeals publishes storage deals on chain
func (s *StorageProviderNodeConnector) PublishDeals(ctx context.Context, deal storagemarket.MinerDeal) (storagemarket.DealID, cid.Cid, error) {
	sig := types.Signature(deal.Proposal.ProposerSignature.Data)

	fcStorageProposal := types.StorageDealProposal{
		PieceRef:  deal.Proposal.PieceRef,
		PieceSize: types.Uint64(deal.Proposal.PieceSize),

		Client:   deal.Proposal.Client,
		Provider: deal.Proposal.Provider,

		ProposalExpiration: types.Uint64(deal.Proposal.ProposalExpiration),
		Duration:           types.Uint64(deal.Proposal.Duration),

		StoragePricePerEpoch: types.Uint64(deal.Proposal.StoragePricePerEpoch.Uint64()),
		StorageCollateral:    types.Uint64(deal.Proposal.StorageCollateral.Uint64()),

		ProposerSignature: &sig,
	}
	params, err := abi.ToEncodedValues([]types.StorageDealProposal{fcStorageProposal})
	if err != nil {
		return 0, cid.Undef, err
	}

	workerAddr, err := s.GetMinerWorker(ctx, s.minerAddr)
	if err != nil {
		return 0, cid.Undef, err
	}

	mcid, cerr, err := s.outbox.Send(
		ctx,
		workerAddr,
		vmaddr.StorageMarketAddress,
		types.ZeroAttoFIL,
		types.NewGasPrice(1),
		types.NewGasUnits(300),
		true,
		fcsm.PublishStorageDeals,
		params,
	)
	if err != nil {
		return 0, cid.Undef, err
	}

	receipt, err := s.wait(ctx, mcid, cerr)
	if err != nil {
		return 0, cid.Undef, err
	}

	dealIDValues, err := abi.Deserialize(receipt.Return[0], abi.UintArray)
	if err != nil {
		return 0, cid.Undef, err
	}

	dealIds, ok := dealIDValues.Val.([]uint64)
	if !ok {
		return 0, cid.Undef, xerrors.New("decoded deal ids are not a []uint64")
	}

	if len(dealIds) < 1 {
		return 0, cid.Undef, xerrors.New("Successful call to publish storage deals did not return deal ids")
	}

	return storagemarket.DealID(dealIds[0]), mcid, err
}

// ListProviderDeals lists all deals for the given provider
func (s *StorageProviderNodeConnector) ListProviderDeals(ctx context.Context, addr address.Address) ([]storagemarket.StorageDeal, error) {
	return s.listDeals(ctx, addr)
}

// OnDealComplete adds the piece to the storage provider
func (s *StorageProviderNodeConnector) OnDealComplete(ctx context.Context, deal storagemarket.MinerDeal, pieceSize uint64, pieceReader io.Reader) error {
	// TODO: storage provider is expecting a sector ID here. This won't work. The sector ID needs to be removed from
	// TODO: the return value, and storage provider needs to call OnDealSectorCommitted which should add Sector ID to its
	// TODO: callback.
	return s.pieceManager.SealPieceIntoNewSector(ctx, deal.DealID, pieceSize, pieceReader)
}

// LocatePieceForDealWithinSector finds the sector, offset and length of a piece associated with the given deal id
func (s *StorageProviderNodeConnector) LocatePieceForDealWithinSector(ctx context.Context, dealID uint64) (sectorNumber uint64, offset uint64, length uint64, err error) {
	var smState spasm.State
	err = s.chainStore.GetActorStateAt(ctx, s.chainStore.Head(), vmaddr.StorageMarketAddress, &smState)
	if err != nil {
		return 0, 0, 0, err
	}

	stateStore := StoreFromCbor(ctx, s.chainStore)
	proposals := adt.AsArray(stateStore, smState.Proposals)

	var minerState spaminer.State
	err = s.chainStore.GetActorStateAt(ctx, s.chainStore.Head(), s.minerAddr, &minerState)
	if err != nil {
		return 0, 0, 0, err
	}

	precommitted := adt.AsMap(stateStore, minerState.PreCommittedSectors)
	var sectorInfo spaminer.SectorPreCommitOnChainInfo
	err = precommitted.ForEach(&sectorInfo, func(key string) error {
		k, err := adt.ParseIntKey(key)
		if err != nil {
			return err
		}
		sectorNumber = uint64(k)

		for _, deal := range sectorInfo.Info.DealIDs {
			if uint64(deal) == dealID {
				offset = uint64(0)
				for _, did := range sectorInfo.Info.DealIDs {
					var proposal spasm.DealProposal
					found, err := proposals.Get(uint64(did), &proposal)
					if err != nil {
						return err
					}
					if !found {
						return errors.Errorf("Could not find miner deal %d in storage market state", did)
					}

					if uint64(did) == dealID {
						sectorNumber = uint64(k)
						length = uint64(proposal.PieceSize)
						return nil // Found!
					}
					offset += uint64(proposal.PieceSize)
				}
			}
		}
		return errors.New("Deal not found")
	})
	return
}
