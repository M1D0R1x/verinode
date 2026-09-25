package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/M1D0R1x/verinode/services/internal/contract"
	"github.com/M1D0R1x/verinode/services/internal/inventory"
	"github.com/M1D0R1x/verinode/services/internal/ledger"
	"github.com/M1D0R1x/verinode/services/internal/participant"
	"github.com/M1D0R1x/verinode/services/internal/rfq"
	"github.com/M1D0R1x/verinode/services/internal/telemetry"
)

type Server struct {
	participantRepo *participant.Repository
	inventoryRepo   *inventory.Repository
	rfqRepo         *rfq.Repository
	contractRepo    *contract.Repository
	ledgerRepo      *ledger.Repository
	logger          *slog.Logger
	handler         http.Handler
}

func NewServer(
	pRepo *participant.Repository,
	iRepo *inventory.Repository,
	rRepo *rfq.Repository,
	cRepo *contract.Repository,
	lRepo *ledger.Repository,
	logger *slog.Logger,
) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Server{
		participantRepo: pRepo,
		inventoryRepo:   iRepo,
		rfqRepo:         rRepo,
		contractRepo:    cRepo,
		ledgerRepo:      lRepo,
		logger:          logger,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)
	s.handler = corsMiddleware(mux)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Health & Benchmark Grades
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /v1/grades", s.handleGetGrades)

	// Participant Endpoints
	mux.HandleFunc("POST /v1/participants", s.handleCreateParticipant)
	mux.HandleFunc("GET /v1/participants", s.handleListParticipants)
	mux.HandleFunc("GET /v1/participants/{id}", s.handleGetParticipant)

	// Inventory Endpoints
	mux.HandleFunc("POST /v1/inventory/blocks", s.handleCreateInventoryBlock)
	mux.HandleFunc("GET /v1/inventory/blocks", s.handleListInventoryBlocks)

	// RFQ & Quotes Endpoints
	mux.HandleFunc("POST /v1/rfqs", s.handleCreateRFQ)
	mux.HandleFunc("GET /v1/rfqs", s.handleListRFQs)
	mux.HandleFunc("GET /v1/rfqs/{id}", s.handleGetRFQ)
	mux.HandleFunc("POST /v1/rfqs/{id}/quotes", s.handleCreateQuote)
	mux.HandleFunc("GET /v1/rfqs/{id}/quotes", s.handleListQuotes)
	mux.HandleFunc("POST /v1/rfqs/{id}/quotes/{quoteId}/accept", s.handleAcceptQuote)

	// Contract Endpoints
	mux.HandleFunc("GET /v1/contracts/{id}", s.handleGetContract)
	mux.HandleFunc("POST /v1/contracts/{id}/advance", s.handleAdvanceContract)
	mux.HandleFunc("POST /v1/contracts/validate-transition", s.handleValidateTransition)

	// Telemetry Verification
	mux.HandleFunc("POST /v1/telemetry/attestations/verify", s.handleVerifyTelemetry)
}

// -----------------------------------------------------------------------------
// Route Handlers
// -----------------------------------------------------------------------------

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "verinode-api-gateway",
		"version":   "0.1.0",
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handleGetGrades(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	grades := []map[string]interface{}{
		{
			"id":                    "H100-SXM-8XNV",
			"gpu_sku":               "NVIDIA H100 SXM 80GB",
			"min_memory_gb":         640,
			"topology":              "SXM5/HGX 8x NVLink 4.0 / NVSwitch (900 GB/s bidirectional)",
			"min_healthy_gpu_count": 8,
			"benchmark_floor": map[string]interface{}{
				"nccl_allreduce_gb_per_sec":  400.0,
				"min_cuda_driver":            "535.129.03",
				"max_ecc_unrecovered_errors": 0,
			},
			"min_cpu_cores": 112,
			"min_ram_gb":    1024,
			"min_nvme_perf": 100000,
		},
	}
	_ = json.NewEncoder(w).Encode(grades)
}

