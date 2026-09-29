package main

// devnet-prove: attempts a REAL Solana devnet anchor using the production
// solana.DevnetClient. Generates a keypair (or loads SOLANA_SETTLEMENT_KEYPAIR),
// requests an airdrop, and anchors a canonical trade memo — printing the live
// signature + Explorer URL. Faucet 429s are reported, not fatal; the client's
// simulated fallback is what keeps the product running when funding is unavailable.

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/solana"
)

func rpc(method string, params []interface{}) (json.RawMessage, error) {
	body, _ := json.Marshal(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	resp, err := http.Post("https://api.devnet.solana.com", "application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var env struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &env)
	if env.Error != nil {
		return nil, fmt.Errorf("%d %s", env.Error.Code, env.Error.Message)
	}
	return env.Result, nil
}

func main() {
	// Load or generate a keypair; persist it so re-runs reuse the funded account.
	var sk ed25519.PrivateKey
	keyEnv := os.Getenv("SOLANA_SETTLEMENT_KEYPAIR")
	keyFile := "services/.devnet-keypair.hex"
	if keyEnv == "" {
		if b, err := os.ReadFile(keyFile); err == nil {
			if raw, err := hex.DecodeString(strings.TrimSpace(string(b))); err == nil && len(raw) == ed25519.PrivateKeySize {
				sk = ed25519.PrivateKey(raw)
			}
		}
	}
	if sk == nil {
		_, sk, _ = ed25519.GenerateKey(nil)
		_ = os.WriteFile(keyFile, []byte(hex.EncodeToString(sk)), 0o600)
	}
	os.Setenv("SOLANA_SETTLEMENT_KEYPAIR", hex.EncodeToString(sk))

	client := solana.NewDevnetClient("https://api.devnet.solana.com", hex.EncodeToString(sk))
	fmt.Printf("signer pubkey: %s\nmode: %s\n", client.Pubkey(), client.Mode())

	// Try airdrop a few times (public faucet is rate-limited).
	funded := false
	for i := 0; i < 3 && !funded; i++ {
		_, err := rpc("requestAirdrop", []interface{}{client.Pubkey(), 200000000})
		if err != nil {
			fmt.Printf("airdrop attempt %d: %v\n", i+1, err)
			time.Sleep(4 * time.Second)
			continue
		}
		fmt.Println("airdrop submitted; waiting for confirmation...")
		time.Sleep(8 * time.Second)
		funded = true
	}

	ctx := context.Background()
	memo := fmt.Sprintf("verinode:canary:demo-trade:%d", time.Now().Unix())
	sig, err := client.AnchorMemo(ctx, memo)
	if err != nil {
		fmt.Printf("anchor error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("signature: %s\n", sig)
	if url := solana.ExplorerURL(sig, "devnet"); url != "" {
		fmt.Printf("EXPLORER: %s\n", url)
	} else {
		fmt.Println("EXPLORER: (simulated proof — fund the signer for a live tx)")
	}
}
