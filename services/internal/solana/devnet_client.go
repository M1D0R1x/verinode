package solana

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// devnet_client.go implements a minimal, dependency-free Solana JSON-RPC client
// capable of anchoring a canonical trade digest on-chain via the SPL Memo program.
//
// It is deliberately small: per Invariant 6 the chain is an OPTIONAL, read-only
// audit mirror, never authoritative. The client anchors a compact
// "verinode:<kind>:<trade_id>:<digest>" memo so that anyone can independently
// verify — from a public Solana Explorer link — that the off-chain canonical
// record was mirrored at a point in time.
//
// Signing keypair resolution (in priority order):
//   1. SOLANA_SETTLEMENT_KEYPAIR — base58 or hex-encoded 64-byte ed25519 secret.
//   2. If unset, the client operates in SIMULATED mode: it returns a
//      deterministic, clearly-labelled proof (sig prefixed "SIMULATED-") so the
//      whole product runs end-to-end without funded keys. Going live is a
//      one-env-var change, not a rearchitecture.

// MemoProgramID is the canonical SPL Memo program address on all Solana clusters.
const MemoProgramID = "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr"

// SystemProgramID is the all-zero program used as a read-only reference account.
const SystemProgramID = "11111111111111111111111111111111"

// DevnetClient submits real transactions to a Solana JSON-RPC endpoint.
type DevnetClient struct {
	endpoint string
	http     *http.Client
	signer   ed25519.PrivateKey // nil => simulated mode
}

// NewDevnetClient builds a client. When keypair is empty it reads
// SOLANA_SETTLEMENT_KEYPAIR from the environment; if that is also empty the
// client runs in simulated mode.
func NewDevnetClient(endpoint, keypair string) *DevnetClient {
	if endpoint == "" {
		endpoint = "https://api.devnet.solana.com"
	}
	if keypair == "" {
		keypair = os.Getenv("SOLANA_SETTLEMENT_KEYPAIR")
	}
	c := &DevnetClient{
		endpoint: endpoint,
		http:     &http.Client{Timeout: 20 * time.Second},
	}
	if keypair != "" {
		if sk, err := parseEd25519Secret(keypair); err == nil {
			c.signer = sk
		}
	}
	return c
}

// Live reports whether the client will submit real transactions.
func (c *DevnetClient) Live() bool { return c.signer != nil }

// Mode returns a human-readable mode string for surfacing to operators.
func (c *DevnetClient) Mode() string {
	if c.Live() {
		return "live-devnet"
	}
	return "simulated"
}

// Pubkey returns the base58 signer address, or the zero address in simulated mode.
func (c *DevnetClient) Pubkey() string {
	if c.signer == nil {
		return SystemProgramID
	}
	return base58Encode(c.signer.Public().(ed25519.PublicKey))
}

// AnchorMemo anchors an arbitrary memo string on-chain and returns the
// transaction signature (base58). In simulated mode it returns a deterministic
// "SIMULATED-<sha256-prefix>" pseudo-signature so callers always get a stable id.
func (c *DevnetClient) AnchorMemo(ctx context.Context, memo string) (string, error) {
	if c.signer == nil {
		sum := sha256.Sum256([]byte(memo))
		return "SIMULATED-" + hex.EncodeToString(sum[:16]), nil
	}

	blockhash, err := c.latestBlockhash(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch blockhash: %w", err)
	}

	txBytes, err := c.buildMemoTx(memo, blockhash)
	if err != nil {
		return "", fmt.Errorf("build tx: %w", err)
	}

	sig, err := c.sendTransaction(ctx, txBytes)
	if err != nil {
		return "", fmt.Errorf("send tx: %w", err)
	}
	return sig, nil
}

// ExplorerURL returns a public Solana Explorer URL for a signature. Simulated
// signatures return an empty string since they are not on-chain.
func ExplorerURL(sig, cluster string) string {
	if sig == "" || len(sig) >= 10 && sig[:10] == "SIMULATED-" {
		return ""
	}
	if cluster == "" {
		cluster = "devnet"
	}
	return "https://explorer.solana.com/tx/" + sig + "?cluster=" + cluster
}

