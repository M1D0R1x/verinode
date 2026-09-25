use anchor_lang::prelude::*;

declare_id!("VnodE7ZcEwXJ8gqJ2MhFhN3CqU9uWv4kL5Y7Xz8A1bC");

#[program]
pub mod verinode_registry {
    use super::*;

    /// Initialize a canonical on-chain physical trade envelope (PDA keyed to trade_id)
    pub fn initialize_trade_envelope(
        ctx: Context<InitializeTradeEnvelope>,
        trade_id: [u8; 36],
        grade_id: [u8; 16],
        price_cents: u64,
        window_start: i64,
        window_end: i64,
    ) -> Result<()> {
        require!(window_end > window_start, VerinodeError::InvalidTimeWindow);
        require!(price_cents > 0, VerinodeError::InvalidPrice);

        let envelope = &mut ctx.accounts.trade_envelope;
        let clock = Clock::get()?;

        envelope.bump = ctx.bumps.trade_envelope;
        envelope.trade_id = trade_id;
        envelope.buyer = ctx.accounts.buyer.key();
        envelope.seller = ctx.accounts.seller.key();
        envelope.grade_id = grade_id;
        envelope.state = TradeState::ContractPending;
        envelope.price_cents = price_cents;
        envelope.window_start = window_start;
        envelope.window_end = window_end;
        envelope.latest_attestation_digest = [0u8; 32];
        envelope.nccl_allreduce_gbps = 0;
        envelope.canary_passed = false;
        envelope.created_at = clock.unix_timestamp;
        envelope.updated_at = clock.unix_timestamp;

        msg!("Canonical trade envelope initialized on Solana for trade_id: {:?}", trade_id);
        Ok(())
    }

    /// Record a verified cryptographic hardware attestation and NCCL canary benchmark floor result
    pub fn record_attestation(
        ctx: Context<RecordAttestation>,
        attestation_digest: [u8; 32],
        nccl_gbps: u32,
        canary_passed: bool,
    ) -> Result<()> {
        let envelope = &mut ctx.accounts.trade_envelope;
        let clock = Clock::get()?;

        // Benchmark floor for benchmark grade H100-SXM-8XNV is 400 GB/s
        if canary_passed && nccl_gbps < 400 {
            return err!(VerinodeError::BelowBenchmarkFloor);
        }

        envelope.latest_attestation_digest = attestation_digest;
        envelope.nccl_allreduce_gbps = nccl_gbps;
        envelope.canary_passed = canary_passed;
        envelope.updated_at = clock.unix_timestamp;

        msg!(
            "Recorded telemetry attestation for trade. NCCL: {} GB/s, Passed: {}",
            nccl_gbps,
            canary_passed
        );
        Ok(())
    }

    /// Advance contract lifecycle state mirroring the off-chain Postgres state machine
    pub fn advance_trade_state(
        ctx: Context<AdvanceTradeState>,
        new_state: TradeState,
    ) -> Result<()> {
        let envelope = &mut ctx.accounts.trade_envelope;
        let clock = Clock::get()?;

        // Enforce valid state progression mirroring off-chain Verinode state machine
        require!(
            is_valid_transition(envelope.state, new_state),
            VerinodeError::IllegalStateTransition
        );

        envelope.state = new_state;
        envelope.updated_at = clock.unix_timestamp;

        msg!("Advanced trade state to {:?}", new_state);
        Ok(())
    }
}

// -----------------------------------------------------------------------------
// State Machine Transitions & Invariants
// -----------------------------------------------------------------------------

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Copy, PartialEq, Eq, Debug, InitSpace)]
pub enum TradeState {
    ContractPending,
    FundedSecured,
    Scheduled,
    DeliveryTest,
    Live,
    Completed,
    Settled,
    Cancelled,
    FailedDelivery,
    Cure,
    Substituted,
    ClaimOpen,
    Disputed,
    Terminated,
}

