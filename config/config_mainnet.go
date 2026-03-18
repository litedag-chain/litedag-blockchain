//go:build !testnet && !unittest

package config

const P2P_BIND_PORT = 6310
const RPC_BIND_PORT = 6311
const STRATUM_BIND_PORT = 6312
const NETWORK_ID uint64 = 0x4f2102a9dc2b9d81 // Network identifier. It MUST be unique for each chain

const NETWORK_NAME = "mainnet"

const MIN_DIFFICULTY = 100_000
const DIFFICULTY_N = 120 // DAA half-life (30 minutes).

// GENESIS BLOCK INFO
const GENESIS_ADDRESS = "v15oxps781teqfug0f2ig4031y9zotogrokjjy0"
const GENESIS_TIMESTAMP = 1755522000 * 1000                          // FIXME: set to launch time
const BLOCK_REWARD_FEE_PERCENT = 10
const TEAM_STAKE_PUBKEY = "198ee3e69f5db0889f56bc5777dc101612b243eb42444b884bdabca801c024d7" // pubkey for v9206blqmfld0p1z73rv43lt6rvf33r22j72ts — reserves delegate ID 1

var SEED_NODES = []string{"node.litedag.com"}

// Addresses blocked from ALL transactions (transfers, staking, delegate ops).
// Only outbound (signer) is checked — receiving LDG still works, making these burn addresses.
// Old Virel treasury: key not accessible, funds permanently locked.
var BLOCKED_ADDRESSES = []string{
	"v139diixrpv0ftmip4mgpuy92u51iq4pnmgjsfn", // old Virel treasury
	"vcywy29s703885w3fc2uvvemv8gybj2yrrlfpz",  // founder wallet (281k from treasury)
	"v14h12wnbwyct2bycxfru8p7fw8z3vl98em9klh",  // founder wallet (350k from treasury)
	"v14z8srqrig7ffqfzy61yomuiywjluw3ooh6uj1",  // founder wallet (791k from treasury)
	"vraabvgm9z13caa9rm5ivyoofxvsy1fd9iqm6q",   // founder wallet (1.3M from treasury)
}

// PROOF OF STAKE
const MIN_STAKE_AMOUNT = 100 * COIN
const REGISTER_DELEGATE_BURN = 1_000 * COIN
const STAKE_UNLOCK_TIME = 60 * 60 * 24 * 30 * 2 / TARGET_BLOCK_TIME // staked funds unlock after 2 months

// HARD-FORKS

// Hardfork V2: transaction version field.
// Tx version: 1, block version: 0
const HARDFORK_V2_HEIGHT = 1

// Hardfork V3: Hybrid PoW/PoS.
// Active from genesis (height 1) on LiteDAG chain.
// Tx version: 1-5, block version: 1
const HARDFORK_V3_HEIGHT = 1
