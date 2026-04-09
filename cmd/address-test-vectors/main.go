package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/litedag-chain/litedag-blockchain/v3/address"
)

type Vector struct {
	Str     string `json:"str"`
	AddrHex string `json:"addr_hex"`
	Pid     uint64 `json:"pid"`
}

func main() {
	testAddresses := []string{
		"v15oxps781teqfug0f2ig4031y9zotogrokjjy0",
		"vhcunjkejkshqf4jjx9gfnu52162mvoa9eoqhi",
		"v9206blqmfld0p1z73rv43lt6rvf33r22j72ts",
		"vnjck6jmrkdyuac15jukzlm4phaz5lus9owma3",
		"v139diixrpv0ftmip4mgpuy92u51iq4pnmgjsfn",
	}

	paymentIds := []uint64{0, 1, 255, 256, 65535, 123456789, 1<<32 - 1, 1<<53 - 1}

	var vectors []Vector

	for _, addrStr := range testAddresses {
		parsed, err := address.FromString(addrStr)
		if err != nil {
			panic(err)
		}

		for _, pid := range paymentIds {
			integrated := address.Integrated{Addr: parsed.Addr, PaymentId: pid}
			intStr := integrated.String()

			rt, err := address.FromString(intStr)
			if err != nil {
				panic(fmt.Sprintf("round-trip failed for pid=%d: %v", pid, err))
			}
			if rt.PaymentId != pid || rt.Addr != parsed.Addr {
				panic("round-trip mismatch")
			}

			vectors = append(vectors, Vector{
				Str:     intStr,
				AddrHex: fmt.Sprintf("%x", rt.Addr[:]),
				Pid:     pid,
			})
		}
	}

	data, err := json.MarshalIndent(vectors, "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("address-test-vectors.json", data, 0644)
	fmt.Printf("Wrote %d vectors to address-test-vectors.json\n", len(vectors))
}
