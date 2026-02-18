# LiteDAG Blockchain
Copyright (c) 2026 The LiteDAG Project. All rights reserved.

LiteDAG is the world's first MiniDAG, a novel system that simulates a Directed Acyclic Graph (DAG) on a linear blockchain.
LiteDAG is secured by multi-chain Proof-of-Work via Merge-Mining.

## Running from source
To run the LiteDAG Blockchain node or CLI wallet you need the latest version of [Go](https://go.dev/) installed.

Build the node:
```sh
go build ./cmd/litedag-node/
```

Build the wallet:
```sh
go build ./cmd/litedag-wallet-cli/
```

You will find binaries inside the `cmd/litedag-node` and `cmd/litedag-wallet-cli` directories.

## Known Issues

- **TODO:** SIGSEGV when LMDB mapsize increases during sync (512MB → 1024MB).
