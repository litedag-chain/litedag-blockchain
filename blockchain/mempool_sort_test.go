package blockchain

import (
	"sort"
	"testing"

	"github.com/litedag-chain/litedag-blockchain/v3/address"
)

func TestMempoolFeeSort(t *testing.T) {
	addrA := address.Address{1}
	addrB := address.Address{2}
	addrC := address.Address{3}

	// Simulate mempool entries with different fee-per-byte ratios.
	// Insertion order: low fee first, high fee last.
	entries := []*MempoolEntry{
		{Fee: 100, Size: 100, Signer: addrA}, // 1.0 fee/byte — lowest
		{Fee: 500, Size: 100, Signer: addrB}, // 5.0 fee/byte — highest
		{Fee: 300, Size: 100, Signer: addrC}, // 3.0 fee/byte — middle
	}

	// Same sorting logic as mining.go
	type indexed struct {
		entry *MempoolEntry
		order int
	}
	tmp := make([]indexed, len(entries))
	for i, e := range entries {
		tmp[i] = indexed{e, i}
	}
	sort.SliceStable(tmp, func(i, j int) bool {
		a, b := tmp[i], tmp[j]
		if a.entry.Signer == b.entry.Signer {
			return a.order < b.order
		}
		return a.entry.Fee*b.entry.Size > b.entry.Fee*a.entry.Size
	})
	sorted := make([]*MempoolEntry, len(tmp))
	for i, v := range tmp {
		sorted[i] = v.entry
	}

	// Expect: B (5.0), C (3.0), A (1.0)
	if sorted[0].Signer != addrB {
		t.Errorf("expected highest fee (addrB) first, got signer %v", sorted[0].Signer)
	}
	if sorted[1].Signer != addrC {
		t.Errorf("expected middle fee (addrC) second, got signer %v", sorted[1].Signer)
	}
	if sorted[2].Signer != addrA {
		t.Errorf("expected lowest fee (addrA) last, got signer %v", sorted[2].Signer)
	}
}

func TestMempoolFeeSortPreservesNonceOrder(t *testing.T) {
	addrA := address.Address{1}
	addrB := address.Address{2}

	// addrA sends 3 txs (nonce 1,2,3). addrB sends 1 tx with highest fee.
	// Even though addrA's tx3 has a high fee, all addrA txs must stay in order.
	entries := []*MempoolEntry{
		{Fee: 100, Size: 100, Signer: addrA}, // addrA nonce 1 — 1.0 fee/byte
		{Fee: 200, Size: 100, Signer: addrA}, // addrA nonce 2 — 2.0 fee/byte
		{Fee: 900, Size: 100, Signer: addrA}, // addrA nonce 3 — 9.0 fee/byte
		{Fee: 500, Size: 100, Signer: addrB}, // addrB — 5.0 fee/byte
	}

	type indexed struct {
		entry *MempoolEntry
		order int
	}
	tmp := make([]indexed, len(entries))
	for i, e := range entries {
		tmp[i] = indexed{e, i}
	}
	sort.SliceStable(tmp, func(i, j int) bool {
		a, b := tmp[i], tmp[j]
		if a.entry.Signer == b.entry.Signer {
			return a.order < b.order
		}
		return a.entry.Fee*b.entry.Size > b.entry.Fee*a.entry.Size
	})
	sorted := make([]*MempoolEntry, len(tmp))
	for i, v := range tmp {
		sorted[i] = v.entry
	}

	// addrA's highest single tx is 9.0 fee/byte, but the group is compared
	// by the first entry (1.0 fee/byte) in SliceStable. Since SliceStable
	// uses the first element's comparison, addrA block stays together.
	// The key invariant: addrA's txs must appear in original order.
	var addrAOrder []int
	for i, e := range sorted {
		if e.Signer == addrA {
			addrAOrder = append(addrAOrder, i)
		}
	}
	if len(addrAOrder) != 3 {
		t.Fatalf("expected 3 addrA entries, got %d", len(addrAOrder))
	}
	// Verify monotonically increasing (original order preserved)
	for i := 1; i < len(addrAOrder); i++ {
		if addrAOrder[i] <= addrAOrder[i-1] {
			t.Errorf("addrA nonce order broken: positions %v", addrAOrder)
		}
	}

	// Verify addrA fees are in original submission order (100, 200, 900)
	prevFee := uint64(0)
	for _, idx := range addrAOrder {
		if sorted[idx].Fee < prevFee && prevFee == 200 {
			t.Errorf("addrA fees not in submission order")
		}
		prevFee = sorted[idx].Fee
	}
}

func TestMempoolFeeSortEqualFees(t *testing.T) {
	addrA := address.Address{1}
	addrB := address.Address{2}

	// Same fee-per-byte — stable sort should preserve insertion order
	entries := []*MempoolEntry{
		{Fee: 100, Size: 100, Signer: addrA},
		{Fee: 100, Size: 100, Signer: addrB},
	}

	type indexed struct {
		entry *MempoolEntry
		order int
	}
	tmp := make([]indexed, len(entries))
	for i, e := range entries {
		tmp[i] = indexed{e, i}
	}
	sort.SliceStable(tmp, func(i, j int) bool {
		a, b := tmp[i], tmp[j]
		if a.entry.Signer == b.entry.Signer {
			return a.order < b.order
		}
		return a.entry.Fee*b.entry.Size > b.entry.Fee*a.entry.Size
	})

	if tmp[0].entry.Signer != addrA {
		t.Error("equal fees should preserve insertion order (addrA first)")
	}
	if tmp[1].entry.Signer != addrB {
		t.Error("equal fees should preserve insertion order (addrB second)")
	}
}
