// snapshot-export reads state from a Virel (or LiteDAG) chain's LMDB database
// and generates genesis/prefunded.go with all address balances at a given height.
//
// One-time tool for fork migration. Remove after mainnet is stable.
// See GENESIS_MIGRATION.md for context.
//
// Usage:
//
//	go run ./cmd/snapshot-export -datadir /path/to/chain-data -height 50000
//
// The -datadir should point to the node's data directory (contains lmdb/).
// The -height flag is required — it specifies the snapshot height and is
// recorded in the generated output for auditability.
//
// The tool reads liquid balances from the state index AND staked funds from
// the delegate index, so migrated users get their full holdings (liquid + staked)
// as liquid LDG on the new chain.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/litedag-chain/litedag-blockchain/v3/adb"
	"github.com/litedag-chain/litedag-blockchain/v3/address"
	"github.com/litedag-chain/litedag-blockchain/v3/chaintype"
	"github.com/litedag-chain/litedag-blockchain/v3/logger"

	lmdbdriver "github.com/litedag-chain/litedag-blockchain/v3/adb/lmdb"
)

type entry struct {
	AddrStr string
	Balance uint64
}

func main() {
	datadir := flag.String("datadir", "", "path to chain data directory (contains lmdb/)")
	outfile := flag.String("out", "genesis/prefunded.go", "output Go source file")
	height := flag.Int64("height", -1, "required: snapshot height (must match chain top height)")
	flag.Parse()

	if *datadir == "" || *height < 0 {
		fmt.Fprintln(os.Stderr, "usage: snapshot-export -datadir <path> -height <N> [-out genesis/prefunded.go]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  -height is required. The chain DB must be synced to exactly this height.")
		fmt.Fprintln(os.Stderr, "  Run without -height to see the current chain height.")
		if *datadir != "" && *height < 0 {
			printChainHeight(*datadir)
		}
		os.Exit(1)
	}

	lmdbPath := *datadir + "/lmdb/"
	if _, err := os.Stat(lmdbPath + "data.mdb"); err != nil {
		fmt.Fprintf(os.Stderr, "LMDB not found at %s: %v\n", lmdbPath, err)
		os.Exit(1)
	}

	log := logger.New()
	db, err := lmdbdriver.New(lmdbPath, 0755, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open LMDB: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	topoIdx := db.Index("topo")
	stateIdx := db.Index("state")
	delegateIdx := db.Index("delegate")

	// Balances accumulator: liquid balance + staked funds per address
	balances := make(map[address.Address]uint64)
	var redirectedBlocked int
	var redirectedAmount uint64
	var skippedBurn int
	var skippedZero int
	var stakedTotal uint64

	err = db.View(func(txn adb.Txn) error {
		// 1. Verify chain height matches requested snapshot height
		topHeight, err := getTopHeight(txn, topoIdx)
		if err != nil {
			return fmt.Errorf("failed to read chain height: %w", err)
		}
		if topHeight != uint64(*height) {
			return fmt.Errorf(
				"height mismatch: requested %d but chain is at %d\n"+
					"  The chain DB must be synced to exactly the requested height.\n"+
					"  Either use -height %d or sync the old node to height %d and stop it.",
				*height, topHeight, topHeight, *height)
		}
		fmt.Fprintf(os.Stderr, "chain height verified: %d\n", topHeight)

		// 2. Read liquid balances from state index
		err = txn.ForEach(stateIdx, func(k, v []byte) error {
			if len(k) != address.SIZE {
				return nil
			}

			addr := address.Address(k)

			if addr == address.INVALID_ADDRESS {
				skippedBurn++
				return nil
			}
			if address.IsBlocked(addr) {
				// Redirect blocked treasury balance to genesis address (new treasury)
				state := &chaintype.State{}
				if err := state.Deserialize(v); err == nil && state.Balance > 0 {
					balances[address.GenesisAddress] += state.Balance
					redirectedBlocked++
					redirectedAmount += state.Balance
					fmt.Fprintf(os.Stderr, "redirected blocked %s → genesis: %d\n", addr, state.Balance)
				}
				return nil
			}
			if addr.IsDelegate() {
				return nil
			}

			state := &chaintype.State{}
			if err := state.Deserialize(v); err != nil {
				fmt.Fprintf(os.Stderr, "warning: bad state for %s: %v\n", addr, err)
				return nil
			}

			if state.Balance > 0 {
				balances[addr] += state.Balance
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("reading state: %w", err)
		}

		// 3. Read staked funds from delegate index and credit to owners
		err = txn.ForEach(delegateIdx, func(k, v []byte) error {
			d := &chaintype.Delegate{}
			if err := d.Deserialize(v); err != nil {
				fmt.Fprintf(os.Stderr, "warning: bad delegate %x: %v\n", k, err)
				return nil
			}

			for _, fund := range d.Funds {
				if fund.Amount == 0 {
					continue
				}
				if address.IsBlocked(fund.Owner) {
					// Redirect staked funds from blocked address to genesis (new treasury)
					balances[address.GenesisAddress] += fund.Amount
					redirectedAmount += fund.Amount
					stakedTotal += fund.Amount
					fmt.Fprintf(os.Stderr, "redirected staked from blocked %s → genesis: %d\n", fund.Owner, fund.Amount)
					continue
				}
				balances[fund.Owner] += fund.Amount
				stakedTotal += fund.Amount
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("reading delegates: %w", err)
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Build sorted entries, compute total supply
	var entries []entry
	var totalSupply uint64
	for addr, balance := range balances {
		if balance == 0 {
			skippedZero++
			continue
		}
		totalSupply += balance
		entries = append(entries, entry{
			AddrStr: addr.String(),
			Balance: balance,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].AddrStr < entries[j].AddrStr
	})

	// Generate output
	var buf strings.Builder
	if err := outputTemplate.Execute(&buf, templateData{
		Height:      uint64(*height),
		Entries:     entries,
		TotalSupply: totalSupply,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "template error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outfile, []byte(buf.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", *outfile, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\nsnapshot export complete:\n")
	fmt.Fprintf(os.Stderr, "  height:          %d\n", *height)
	fmt.Fprintf(os.Stderr, "  addresses:       %d\n", len(entries))
	fmt.Fprintf(os.Stderr, "  liquid supply:    %.2f LDG\n", float64(totalSupply-stakedTotal)/1e9)
	fmt.Fprintf(os.Stderr, "  staked supply:   %.2f LDG\n", float64(stakedTotal)/1e9)
	fmt.Fprintf(os.Stderr, "  total supply:    %.2f LDG (%d atomic)\n", float64(totalSupply)/1e9, totalSupply)
	fmt.Fprintf(os.Stderr, "  blocked → treasury: %d addresses (%.2f LDG)\n", redirectedBlocked, float64(redirectedAmount)/1e9)
	fmt.Fprintf(os.Stderr, "  excluded burn:    %d\n", skippedBurn)
	fmt.Fprintf(os.Stderr, "  excluded zero:    %d\n", skippedZero)
	fmt.Fprintf(os.Stderr, "  output:          %s\n", *outfile)
}

// getTopHeight reads the topo index to find the highest block height.
func getTopHeight(txn adb.Txn, topoIdx adb.Index) (uint64, error) {
	var maxHeight uint64
	var found bool
	err := txn.ForEach(topoIdx, func(k, v []byte) error {
		if len(k) != 8 {
			return nil
		}
		h := binary.LittleEndian.Uint64(k)
		if !found || h > maxHeight {
			maxHeight = h
			found = true
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("topo index is empty")
	}
	return maxHeight, nil
}

// printChainHeight opens the DB just to report the current height.
func printChainHeight(datadir string) {
	lmdbPath := datadir + "/lmdb/"
	if _, err := os.Stat(lmdbPath + "data.mdb"); err != nil {
		return
	}
	log := logger.New()
	db, err := lmdbdriver.New(lmdbPath, 0755, log)
	if err != nil {
		return
	}
	defer db.Close()
	topoIdx := db.Index("topo")
	db.View(func(txn adb.Txn) error {
		h, err := getTopHeight(txn, topoIdx)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "\n  chain is at height %d — use: -height %d\n", h, h)
		return nil
	})
}

type templateData struct {
	Height      uint64
	Entries     []entry
	TotalSupply uint64
}

var outputTemplate = template.Must(template.New("prefunded").Parse(`package genesis

import "github.com/litedag-chain/litedag-blockchain/v3/address"

// Generated by cmd/snapshot-export. DO NOT EDIT.
// Source: old Virel chain snapshot at height {{.Height}}
// Includes liquid balances + staked funds (credited as liquid LDG).
// See GENESIS_MIGRATION.md for architecture, rationale, and cleanup plan.

const PrefundedSupply uint64 = {{.TotalSupply}}

var PrefundedBalances = map[address.Address]uint64{
{{- range .Entries}}
	mustAddr("{{.AddrStr}}"): {{.Balance}},
{{- end}}
}

func mustAddr(s string) address.Address {
	a, err := address.FromString(s)
	if err != nil {
		panic("invalid prefunded address " + s + ": " + err.Error())
	}
	return a.Addr
}
`))
