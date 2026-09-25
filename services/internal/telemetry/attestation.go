package telemetry

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidSignature      = errors.New("telemetry: cryptographic signature verification failed")
	ErrInsufficientGPUs      = errors.New("telemetry: hardware report has insufficient healthy GPUs")
	ErrInsufficientMemory    = errors.New("telemetry: hardware report has insufficient GPU memory")
	ErrTopologyMismatch      = errors.New("telemetry: NVLink/interconnect topology does not match grade specification")
	ErrHardwareECCErrors     = errors.New("telemetry: hardware report has unrecovered ECC memory errors")
	ErrCanaryBenchmarkFailed = errors.New("telemetry: canary synthetic benchmark failed or below floor threshold")
	ErrMalformedReport       = errors.New("telemetry: malformed or empty telemetry report")
)

// HardwareReport represents host-level hardware telemetry signed by the supplier's telemetry agent.
// STRICT INVARIANT: ZERO customer workload data (model weights, training code, prompts) may be included.
type HardwareReport struct {
	AgentID              string    `json:"agent_id"`
	BlockID              string    `json:"block_id"`
	Timestamp            time.Time `json:"timestamp"`
	GPUSKU               string    `json:"gpu_sku"`
	GPUCount             int       `json:"gpu_count"`
	MemoryGB             int       `json:"memory_gb"`
	Topology             string    `json:"topology"`
	DriverVersion        string    `json:"driver_version"`
	ECCErrorsUnrecovered int       `json:"ecc_errors_unrecovered"`
	SequenceNumber       uint64    `json:"sequence_number"`
}

// CanaryResult captures the output of a synthetic benchmark (e.g. NCCL all-reduce) executed on the host.
type CanaryResult struct {
	AgentID           string                 `json:"agent_id"`
	BlockID           string                 `json:"block_id"`
	Timestamp         time.Time              `json:"timestamp"`
	TestName          string                 `json:"test_name"`
	Passed            bool                   `json:"passed"`
	NCCLAllReduceGBPS float64                `json:"nccl_allreduce_gb_per_sec"`
	Metrics           map[string]interface{} `json:"metrics,omitempty"`
}

// CanonicalPayload returns a deterministic JSON byte slice for cryptographic signing.
func (h HardwareReport) CanonicalPayload() ([]byte, error) {
	if h.AgentID == "" || h.BlockID == "" {
		return nil, ErrMalformedReport
	}
	return json.Marshal(h)
}

// CanonicalPayload returns a deterministic JSON byte slice for cryptographic signing.
func (c CanaryResult) CanonicalPayload() ([]byte, error) {
	if c.AgentID == "" || c.BlockID == "" {
		return nil, ErrMalformedReport
	}
	return json.Marshal(c)
}

// VerifyHardwareAttestation verifies the cryptographic signature of a hardware report.
func VerifyHardwareAttestation(pubKey ed25519.PublicKey, report HardwareReport, signature []byte) error {
	payload, err := report.CanonicalPayload()
	if err != nil {
		return err
	}
	if len(signature) != ed25519.SignatureSize {
		return ErrInvalidSignature
	}
	if !ed25519.Verify(pubKey, payload, signature) {
		return ErrInvalidSignature
	}
	return nil
}

// VerifyCanaryResult verifies the cryptographic signature of a canary test result.
func VerifyCanaryResult(pubKey ed25519.PublicKey, canary CanaryResult, signature []byte) error {
	payload, err := canary.CanonicalPayload()
	if err != nil {
		return err
	}
	if len(signature) != ed25519.SignatureSize {
		return ErrInvalidSignature
	}
	if !ed25519.Verify(pubKey, payload, signature) {
		return ErrInvalidSignature
	}
	return nil
}

// ValidateAgainstGrade ensures the hardware meets the required grade threshold.
func ValidateAgainstGrade(
	report HardwareReport,
	minGPUCount int,
	minMemoryGB int,
	requiredTopologySubstr string,
) error {
	if report.GPUCount < minGPUCount {
		return fmt.Errorf("%w: found %d, required at least %d", ErrInsufficientGPUs, report.GPUCount, minGPUCount)
	}
	if report.MemoryGB < minMemoryGB {
		return fmt.Errorf("%w: found %d GB, required at least %d GB", ErrInsufficientMemory, report.MemoryGB, minMemoryGB)
	}

	topoLower := strings.ToLower(report.Topology)
	if requiredTopologySubstr != "" {
		reqLower := strings.ToLower(requiredTopologySubstr)
		// Check for presence of required topology and absence of negation
		if !strings.Contains(topoLower, reqLower) || strings.Contains(topoLower, "no "+reqLower) || strings.Contains(topoLower, "without "+reqLower) {
			return fmt.Errorf("%w: %q does not satisfy %q", ErrTopologyMismatch, report.Topology, requiredTopologySubstr)
		}
	}

	if report.ECCErrorsUnrecovered > 0 {
		return fmt.Errorf("%w: detected %d unrecovered ECC errors", ErrHardwareECCErrors, report.ECCErrorsUnrecovered)
	}
	return nil
}

// ValidateCanaryAgainstFloor ensures the canary execution meets the grade benchmark floor.
func ValidateCanaryAgainstFloor(canary CanaryResult, floorGBPS float64) error {
	if !canary.Passed {
		return fmt.Errorf("%w: canary test suite reported failure", ErrCanaryBenchmarkFailed)
	}
	if canary.NCCLAllReduceGBPS < floorGBPS {
		return fmt.Errorf("%w: NCCL bandwidth %.2f GB/s below minimum floor %.2f GB/s",
			ErrCanaryBenchmarkFailed, canary.NCCLAllReduceGBPS, floorGBPS)
	}
	return nil
}
