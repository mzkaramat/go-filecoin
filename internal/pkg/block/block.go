package block

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	fbig "github.com/filecoin-project/specs-actors/actors/abi/big"
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	node "github.com/ipfs/go-ipld-format"
	mh "github.com/multiformats/go-multihash"

	e "github.com/filecoin-project/go-filecoin/internal/pkg/enccid"
	"github.com/filecoin-project/go-filecoin/internal/pkg/encoding"
	"github.com/filecoin-project/go-filecoin/internal/pkg/types"
)

// Block is a block in the blockchain.
type Block struct {
	// control field for encoding struct as an array
	_ struct{} `cbor:",toarray"`

	// Miner is the address of the miner actor that mined this block.
	Miner address.Address `json:"miner"`

	// Ticket is the ticket submitted with this block.
	Ticket Ticket `json:"ticket"`

	// EPoStInfo wraps all data for verifying this block's Election PoSt
	EPoStInfo EPoStInfo `json:"ePoStInfo"`

	// Parents is the set of parents this block was based on. Typically one,
	// but can be several in the case where there were multiple winning ticket-
	// holders for an epoch.
	Parents TipSetKey `json:"parents"`

	// ParentWeight is the aggregate chain weight of the parent set.
	ParentWeight fbig.Int `json:"parentWeight"`

	// Height is the chain height of this block.
	Height uint64 `json:"height"`

	// StateRoot is a cid pointer to the state tree after application of the
	// transactions state transitions.
	StateRoot e.Cid `json:"stateRoot,omitempty"`

	// MessageReceipts is a set of receipts matching to the sending of the `Messages`.
	MessageReceipts e.Cid `json:"messageReceipts,omitempty"`

	// Messages is the set of messages included in this block
	Messages e.Cid `json:"messages,omitempty"`

	// The aggregate signature of all BLS signed messages in the block
	BLSAggregateSig types.Signature `json:"blsAggregateSig"`

	// The timestamp, in seconds since the Unix epoch, at which this block was created.
	Timestamp uint64 `json:"timestamp"`

	// The signature of the miner's worker key over the block
	BlockSig types.Signature `json:"blocksig"`

	// ForkSignaling is extra data used by miners to communicate
	ForkSignaling uint64

	cachedCid cid.Cid

	cachedBytes []byte
}

// IndexMessagesField is the message field position in the encoded block
const IndexMessagesField = 8

// IndexParentsField is the parents field position in the encoded block
const IndexParentsField = 3

// Cid returns the content id of this block.
func (b *Block) Cid() cid.Cid {
	if b.cachedCid == cid.Undef {
		if b.cachedBytes == nil {
			bytes, err := encoding.Encode(b)
			if err != nil {
				panic(err)
			}
			b.cachedBytes = bytes
		}
		c, err := cid.Prefix{
			Version:  1,
			Codec:    cid.DagCBOR,
			MhType:   types.DefaultHashFunction,
			MhLength: -1,
		}.Sum(b.cachedBytes)
		if err != nil {
			panic(err)
		}

		b.cachedCid = c
	}

	return b.cachedCid
}

// ToNode converts the Block to an IPLD node.
func (b *Block) ToNode() node.Node {
	// Use 32 byte / 256 bit digest. TODO pull this out into a constant?
	// obj, err := cbor.WrapObject(b, types.DefaultHashFunction, -1)
	// if err != nil {
	// 	panic(err)
	// }
	mhType := uint64(mh.BLAKE2B_MIN + 31)
	mhLen := -1

	data, err := encoding.Encode(b)
	if err != nil {
		panic(err)
	}

	hash, err := mh.Sum(data, mhType, mhLen)
	if err != nil {
		panic(err)
	}
	c := cid.NewCidV1(cid.DagCBOR, hash)

	blk, err := blocks.NewBlockWithCid(data, c)
	if err != nil {
		panic(err)
	}
	node, err := cbor.DecodeBlock(blk)
	if err != nil {
		panic(err)
	}
	return node
}

func (b *Block) String() string {
	errStr := "(error encoding Block)"
	cid := b.Cid()
	js, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return errStr
	}
	return fmt.Sprintf("Block cid=[%v]: %s", cid, string(js))
}

// DecodeBlock decodes raw cbor bytes into a Block.
func DecodeBlock(b []byte) (*Block, error) {
	var out Block
	if err := encoding.Decode(b, &out); err != nil {
		return nil, err
	}

	out.cachedBytes = b

	return &out, nil
}

// Equals returns true if the Block is equal to other.
func (b *Block) Equals(other *Block) bool {
	return b.Cid().Equals(other.Cid())
}

// SignatureData returns the block's bytes without the blocksig for signature
// creating and verification
func (b *Block) SignatureData() []byte {
	tmp := &Block{
		Miner:           b.Miner,
		Ticket:          b.Ticket,  // deep copy needed??
		Parents:         b.Parents, // deep copy needed??
		ParentWeight:    b.ParentWeight,
		Height:          b.Height,
		Messages:        b.Messages,
		StateRoot:       b.StateRoot,
		MessageReceipts: b.MessageReceipts,
		EPoStInfo:       b.EPoStInfo,
		Timestamp:       b.Timestamp,
		BLSAggregateSig: b.BLSAggregateSig,
		ForkSignaling:   b.ForkSignaling,
		// BlockSig omitted
	}

	return tmp.ToNode().RawData()
}
