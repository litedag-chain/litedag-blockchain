//go:build !testnet && !unittest

package config

const P2P_BIND_PORT = 6310
const RPC_BIND_PORT = 6311
const STRATUM_BIND_PORT = 6312
const NETWORK_ID uint64 = 0x4f2102a9dc2b9d81 // Network identifier. It MUST be unique for each chain

const NETWORK_NAME = "mainnet"

const MIN_DIFFICULTY = 1_000
const DIFFICULTY_N = 120 // DAA half-life (30 minutes).

// GENESIS BLOCK INFO — testnet wallet (testnet.keys on VPS)
const GENESIS_ADDRESS = "v4pjzqm0wq8xbkqquf8psicpe5a387asbb1orv"
const GENESIS_TIMESTAMP = 1772110000 * 1000 // 2026-02-25
const BLOCK_REWARD_FEE_PERCENT = 10
const TEAM_STAKE_PUBKEY = "017478a6c73796f28bd31c386102638c0154c09d806632d9d770defb12a7a476"

var SEED_NODES = []string{"node.litedag.com"}

// Addresses blocked from ALL transactions (transfers, staking, delegate ops).
// Only outbound (signer) is checked — receiving LDG still works, making these burn addresses.
// Old Virel treasury: key not accessible, funds permanently locked.
var BLOCKED_ADDRESSES = []string{"v139diixrpv0ftmip4mgpuy92u51iq4pnmgjsfn"}

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
