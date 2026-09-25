package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/telemetry"
)

func main() {
	agentID := flag.String("agent-id", "agent_mock_h100_01", "Supplier telemetry agent identifier")
	blockID := flag.String("block-id", "block_reservation_test_01", "Inventory block identifier")
	gpuSKU := flag.String("sku", "NVIDIA H100 SXM 80GB", "GPU SKU")
	gpuCount := flag.Int("count", 8, "Number of healthy GPUs")
	memoryGB := flag.Int("mem", 640, "Total GPU memory in GB")
	topology := flag.String("topology", "SXM5/HGX 8x NVLink 4.0 / NVSwitch", "GPU interconnect topology")
	verifyOnly := flag.Bool("verify", false, "Verify an incoming report from stdin")
	flag.Parse()

	if *verifyOnly {
		log.Println("Verify mode: reading payload from stdin...")
		var input struct {
			Report    telemetry.HardwareReport `json:"report"`
			Signature string                   `json:"signature"`
			PublicKey string                   `json:"public_key"`
		}
		if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
			log.Fatalf("failed to decode verification input: %v", err)
		}

		pubKeyBytes, err := base64.StdEncoding.DecodeString(input.PublicKey)
		if err != nil {
			log.Fatalf("invalid base64 public key: %v", err)
		}
		sigBytes, err := base64.StdEncoding.DecodeString(input.Signature)
		if err != nil {
			log.Fatalf("invalid base64 signature: %v", err)
		}

		if err := telemetry.VerifyHardwareAttestation(ed25519.PublicKey(pubKeyBytes), input.Report, sigBytes); err != nil {
			log.Fatalf("ATTESTATION REJECTED: %v", err)
		}
		if err := telemetry.ValidateAgainstGrade(input.Report, 8, 640, "NVLink"); err != nil {
			log.Fatalf("GRADE COMPLIANCE FAILED: %v", err)
		}
		fmt.Println("✓ Attestation cryptographically verified and meets Grade floor")
		return
	}

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate key: %v", err)
	}

	report := telemetry.HardwareReport{
		AgentID:              *agentID,
		BlockID:              *blockID,
		Timestamp:            time.Now().UTC(),
		GPUSKU:               *gpuSKU,
		GPUCount:             *gpuCount,
		MemoryGB:             *memoryGB,
		Topology:             *topology,
		DriverVersion:        "535.129.03",
		ECCErrorsUnrecovered: 0,
		SequenceNumber:       1,
	}

	payload, err := report.CanonicalPayload()
	if err != nil {
		log.Fatalf("failed to serialize payload: %v", err)
	}

	signature := ed25519.Sign(privKey, payload)

	output := map[string]interface{}{
		"report":     report,
		"signature":  base64.StdEncoding.EncodeToString(signature),
		"public_key": base64.StdEncoding.EncodeToString(pubKey),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		log.Fatalf("failed to output JSON: %v", err)
	}
}