pub fn is_valid_transition(current: TradeState, next: TradeState) -> bool {
    match (current, next) {
        (TradeState::ContractPending, TradeState::FundedSecured) => true,
        (TradeState::ContractPending, TradeState::Cancelled) => true,
        (TradeState::FundedSecured, TradeState::Scheduled) => true,
        (TradeState::FundedSecured, TradeState::Cancelled) => true,
        (TradeState::Scheduled, TradeState::DeliveryTest) => true,
        (TradeState::DeliveryTest, TradeState::Live) => true,
        (TradeState::DeliveryTest, TradeState::FailedDelivery) => true,
        (TradeState::Live, TradeState::Completed) => true,
        (TradeState::Live, TradeState::ClaimOpen) => true,
        (TradeState::Completed, TradeState::Settled) => true,
        (TradeState::ClaimOpen, TradeState::Disputed) => true,
        (TradeState::Disputed, TradeState::Settled) => true,
        (TradeState::Disputed, TradeState::Terminated) => true,
        _ => false,
    }
}

// -----------------------------------------------------------------------------
// Account Contexts
// -----------------------------------------------------------------------------

#[derive(Accounts)]
#[instruction(trade_id: [u8; 36])]
pub struct InitializeTradeEnvelope<'info> {
    #[account(
        init,
        payer = payer,
        space = 8 + TradeEnvelope::INIT_SPACE,
        seeds = [b"trade_envelope", trade_id.as_ref()],
        bump
    )]
    pub trade_envelope: Account<'info, TradeEnvelope>,

    /// CHECK: Buyer legal identity public key mirror
    pub buyer: AccountInfo<'info>,

    /// CHECK: Seller supplier public key mirror
    pub seller: AccountInfo<'info>,

    #[account(mut)]
    pub payer: Signer<'info>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct RecordAttestation<'info> {
    #[account(
        mut,
        seeds = [b"trade_envelope", trade_envelope.trade_id.as_ref()],
        bump = trade_envelope.bump,
        has_one = seller @ VerinodeError::UnauthorizedSigner
    )]
    pub trade_envelope: Account<'info, TradeEnvelope>,

    pub seller: Signer<'info>,
}

#[derive(Accounts)]
pub struct AdvanceTradeState<'info> {
    #[account(
        mut,
        seeds = [b"trade_envelope", trade_envelope.trade_id.as_ref()],
        bump = trade_envelope.bump
    )]
    pub trade_envelope: Account<'info, TradeEnvelope>,

    pub authority: Signer<'info>,
}

// -----------------------------------------------------------------------------
// Account Data Layout
// -----------------------------------------------------------------------------

#[account]
#[derive(InitSpace)]
pub struct TradeEnvelope {
    pub bump: u8,
    pub trade_id: [u8; 36],
    pub buyer: Pubkey,
    pub seller: Pubkey,
    pub grade_id: [u8; 16],
    pub state: TradeState,
    pub price_cents: u64,
    pub window_start: i64,
    pub window_end: i64,
    pub latest_attestation_digest: [u8; 32],
    pub nccl_allreduce_gbps: u32,
    pub canary_passed: bool,
    pub created_at: i64,
    pub updated_at: i64,
}

// -----------------------------------------------------------------------------
// Errors
// -----------------------------------------------------------------------------

#[error_code]
pub enum VerinodeError {
    #[msg("window_end must be strictly greater than window_start")]
    InvalidTimeWindow,
    #[msg("price_cents must be strictly greater than zero")]
    InvalidPrice,
    #[msg("Signer is not authorized to submit hardware attestation for this trade")]
    UnauthorizedSigner,
    #[msg("NCCL all-reduce bandwidth is below benchmark floor (400 GB/s)")]
    BelowBenchmarkFloor,
    #[msg("Illegal state transition rejected by Verinode 18-state transition machine")]
    IllegalStateTransition,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_transitions() {
        assert!(is_valid_transition(TradeState::ContractPending, TradeState::FundedSecured));
        assert!(is_valid_transition(TradeState::FundedSecured, TradeState::Scheduled));
        assert!(is_valid_transition(TradeState::Scheduled, TradeState::DeliveryTest));
        assert!(is_valid_transition(TradeState::DeliveryTest, TradeState::Live));
        assert!(is_valid_transition(TradeState::Live, TradeState::Completed));
        assert!(is_valid_transition(TradeState::Completed, TradeState::Settled));
    }

    #[test]
    fn test_illegal_transitions() {
        // Cannot jump directly from Live to Settled
        assert!(!is_valid_transition(TradeState::Live, TradeState::Settled));
        // Terminal states cannot be reactivated
        assert!(!is_valid_transition(TradeState::Settled, TradeState::ContractPending));
        assert!(!is_valid_transition(TradeState::Terminated, TradeState::Live));
    }
}

