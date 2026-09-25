package telemetry_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/telemetry"
)

func TestHardwareAttestation_CryptographicVerification(t *testing.T) {
	t.Parallel()

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 keypair: %v", err)
	}

	otherPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate second keypair: %v", err)
	}

	report := telemetry.HardwareReport{
		AgentID:              "agent_supplier_node_001",
		BlockID:              "block_h100_cluster_77",
		Timestamp:            time.Now().UTC().Truncate(time.Second),
		GPUSKU:               "NVIDIA H100 SXM 80GB",
		GPUCount:             8,
		MemoryGB:             640,
		Topology:             "SXM5/HGX 8x NVLink 4.0 / NVSwitch",
		DriverVersion:        "535.129.03",
		ECCErrorsUnrecovered: 0,
		SequenceNumber:       1,
	}

	payload, err := report.CanonicalPayload()
	if err != nil {
		t.Fatalf("failed to get canonical payload: %v", err)
	}

	signature := ed25519.Sign(privKey, payload)

	t.Run("Valid signature passes", func(t *testing.T) {
		t.Parallel()
		if err := telemetry.VerifyHardwareAttestation(pubKey, report, signature); err != nil {
			t.Errorf("expected valid signature, got error: %v", err)
		}
	})

	t.Run("Tampered report payload fails verification", func(t *testing.T) {
		t.Parallel()
		tamperedReport := report
		tamperedReport.GPUCount = 4 // Attacker spoofed 8 -> 4 or vice versa

		err := telemetry.VerifyHardwareAttestation(pubKey, tamperedReport, signature)
		if !errors.Is(err, telemetry.ErrInvalidSignature) {
			t.Errorf("expected ErrInvalidSignature on tampered report, got: %v", err)
		}
	})

	t.Run("Wrong public key fails verification", func(t *testing.T) {
		t.Parallel()
		err := telemetry.VerifyHardwareAttestation(otherPubKey, report, signature)
		if !errors.Is(err, telemetry.ErrInvalidSignature) {
			t.Errorf("expected ErrInvalidSignature with wrong key, got: %v", err)
		}
	})
}

func TestValidateAgainstGrade_TableDriven(t *testing.T) {
	t.Parallel()

	baseReport := telemetry.HardwareReport{
		AgentID:              "agent_001",
		BlockID:              "block_001",
		Timestamp:            time.Now().UTC(),
		GPUSKU:               "NVIDIA H100 SXM 80GB",
		GPUCount:             8,
		MemoryGB:             640,
		Topology:             "SXM5/HGX 8x NVLink 4.0 / NVSwitch",
		DriverVersion:        "535.129.03",
		ECCErrorsUnrecovered: 0,
	}

	tests := []struct {
		name        string
		modifier    func(r *telemetry.HardwareReport)
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "Valid: 8x H100 SXM healthy matches spec",
			modifier:    func(r *telemetry.HardwareReport) {},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "Invalid: insufficient healthy GPUs (7 of 8)",
			modifier: func(r *telemetry.HardwareReport) {
				r.GPUCount = 7
			},
			wantErr:     true,
			expectedErr: telemetry.ErrInsufficientGPUs,
		},
		{
			name: "Invalid: insufficient memory (320 GB)",
			modifier: func(r *telemetry.HardwareReport) {
				r.MemoryGB = 320
			},
			wantErr:     true,
			expectedErr: telemetry.ErrInsufficientMemory,
		},
		{
			name: "Invalid: topology mismatch (PCIe instead of NVLink/SXM)",
			modifier: func(r *telemetry.HardwareReport) {
				r.Topology = "PCIe Gen5 x16 (no NVLink)"
			},
			wantErr:     true,
			expectedErr: telemetry.ErrTopologyMismatch,
		},
		{
			name: "Invalid: unrecovered hardware ECC memory errors",
			modifier: func(r *telemetry.HardwareReport) {
				r.ECCErrorsUnrecovered = 2
			},
			wantErr:     true,
			expectedErr: telemetry.ErrHardwareECCErrors,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := baseReport
			tt.modifier(&r)

			err := telemetry.ValidateAgainstGrade(r, 8, 640, "NVLink")
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateAgainstGrade() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestCanaryResult_VerificationAndFloor(t *testing.T) {
	t.Parallel()

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	canary := telemetry.CanaryResult{
		AgentID:           "agent_canary_01",
		BlockID:           "block_01",
		Timestamp:         time.Now().UTC(),
		TestName:          "nccl_allreduce_synthetic",
		Passed:            true,
		NCCLAllReduceGBPS: 425.5, // Floor is 400.0
	}

	payload, err := canary.CanonicalPayload()
	if err != nil {
		t.Fatalf("failed to get canonical payload: %v", err)
	}

	sig := ed25519.Sign(privKey, payload)

	if err := telemetry.VerifyCanaryResult(pubKey, canary, sig); err != nil {
		t.Errorf("canary signature verification failed: %v", err)
	}

	if err := telemetry.ValidateCanaryAgainstFloor(canary, 400.0); err != nil {
		t.Errorf("canary floor check failed unexpectedly: %v", err)
	}

	// Below floor must fail
	slowCanary := canary
	slowCanary.NCCLAllReduceGBPS = 280.0 // Slow/degraded interconnect
	if err := telemetry.ValidateCanaryAgainstFloor(slowCanary, 400.0); !errors.Is(err, telemetry.ErrCanaryBenchmarkFailed) {
		t.Errorf("expected ErrCanaryBenchmarkFailed for slow canary, got: %v", err)
	}
}
