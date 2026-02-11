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
const GENESIS_ADDRESS = "v139diixrpv0ftmip4mgpuy92u51iq4pnmgjsfn"
const GENESIS_TIMESTAMP = 1755522000 * 1000
const BLOCK_REWARD_FEE_PERCENT = 10
const TEAM_STAKE_PUBKEY = "3959a30cb83649dd38389dd6717cbadab6ceb92cd9e4c4352abfcf168bbf592e"

var SEED_NODES = []string{"127.0.0.1"}

// Addresses blocked from sending transactions (old Virel treasury, key not accessible)
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
// Hardfork date set for 2025-09-27 13:25:00 GMT.
// Tx version: 1-5, block version: 1
const HARDFORK_V3_HEIGHT = 1