func (s *Server) handleCreateParticipant(w http.ResponseWriter, r *http.Request) {
	if s.participantRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	var req participant.Participant
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	if err := s.participantRepo.Create(r.Context(), &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Participant Creation Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(req)
}

func (s *Server) handleListParticipants(w http.ResponseWriter, r *http.Request) {
	if s.participantRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	participants, err := s.participantRepo.List(r.Context())
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Listing Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(participants)
}

func (s *Server) handleGetParticipant(w http.ResponseWriter, r *http.Request) {
	if s.participantRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	id := r.PathValue("id")
	p, err := s.participantRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, participant.ErrParticipantNotFound) {
			writeProblem(w, http.StatusNotFound, "Participant Not Found", id)
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Database Query Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

func (s *Server) handleCreateInventoryBlock(w http.ResponseWriter, r *http.Request) {
	if s.inventoryRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	var req inventory.InventoryBlock
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	if err := s.inventoryRepo.Create(r.Context(), &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Inventory Block Creation Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(req)
}

func (s *Server) handleListInventoryBlocks(w http.ResponseWriter, r *http.Request) {
	if s.inventoryRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	gradeID := r.URL.Query().Get("grade_id")
	if gradeID == "" {
		gradeID = "H100-SXM-8XNV"
	}
	region := r.URL.Query().Get("region_bucket")
	if region == "" {
		region = "US-East"
	}

	start := time.Now().UTC()
	end := start.Add(168 * time.Hour)

	blocks, err := s.inventoryRepo.FindAvailableBlocks(r.Context(), gradeID, region, start, end)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Inventory Query Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(blocks)
}

func (s *Server) handleCreateRFQ(w http.ResponseWriter, r *http.Request) {
	if s.rfqRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	var req rfq.RFQ
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	if err := s.rfqRepo.CreateRFQ(r.Context(), &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "RFQ Creation Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(req)
}

func (s *Server) handleListRFQs(w http.ResponseWriter, r *http.Request) {
	// Fallback/standard listing
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]map[string]interface{}{})
}

func (s *Server) handleGetRFQ(w http.ResponseWriter, r *http.Request) {
	if s.rfqRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	id := r.PathValue("id")
	req, err := s.rfqRepo.GetRFQByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, rfq.ErrRFQNotFound) {
			writeProblem(w, http.StatusNotFound, "RFQ Not Found", id)
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Query Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(req)
}

func (s *Server) handleCreateQuote(w http.ResponseWriter, r *http.Request) {
	if s.rfqRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	rfqID := r.PathValue("id")
	var q rfq.Quote
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}
	q.RFQID = rfqID

	if err := s.rfqRepo.CreateQuote(r.Context(), &q); err != nil {
		writeProblem(w, http.StatusBadRequest, "Quote Submission Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(q)
}

func (s *Server) handleListQuotes(w http.ResponseWriter, r *http.Request) {
	if s.rfqRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Database repository is not connected")
		return
	}

	rfqID := r.PathValue("id")
	quotes, err := s.rfqRepo.GetQuotesByRFQ(r.Context(), rfqID)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Quote Query Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(quotes)
}

func (s *Server) handleAcceptQuote(w http.ResponseWriter, r *http.Request) {
	if s.rfqRepo == nil || s.contractRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Repositories not configured")
		return
	}

	quoteID := r.PathValue("quoteId")
	var req struct {
		BuyerID string `json:"buyer_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.BuyerID == "" {
		writeProblem(w, http.StatusBadRequest, "Missing Buyer ID", "buyer_id is required in request body")
		return
	}

	// 1. Atomically lock quote & RFQ, transition status
	details, err := s.rfqRepo.AcceptQuote(r.Context(), quoteID, req.BuyerID)
	if err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Acceptance Rejected", err.Error())
		return
	}

	// 2. Spawn contract with canonical trade_id
	contractRec := &contract.ContractRecord{
		RFQID:           &details.RFQID,
		QuoteID:         &details.QuoteID,
		BuyerID:         details.BuyerID,
		SellerID:        details.SellerID,
		GradeID:         details.GradeID,
		TemplateVersion: "v1.0.0-institutional",
		State:           contract.StateContractPending,
	}

	if err := s.contractRepo.CreateContract(r.Context(), contractRec); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Contract Spawning Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "accepted",
		"contract_id": contractRec.ID, // canonical trade_id
		"contract":    contractRec,
	})
}

func (s *Server) handleGetContract(w http.ResponseWriter, r *http.Request) {
	if s.contractRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Contract repository not configured")
		return
	}

	tradeID := r.PathValue("id")
	c, err := s.contractRepo.GetByID(r.Context(), tradeID)
	if err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			writeProblem(w, http.StatusNotFound, "Contract Not Found", tradeID)
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Database Error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}

func (s *Server) handleAdvanceContract(w http.ResponseWriter, r *http.Request) {
	if s.contractRepo == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Database Unavailable", "Contract repository not configured")
		return
	}

	tradeID := r.PathValue("id")
	var req struct {
		NextState      string                 `json:"next_state"`
		Actor          string                 `json:"actor"`
		Reason         string                 `json:"reason"`
		IdempotencyKey string                 `json:"idempotency_key"`
		EvidenceHash   string                 `json:"evidence_hash,omitempty"`
		AuthDecision   map[string]interface{} `json:"authorization_decision,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	nextState, err := contract.ParseState(req.NextState)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Invalid Next State", err.Error())
		return
	}

	evt := contract.TransitionEvent{
		Actor:                 req.Actor,
		Reason:                req.Reason,
		IdempotencyKey:        req.IdempotencyKey,
		EvidenceHash:          req.EvidenceHash,
		AuthorizationDecision: req.AuthDecision,
	}

	if err := s.contractRepo.AdvanceContractState(r.Context(), tradeID, nextState, evt); err != nil {
		status := http.StatusUnprocessableEntity
		if errors.Is(err, contract.ErrContractNotFound) {
			status = http.StatusNotFound
		}
		writeProblem(w, status, "State Transition Rejected", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"trade_id":   tradeID,
		"next_state": nextState,
		"updated_at": time.Now().UTC(),
	})
}

