package cborutil

import (
	"context"
	"time"

	"github.com/filecoin-project/go-filecoin/internal/pkg/encoding"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// IpldStore is a go-filecoin implementation of the go-hamt-ipld CborStore
// interface.
type IpldStore struct {
	blocks Blocks
}

// Blocks is the interface of block storage needed by the IpldStore
type Blocks interface {
	GetBlock(context.Context, cid.Cid) (blocks.Block, error)
	AddBlock(blocks.Block) error
}

// Blockstore is the interface of internal block storage used to implement
// a default Blocks interface.
type Blockstore interface {
	Get(cid.Cid) (blocks.Block, error)
	Put(blocks.Block) error
}

type bswrapper struct {
	bs Blockstore
}

func (bs *bswrapper) GetBlock(_ context.Context, c cid.Cid) (blocks.Block, error) {
	return bs.bs.Get(c)
}

func (bs *bswrapper) AddBlock(blk blocks.Block) error {
	return bs.bs.Put(blk)
}

// NewIpldStore returns an ipldstore backed by a blockstore.
func NewIpldStore(bs Blockstore) *IpldStore {
	return &IpldStore{blocks: &bswrapper{bs}}
}

// Get decodes the cbor bytes in the ipld node pointed to by cid c into out.
func (s *IpldStore) Get(ctx context.Context, c cid.Cid, out interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	blk, err := s.blocks.GetBlock(ctx, c)
	if err != nil {
		return err
	}
	return encoding.Decode(blk.RawData(), out)
}

// Put encodes the interface into cbor bytes and stores them as a block
func (s *IpldStore) Put(ctx context.Context, v interface{}) (cid.Cid, error) {
	mhType := uint64(mh.BLAKE2B_MIN + 31)
	mhLen := -1
	data, err := encoding.Encode(v)
	if err != nil {
		return cid.Undef, err
	}

	hash, err := mh.Sum(data, mhType, mhLen)
	if err != nil {
		return cid.Undef, err
	}
	c := cid.NewCidV1(cid.DagCBOR, hash)

	blk, err := blocks.NewBlockWithCid(data, c)
	if err != nil {
		return cid.Undef, err
	}

	if err := s.blocks.AddBlock(blk); err != nil {
		return cid.Undef, err
	}

	return c, nil
}
