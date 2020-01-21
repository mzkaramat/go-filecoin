package block

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/filecoin-project/go-filecoin/internal/pkg/encoding"
	"github.com/filecoin-project/go-filecoin/internal/pkg/types"
	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	node "github.com/ipfs/go-ipld-format"
	"github.com/polydawn/refmt/obj/atlas"

	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/address"
)

func init() {
	before := Block{Height: types.Uint64(666)}
	fmt.Printf("before: %v\n", before)
	is, err := StructToTuple(reflect.ValueOf(before))
	if err != nil {
		panic(err)
	}
	after, err := TupleToStruct(is, reflect.TypeOf(Block{}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("after: %v\n", after)
}

type BlockTuple []interface{}

// StructToTuple converts any struct value to an []interface{}{} value where
// the ith slice element is the ith public field.
func StructToTuple(val reflect.Value) (reflect.Value, error) {
	if val.Kind() != reflect.Struct {
		return reflect.ValueOf(nil), fmt.Errorf("struct to tuple expects struct")
	}
	n := val.NumField()
	tuple := make([]interface{}, 0)
	for i := 0; i < n; i++ {
		f := val.Field(i)
		if f.CanInterface() { // only exported fields
			tuple = append(tuple, f.Interface())
		}
	}
	return reflect.ValueOf(tuple), nil
}

// TupleToStruct Converts an []interface{}{} value to a struct value where the
// ith public field is set to the ith slice element.
func TupleToStruct(tupleVal reflect.Value, structType reflect.Type) (reflect.Value, error) {
	fmt.Printf("the tuple: %v\n", tupleVal)
	structPtrVal := reflect.New(structType)
	structVal := structPtrVal.Elem()
	if structVal.Kind() != reflect.Struct {
		return reflect.ValueOf(nil), fmt.Errorf("TupleToStruct expects struct type")
	}
	if tupleVal.Kind() != reflect.Slice {
		fmt.Printf("TupleToStruct expects slice kind not: %v\n", tupleVal.Kind())
		return reflect.ValueOf(nil), fmt.Errorf("TupleToStruct expects []interface{} value")
	}
	n := structVal.NumField()
	j := 0 // total tuple values consumed
	for i := 0; i < n; i++ {
		f := structVal.Field(i)
		if f.CanInterface() { // only exported fields
			tupleI := tupleVal.Index(j)
			f.Set(tupleI.Elem())
			j++
		}
	}
	return structVal, nil
}

var blockAtlasEntry = atlas.BuildEntry(Block{}).Transform().
	TransformMarshal(func(liveForm reflect.Value) (serialForm reflect.Value, err error) {
		return StructToTuple(liveForm)
	}, reflect.TypeOf(BlockTuple{})).
	TransformUnmarshal(func(serialForm reflect.Value) (liveForm reflect.Value, err error) {
		return TupleToStruct(serialForm, reflect.TypeOf(Block{}))
	}, reflect.TypeOf(BlockTuple{})).
	Complete()

// Block is a block in the blockchain.
type Block struct {
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
	ParentWeight types.Uint64 `json:"parentWeight"`

	// Height is the chain height of this block.
	Height types.Uint64 `json:"height"`

	// StateRoot is a cid pointer to the state tree after application of the
	// transactions state transitions.
	StateRoot cid.Cid `json:"stateRoot,omitempty" refmt:",omitempty"`

	// MessageReceipts is a set of receipts matching to the sending of the `Messages`.
	MessageReceipts cid.Cid `json:"messageReceipts,omitempty" refmt:",omitempty"`

	// Messages is the set of messages included in this block
	Messages types.TxMeta `json:"messages,omitempty" refmt:",omitempty"`

	// The aggregate signature of all BLS signed messages in the block
	BLSAggregateSig types.Signature `json:"blsAggregateSig"`

	// The timestamp, in seconds since the Unix epoch, at which this block was created.
	Timestamp types.Uint64 `json:"timestamp"`

	// The signature of the miner's worker key over the block
	BlockSig types.Signature `json:"blocksig"`

	// ForkSignaling is extra data used by miners to communicate
	ForkSignaling types.Uint64

	cachedCid cid.Cid

	cachedBytes []byte
}

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
	obj, err := cbor.WrapObject(b, types.DefaultHashFunction, -1)
	if err != nil {
		panic(err)
	}

	return obj
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
