package phase3

import (
	"time"
)

// replacement.go implements automated substitution routing (Phase 3, docs/02 §Phase3
// and docs/06 §1.8): when a delivery fails acceptance, the system proposes a
// same-grade replacement, then a buyer-approved superior grade, then a cash remedy —
// never a silent failure and never a mixed-card "H100 equivalent" (Invariant: grade
// is exact or it isn't delivered; docs/02 §3 non-goals).

// CandidateBlock is an available inventory block considered for substitution.
type CandidateBlock struct {
	BlockID     string    `json:"block_id"`
	GradeID     string    `json:"grade_id"`
	SellerID    string    `json:"seller_id"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	PriceCents  int64     `json:"price_cents"`
}

// SubstitutionRequest describes the failed contract needing a remedy.
type SubstitutionRequest struct {
	ContractID       string           `json:"contract_id"`
	FailedGradeID    string           `json:"failed_grade_id"`
	WindowStart      time.Time        `json:"window_start"`
	WindowEnd        time.Time        `json:"window_end"`
	OriginalPrice    int64            `json:"original_price_cents"`
	CureDeadline     time.Time        `json:"cure_deadline"`
	BuyerApprovesSup bool             `json:"buyer_approves_superior"`
	Candidates       []CandidateBlock `json:"candidates"`
}

// SubstitutionOffer is the routed remedy.
type SubstitutionOffer struct {
	Kind          string    `json:"kind"` // same_grade | superior_grade | cash_remedy
	BlockID       string    `json:"block_id,omitempty"`
	OfferedGrade  string    `json:"offered_grade_id,omitempty"`
	CashRemedyCts int64     `json:"cash_remedy_cents,omitempty"`
	CureDeadline  time.Time `json:"cure_deadline"`
	Rationale     string    `json:"rationale"`
}

// gradeRank orders known grades so "superior" is well-defined. Higher = better.
// Only exact grades are ever offered — this rank never blends cards.
var gradeRank = map[string]int{
	"H100-SXM-8XNV": 10,
	"H200-SXM-8XNV": 20,
	"B200-SXM-8XNV": 30,
}

// RouteReplacement applies the replacement hierarchy and returns the best remedy
// that fits inside the cure window. It never returns nil: a cash remedy is the
// guaranteed floor so delivery failure is never silent.
func RouteReplacement(req SubstitutionRequest) SubstitutionOffer {
	fits := func(c CandidateBlock) bool {
		return !c.WindowStart.After(req.WindowStart) && !c.WindowEnd.Before(req.WindowEnd) &&
			!time.Now().After(req.CureDeadline)
	}

	// 1. Same-grade replacement (preferred): cheapest fitting block of the exact grade.
	var bestSame *CandidateBlock
	for i := range req.Candidates {
		c := req.Candidates[i]
		if c.GradeID == req.FailedGradeID && fits(c) {
			if bestSame == nil || c.PriceCents < bestSame.PriceCents {
				bestSame = &req.Candidates[i]
			}
		}
	}
	if bestSame != nil {
		return SubstitutionOffer{
			Kind:         "same_grade",
			BlockID:      bestSame.BlockID,
			OfferedGrade: bestSame.GradeID,
			CureDeadline: req.CureDeadline,
			Rationale:    "Exact-grade replacement available within the cure window; no price change to the buyer.",
		}
	}

	// 2. Superior grade — only with explicit buyer approval (never auto-downgrade,
	//    never a mixed-card substitute). Lowest-rank grade that is strictly superior.
	if req.BuyerApprovesSup {
		failedRank := gradeRank[req.FailedGradeID]
		var bestSup *CandidateBlock
		for i := range req.Candidates {
			c := req.Candidates[i]
			if gradeRank[c.GradeID] > failedRank && fits(c) {
				if bestSup == nil || gradeRank[c.GradeID] < gradeRank[bestSup.GradeID] {
					bestSup = &req.Candidates[i]
				}
			}
		}
		if bestSup != nil {
			return SubstitutionOffer{
				Kind:         "superior_grade",
				BlockID:      bestSup.BlockID,
				OfferedGrade: bestSup.GradeID,
				CureDeadline: req.CureDeadline,
				Rationale:    "No same-grade block fits; buyer pre-approved a superior grade at no extra charge.",
			}
		}
	}

	// 3. Cash remedy floor — guaranteed, never silent.
	return SubstitutionOffer{
		Kind:          "cash_remedy",
		CashRemedyCts: req.OriginalPrice,
		CureDeadline:  req.CureDeadline,
		Rationale:     "No qualifying replacement within the cure window; full contract value is refunded per the make-good SLA.",
	}
}
