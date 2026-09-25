// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title VerinodeRegistry
 * @notice Enterprise EVM & Arbitrum Bilateral Capacity Reservation and Telemetry Attestation Mirror.
 * @dev Enforces Invariant 1 (Physical forward commodity exclusion), Invariant 4 (Zero customer workload data),
 *      and Invariant 6 (Off-chain PostgreSQL is legal source of truth; Arbitrum acts as settlement & proof mirror).
 */
contract VerinodeRegistry {
    // -------------------------------------------------------------------------
    // Custom Errors (solidity-security best practice)
    // -------------------------------------------------------------------------
    error Unauthorized();
    error ReentrancyGuard();
    error InvalidTradeId();
    error InvalidCounterparty();
    error InvalidTimeWindow();
    error InvalidDepositAmount();
    error TradeAlreadyExists();
    error TradeNotFound();
    error IllegalStateTransition(uint8 currentState, uint8 targetState);
    error BelowBenchmarkFloor(uint32 reportedGbps, uint32 requiredFloor);
    error EscrowTransferFailed();

    // -------------------------------------------------------------------------
    // State Enums (mirrors Verinode 18-state machine)
    // -------------------------------------------------------------------------
    enum TradeState {
        ContractPending, // 0
        FundedSecured,   // 1
        Scheduled,       // 2
        DeliveryTest,    // 3
        Live,            // 4
        Completed,       // 5
        Settled,         // 6
        Cancelled,       // 7
        FailedDelivery,  // 8
        Cure,            // 9
        Substituted,     // 10
        ClaimOpen,       // 11
        Disputed,        // 12
        Terminated       // 13
    }

    struct TradeEnvelope {
        bytes32 tradeId;             // Canonical UUID hash
        address buyer;               // Institutional buyer address
        address seller;              // Infrastructure supplier address
        bytes16 gradeId;             // e.g. "H100-SXM-8XNV"
        uint256 escrowAmount;        // Escrow balance in wei (or USDC token)
        uint64 windowStart;          // Unix epoch seconds
        uint64 windowEnd;            // Unix epoch seconds
        TradeState state;            // Current lifecycle state
        bytes32 attestationDigest;   // SHA-256 of verified telemetry report
        uint32 ncclGbps;             // Measured NCCL AllReduce bandwidth
        bool canaryPassed;           // Evaluator passing gate
        uint64 createdAt;
        uint64 updatedAt;
    }

    // -------------------------------------------------------------------------
    // Storage
    // -------------------------------------------------------------------------
    address public owner;
    address public settlementOperator;
    uint32 public constant NCCL_BENCHMARK_FLOOR = 400; // 400 GB/s floor for H100 SXM

    mapping(bytes32 => TradeEnvelope) public trades;
    mapping(bytes32 => bool) public tradeExists;

    uint256 private _status; // Reentrancy guard slot

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------
    event TradeInitialized(
        bytes32 indexed tradeId,
        address indexed buyer,
        address indexed seller,
        bytes16 gradeId,
        uint256 escrowAmount,
        uint64 windowStart,
        uint64 windowEnd
    );

    event AttestationRecorded(
        bytes32 indexed tradeId,
        bytes32 attestationDigest,
        uint32 ncclGbps,
        bool canaryPassed
    );

    event TradeStateAdvanced(
        bytes32 indexed tradeId,
        TradeState priorState,
        TradeState newState
    );

    event EscrowReleased(
        bytes32 indexed tradeId,
        address indexed recipient,
        uint256 amount
    );

    event EscrowRefunded(
        bytes32 indexed tradeId,
        address indexed buyer,
        uint256 amount
    );

    // -------------------------------------------------------------------------
    // Modifiers
    // -------------------------------------------------------------------------
    modifier onlyOwner() {
        if (msg.sender != owner) revert Unauthorized();
        _;
    }

    modifier onlyOperator() {
        if (msg.sender != settlementOperator && msg.sender != owner) revert Unauthorized();
        _;
    }

    modifier nonReentrant() {
        if (_status == 2) revert ReentrancyGuard();
        _status = 2;
        _;
        _status = 1;
    }

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------
    constructor(address _operator) {
        if (_operator == address(0)) revert InvalidCounterparty();
        owner = msg.sender;
        settlementOperator = _operator;
        _status = 1;
    }

    // -------------------------------------------------------------------------
    // Core Functions
    // -------------------------------------------------------------------------

    /**
     * @notice Initialize a new bilateral capacity forward reservation on Arbitrum.
     * @param tradeId Canonical trade UUID keyed from off-chain database.
     * @param buyer Address of the compute buyer.
     * @param seller Address of the compute provider.
     * @param gradeId Standardized hardware grade identifier (e.g. H100-SXM-8XNV).
     * @param windowStart Delivery window start epoch.
     * @param windowEnd Delivery window end epoch.
     */
    function initializeTrade(
        bytes32 tradeId,
        address buyer,
        address seller,
        bytes16 gradeId,
        uint64 windowStart,
        uint64 windowEnd
    ) external payable nonReentrant {
        if (tradeId == bytes32(0)) revert InvalidTradeId();
        if (buyer == address(0) || seller == address(0)) revert InvalidCounterparty();
        if (windowEnd <= windowStart) revert InvalidTimeWindow();
        if (tradeExists[tradeId]) revert TradeAlreadyExists();

        trades[tradeId] = TradeEnvelope({
            tradeId: tradeId,
            buyer: buyer,
            seller: seller,
            gradeId: gradeId,
            escrowAmount: msg.value,
            windowStart: windowStart,
            windowEnd: windowEnd,
            state: msg.value > 0 ? TradeState.FundedSecured : TradeState.ContractPending,
            attestationDigest: bytes32(0),
            ncclGbps: 0,
            canaryPassed: false,
            createdAt: uint64(block.timestamp),
            updatedAt: uint64(block.timestamp)
        });

        tradeExists[tradeId] = true;

        emit TradeInitialized(tradeId, buyer, seller, gradeId, msg.value, windowStart, windowEnd);
    }

    /**
     * @notice Fund escrow for an existing bilateral trade reservation.
     */
    function depositEscrow(bytes32 tradeId) external payable nonReentrant {
        if (!tradeExists[tradeId]) revert TradeNotFound();
        if (msg.value == 0) revert InvalidDepositAmount();

        TradeEnvelope storage env = trades[tradeId];
        env.escrowAmount += msg.value;

        if (env.state == TradeState.ContractPending) {
            env.state = TradeState.FundedSecured;
            emit TradeStateAdvanced(tradeId, TradeState.ContractPending, TradeState.FundedSecured);
        }

        env.updatedAt = uint64(block.timestamp);
    }

    /**
     * @notice Record a verified hardware attestation and canary test result.
     * @dev Validates the NCCL AllReduce benchmark floor (>= 400 GB/s).
     */
    function recordAttestation(
        bytes32 tradeId,
        bytes32 attestationDigest,
        uint32 ncclGbps,
        bool canaryPassed
    ) external onlyOperator {
        if (!tradeExists[tradeId]) revert TradeNotFound();

        // Enforce benchmark floor
        if (canaryPassed && ncclGbps < NCCL_BENCHMARK_FLOOR) {
            revert BelowBenchmarkFloor(ncclGbps, NCCL_BENCHMARK_FLOOR);
        }

        TradeEnvelope storage env = trades[tradeId];
        env.attestationDigest = attestationDigest;
        env.ncclGbps = ncclGbps;
        env.canaryPassed = canaryPassed;
        env.updatedAt = uint64(block.timestamp);

        emit AttestationRecorded(tradeId, attestationDigest, ncclGbps, canaryPassed);
    }

    /**
     * @notice Advance contract lifecycle state according to the Verinode 18-state transition rules.
     */
    function advanceTradeState(bytes32 tradeId, TradeState newState) external onlyOperator {
        if (!tradeExists[tradeId]) revert TradeNotFound();

        TradeEnvelope storage env = trades[tradeId];
        TradeState prior = env.state;

        if (!_isValidTransition(prior, newState)) {
            revert IllegalStateTransition(uint8(prior), uint8(newState));
        }

        env.state = newState;
        env.updatedAt = uint64(block.timestamp);

        emit TradeStateAdvanced(tradeId, prior, newState);
    }

    /**
     * @notice Release escrowed funds to the seller upon verified delivery completion and settlement.
     */
    function releaseEscrowToSeller(bytes32 tradeId) external onlyOperator nonReentrant {
        if (!tradeExists[tradeId]) revert TradeNotFound();

        TradeEnvelope storage env = trades[tradeId];
        if (env.state != TradeState.Completed && env.state != TradeState.Settled) {
            revert IllegalStateTransition(uint8(env.state), uint8(TradeState.Settled));
        }

        uint256 amount = env.escrowAmount;
        if (amount == 0) revert InvalidDepositAmount();

        env.escrowAmount = 0;
        env.state = TradeState.Settled;
        env.updatedAt = uint64(block.timestamp);

        (bool success, ) = env.seller.call{value: amount}("");
        if (!success) revert EscrowTransferFailed();

        emit EscrowReleased(tradeId, env.seller, amount);
    }

    /**
     * @notice Refund escrowed funds to the buyer upon failed delivery or cancellation.
     */
    function refundEscrowToBuyer(bytes32 tradeId) external onlyOperator nonReentrant {
        if (!tradeExists[tradeId]) revert TradeNotFound();

        TradeEnvelope storage env = trades[tradeId];
        if (env.state != TradeState.FailedDelivery && env.state != TradeState.Cancelled && env.state != TradeState.Terminated) {
            revert IllegalStateTransition(uint8(env.state), uint8(TradeState.Cancelled));
        }

        uint256 amount = env.escrowAmount;
        if (amount == 0) revert InvalidDepositAmount();

        env.escrowAmount = 0;
        env.updatedAt = uint64(block.timestamp);

        (bool success, ) = env.buyer.call{value: amount}("");
        if (!success) revert EscrowTransferFailed();

        emit EscrowRefunded(tradeId, env.buyer, amount);
    }

    // -------------------------------------------------------------------------
    // Internal State Machine Validator
    // -------------------------------------------------------------------------
    function _isValidTransition(TradeState current, TradeState next) internal pure returns (bool) {
        if (current == TradeState.ContractPending) {
            return next == TradeState.FundedSecured || next == TradeState.Cancelled;
        }
        if (current == TradeState.FundedSecured) {
            return next == TradeState.Scheduled || next == TradeState.Cancelled;
        }
        if (current == TradeState.Scheduled) {
            return next == TradeState.DeliveryTest;
        }
        if (current == TradeState.DeliveryTest) {
            return next == TradeState.Live || next == TradeState.FailedDelivery;
        }
        if (current == TradeState.Live) {
            return next == TradeState.Completed || next == TradeState.ClaimOpen;
        }
        if (current == TradeState.Completed) {
            return next == TradeState.Settled;
        }
        if (current == TradeState.ClaimOpen) {
            return next == TradeState.Disputed;
        }
        if (current == TradeState.Disputed) {
            return next == TradeState.Settled || next == TradeState.Terminated;
        }
        return false;
    }
}