// -----------------------------------------------------------------------------
// JSON-RPC plumbing
// -----------------------------------------------------------------------------

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *DevnetClient) call(ctx context.Context, method string, params []interface{}, out interface{}) error {
	body, _ := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *rpcError       `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return err
	}
	if envelope.Error != nil {
		return fmt.Errorf("rpc %s: %d %s", method, envelope.Error.Code, envelope.Error.Message)
	}
	if out != nil {
		return json.Unmarshal(envelope.Result, out)
	}
	return nil
}

func (c *DevnetClient) latestBlockhash(ctx context.Context) ([32]byte, error) {
	var res struct {
		Value struct {
			Blockhash string `json:"blockhash"`
		} `json:"value"`
	}
	if err := c.call(ctx, "getLatestBlockhash", []interface{}{map[string]string{"commitment": "finalized"}}, &res); err != nil {
		return [32]byte{}, err
	}
	raw, err := base58Decode(res.Value.Blockhash)
	if err != nil || len(raw) != 32 {
		return [32]byte{}, fmt.Errorf("invalid blockhash")
	}
	var bh [32]byte
	copy(bh[:], raw)
	return bh, nil
}

func (c *DevnetClient) sendTransaction(ctx context.Context, tx []byte) (string, error) {
	encoded := base64.StdEncoding.EncodeToString(tx)
	var sig string
	err := c.call(ctx, "sendTransaction", []interface{}{
		encoded,
		map[string]interface{}{"encoding": "base64", "skipPreflight": false, "preflightCommitment": "confirmed"},
	}, &sig)
	if err != nil {
		return "", err
	}
	return sig, nil
}

// -----------------------------------------------------------------------------
// Transaction assembly (legacy message format, single signer)
// -----------------------------------------------------------------------------

// buildMemoTx assembles a signed legacy transaction with one Memo instruction.
func (c *DevnetClient) buildMemoTx(memo string, blockhash [32]byte) ([]byte, error) {
	payer := c.signer.Public().(ed25519.PublicKey)
	memoProg, err := base58Decode(MemoProgramID)
	if err != nil {
		return nil, err
	}

	// Account list: [payer (signer, writable), memo program (invoked)].
	// Header: numRequiredSignatures=1, numReadonlySigned=0, numReadonlyUnsigned=1.
	var msg bytes.Buffer
	msg.WriteByte(1) // numRequiredSignatures
	msg.WriteByte(0) // numReadonlySignedAccounts
	msg.WriteByte(1) // numReadonlyUnsignedAccounts (the memo program)

	// Compact-u16 account count = 2
	writeCompactU16(&msg, 2)
	msg.Write(payer)   // index 0
	msg.Write(memoProg) // index 1

	// Recent blockhash
	msg.Write(blockhash[:])

	// Instructions: compact-u16 count = 1
	writeCompactU16(&msg, 1)
	msg.WriteByte(1)            // program id index -> memo program
	writeCompactU16(&msg, 0)    // 0 account indexes for a signer-only memo
	writeCompactU16(&msg, len(memo))
	msg.WriteString(memo)

	message := msg.Bytes()
	sig := ed25519.Sign(c.signer, message)

	// Wire tx = compact-u16 sig count (1) + 64-byte sig + message
	var tx bytes.Buffer
	writeCompactU16(&tx, 1)
	tx.Write(sig)
	tx.Write(message)
	return tx.Bytes(), nil
}

// writeCompactU16 encodes a length in Solana's shortvec (compact-u16) format.
func writeCompactU16(buf *bytes.Buffer, n int) {
	for {
		b := byte(n & 0x7f)
		n >>= 7
		if n == 0 {
			buf.WriteByte(b)
			return
		}
		buf.WriteByte(b | 0x80)
	}
}

// parseEd25519Secret accepts a hex (64 or 128 char) or base58-encoded 64-byte
// ed25519 secret key (Solana keypair format: 32-byte seed || 32-byte pubkey),
// or a JSON byte array as produced by `solana-keygen`.
func parseEd25519Secret(s string) (ed25519.PrivateKey, error) {
	// JSON byte-array form: [12,34,...]
	if len(s) > 0 && s[0] == '[' {
		var arr []byte
		if err := json.Unmarshal([]byte(s), &arr); err == nil && len(arr) == ed25519.PrivateKeySize {
			return ed25519.PrivateKey(arr), nil
		}
	}
	// Hex form
	if raw, err := hex.DecodeString(s); err == nil {
		if len(raw) == ed25519.PrivateKeySize {
			return ed25519.PrivateKey(raw), nil
		}
		if len(raw) == ed25519.SeedSize {
			return ed25519.NewKeyFromSeed(raw), nil
		}
	}
	// Base58 form
	if raw, err := base58Decode(s); err == nil {
		if len(raw) == ed25519.PrivateKeySize {
			return ed25519.PrivateKey(raw), nil
		}
		if len(raw) == ed25519.SeedSize {
			return ed25519.NewKeyFromSeed(raw), nil
		}
	}
	return nil, fmt.Errorf("unrecognized ed25519 secret encoding")
}

// deterministicMemoPayload builds the canonical memo string anchored on-chain.
func deterministicMemoPayload(kind, tradeID string, digest [32]byte, extra uint32) string {
	var b bytes.Buffer
	b.WriteString("verinode:")
	b.WriteString(kind)
	b.WriteString(":")
	b.WriteString(tradeID)
	b.WriteString(":")
	b.WriteString(hex.EncodeToString(digest[:]))
	if extra > 0 {
		var e [4]byte
		binary.BigEndian.PutUint32(e[:], extra)
		b.WriteString(":")
		b.WriteString(hex.EncodeToString(e[:]))
	}
	return b.String()
}