func (s *Server) handleValidateTransition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentState   string                 `json:"current_state"`
		NextState      string                 `json:"next_state"`
		Actor          string                 `json:"actor"`
		Reason         string                 `json:"reason"`
		IdempotencyKey string                 `json:"idempotency_key"`
		AuthDecision   map[string]interface{} `json:"authorization_decision"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	current, err := contract.ParseState(req.CurrentState)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Invalid Current State", err.Error())
		return
	}

	next, err := contract.ParseState(req.NextState)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Invalid Target State", err.Error())
		return
	}

	evt := contract.TransitionEvent{
		Actor:                 req.Actor,
		Reason:                req.Reason,
		IdempotencyKey:        req.IdempotencyKey,
		AuthorizationDecision: req.AuthDecision,
	}

	if err := contract.ValidateTransition(current, next, evt); err != nil {
		status := http.StatusUnprocessableEntity
		if errors.Is(err, contract.ErrMissingActor) ||
			errors.Is(err, contract.ErrMissingReason) ||
			errors.Is(err, contract.ErrMissingIdempotency) {
			status = http.StatusBadRequest
		}
		writeProblem(w, status, "Transition Rejected", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":         true,
		"current_state": current,
		"next_state":    next,
		"timestamp":     time.Now().UTC(),
	})
}

func (s *Server) handleVerifyTelemetry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Report    telemetry.HardwareReport `json:"report"`
		Signature string                   `json:"signature"`
		PublicKey string                   `json:"public_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Malformed JSON", err.Error())
		return
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		writeProblem(w, http.StatusBadRequest, "Invalid Public Key", "Public key must be base64-encoded ed25519 key")
		return
	}

	sigBytes, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		writeProblem(w, http.StatusBadRequest, "Invalid Signature", "Signature must be base64-encoded ed25519 signature")
		return
	}

	if err := telemetry.VerifyHardwareAttestation(ed25519.PublicKey(pubKeyBytes), req.Report, sigBytes); err != nil {
		writeProblem(w, http.StatusUnauthorized, "Cryptographic Verification Failed", err.Error())
		return
	}

	topologyRequirement := "NVLink"
	if strings.Contains(strings.ToLower(req.Report.Topology), "pcie") {
		topologyRequirement = "PCIe"
	}

	if err := telemetry.ValidateAgainstGrade(req.Report, 8, 640, topologyRequirement); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Hardware Grade Compliance Failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"verified":        true,
		"agent_id":        req.Report.AgentID,
		"block_id":        req.Report.BlockID,
		"grade_compliant": true,
		"timestamp":       time.Now().UTC(),
	})
}
