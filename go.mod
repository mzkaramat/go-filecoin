module github.com/filecoin-project/go-filecoin

go 1.13

require (
	contrib.go.opencensus.io/exporter/jaeger v0.1.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/AndreasBriese/bbloom v0.0.0-20190825152654-46b345b51c96 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/Workiva/go-datastructures v1.0.50
	github.com/cskr/pubsub v1.0.2
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.7.3-0.20190315170154-87d593639c77
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.3.3 // indirect
	github.com/filecoin-project/filecoin-ffi v0.0.0-20191221090835-c7bbef445934
	github.com/filecoin-project/go-address v0.0.1
	github.com/filecoin-project/go-amt-ipld/v2 v2.0.1-0.20200131012142-05d80eeccc5e
	github.com/filecoin-project/go-cbor-util v0.0.0-20191219014500-08c40a1e63a2
	github.com/filecoin-project/go-data-transfer v0.0.0-20191219005021-4accf56bd2ce
	github.com/filecoin-project/go-fil-markets v0.0.0-20200204152108-87675b3f9b04
	github.com/filecoin-project/go-leb128 v0.0.0-20190212224330-8d79a5489543
	github.com/filecoin-project/go-paramfetch v0.0.1
	github.com/filecoin-project/go-sectorbuilder v0.0.2-0.20200114015900-4103afa82689
	github.com/filecoin-project/go-statestore v0.1.0
	github.com/filecoin-project/go-storage-miner v0.0.0-20200122233640-6a01988b8217
	github.com/filecoin-project/specs-actors v0.0.0-20200210130641-2d1fbd8672cf
	github.com/fxamacker/cbor v1.5.0
	github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golangci/golangci-lint v1.21.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gorilla/mux v1.7.0 // indirect
	github.com/ipfs/go-bitswap v0.1.8
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-blockservice v0.1.3-0.20190908200855-f22eea50656c
	github.com/ipfs/go-car v0.0.3-0.20200124090545-1a340009d896
	github.com/ipfs/go-cid v0.0.5
	github.com/ipfs/go-datastore v0.1.1
	github.com/ipfs/go-ds-badger v0.0.7
	github.com/ipfs/go-fs-lock v0.0.1
	github.com/ipfs/go-graphsync v0.0.4
	github.com/ipfs/go-hamt-ipld v0.0.15-0.20200131012125-dd88a59d3f2e
	github.com/ipfs/go-ipfs-blockstore v0.1.1
	github.com/ipfs/go-ipfs-chunker v0.0.1
	github.com/ipfs/go-ipfs-cmdkit v0.0.1
	github.com/ipfs/go-ipfs-cmds v0.0.1
	github.com/ipfs/go-ipfs-exchange-interface v0.0.1
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipfs-files v0.0.4
	github.com/ipfs/go-ipfs-keystore v0.0.1
	github.com/ipfs/go-ipfs-routing v0.1.0
	github.com/ipfs/go-ipld-cbor v0.0.5-0.20200204214505-252690b78669
	github.com/ipfs/go-ipld-format v0.0.2
	github.com/ipfs/go-log v1.0.1
	github.com/ipfs/go-log/v2 v2.0.2 // indirect
	github.com/ipfs/go-merkledag v0.2.4
	github.com/ipfs/go-path v0.0.1
	github.com/ipfs/go-unixfs v0.2.2
	github.com/ipfs/iptb v1.3.8-0.20190401234037-98ccf4228a73
	github.com/ipld/go-ipld-prime v0.0.2-0.20191108012745-28a82f04c785
	github.com/ipsn/go-secp256k1 v0.0.0-20180726113642-9d62b9f0bc52
	github.com/jbenet/goprocess v0.1.3
	github.com/jstemmer/go-junit-report v0.9.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/libp2p/go-libp2p v0.4.1-0.20191006140250-5f60501a04d5
	github.com/libp2p/go-libp2p-autonat-svc v0.1.0
	github.com/libp2p/go-libp2p-circuit v0.1.3
	github.com/libp2p/go-libp2p-core v0.3.0
	github.com/libp2p/go-libp2p-kad-dht v0.1.1
	github.com/libp2p/go-libp2p-peerstore v0.1.4
	github.com/libp2p/go-libp2p-pubsub v0.2.5
	github.com/libp2p/go-libp2p-secio v0.2.1 // indirect
	github.com/libp2p/go-libp2p-swarm v0.2.2
	github.com/libp2p/go-libp2p-testing v0.1.1 // indirect
	github.com/libp2p/go-stream-muxer v0.0.1
	github.com/libp2p/go-yamux v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/multiformats/go-multiaddr v0.2.0
	github.com/multiformats/go-multiaddr-dns v0.2.0
	github.com/multiformats/go-multiaddr-net v0.1.1
	github.com/multiformats/go-multibase v0.0.1
	github.com/multiformats/go-multihash v0.0.13
	github.com/onsi/ginkgo v1.10.3 // indirect
	github.com/onsi/gomega v1.7.1 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/otiai10/copy v1.0.2
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/polydawn/refmt v0.0.0-20190809202753-05966cbd336a
	github.com/prometheus/client_golang v1.2.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.5.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20200206220010-03c9665e2a66
	github.com/whyrusleeping/go-logging v0.0.1
	github.com/whyrusleeping/go-sysinfo v0.0.0-20190219211824-4a357d4b90b1
	go.opencensus.io v0.22.2
	go.uber.org/atomic v1.5.1 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20200207205829-a95e85b341fd
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/net v0.0.0-20191101175033-0deb6923b6d9 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20191216173652-a0e659d51361
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/api v0.13.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6 // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/urfave/cli.v2 v2.0.0-20180128182452-d3ae77c26ac8
	gopkg.in/yaml.v2 v2.2.5 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/filecoin-project/filecoin-ffi => ./vendors/filecoin-ffi
