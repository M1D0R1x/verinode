import { Contract, ContractState, DeliveryGrade, Quote, RFQ } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface GatewayHealth {
  status: string;
  service: string;
  version: string;
  timestamp: string;
}

export interface TransitionValidationRequest {
  current_state: string;
  next_state: string;
  actor: string;
  reason: string;
  idempotency_key: string;
  authorization_decision?: Record<string, unknown>;
}

export interface TransitionValidationResponse {
  valid: boolean;
  current_state: string;
  next_state: string;
  timestamp: string;
}

export interface TelemetryVerifyRequest {
  report: {
    agent_id: string;
    block_id: string;
    timestamp: string;
    gpu_sku: string;
    gpu_count: number;
    memory_gb: number;
    topology: string;
    driver_version: string;
    cuda_version: string;
    pcie_link_gen: number;
    pcie_link_width: number;
    nvlink_active: boolean;
    ecc_unrecovered_errors: number;
    thermal_throttling_count: number;
    nccl_allreduce_gb_per_sec: number;
  };
  signature: string;
  public_key: string;
}

export interface TelemetryVerifyResponse {
  verified: boolean;
  agent_id: string;
  block_id: string;
  grade_compliant: boolean;
  timestamp: string;
}

export interface Participant {
  id?: string;
  legal_name: string;
  jurisdiction: string;
  role: "buyer" | "seller" | "both";
  kyc_status?: string;
  credit_limit_cents?: number;
}

export interface InventoryBlock {
  id?: string;
  seller_id: string;
  grade_id: string;
  region_bucket: string;
  facility_ref?: string;
  window_start: string;
  window_end: string;
  status?: string;
}

export class VerinodeApiError extends Error {
  constructor(
    public status: number,
    public title: string,
    public detail: string
  ) {
    super(`[${status}] ${title}: ${detail}`);
    this.name = "VerinodeApiError";
  }
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errorJson: { title?: string; detail?: string } = {};
    try {
      errorJson = await res.json();
    } catch {
      // ignore
    }
    throw new VerinodeApiError(
      res.status,
      errorJson.title || res.statusText,
      errorJson.detail || "An unexpected error occurred contacting API gateway"
    );
  }
  return res.json() as Promise<T>;
}

export const api = {
  // System Health & Grades
  async getHealth(): Promise<GatewayHealth> {
    const res = await fetch(`${API_BASE}/healthz`, { cache: "no-store" });
    return handleResponse<GatewayHealth>(res);
  },

  async getGrades(): Promise<DeliveryGrade[]> {
    const res = await fetch(`${API_BASE}/v1/grades`, { cache: "no-store" });
    return handleResponse<DeliveryGrade[]>(res);
  },

  // Participants
  async createParticipant(p: Participant): Promise<Participant> {
    const res = await fetch(`${API_BASE}/v1/participants`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(p),
    });
    return handleResponse<Participant>(res);
  },

  async listParticipants(): Promise<Participant[]> {
    const res = await fetch(`${API_BASE}/v1/participants`, { cache: "no-store" });
    return handleResponse<Participant[]>(res);
  },

  // Inventory Blocks
  async createInventoryBlock(b: InventoryBlock): Promise<InventoryBlock> {
    const res = await fetch(`${API_BASE}/v1/inventory/blocks`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(b),
    });
    return handleResponse<InventoryBlock>(res);
  },

  async listInventoryBlocks(params?: { grade_id?: string; region_bucket?: string }): Promise<InventoryBlock[]> {
    const searchParams = new URLSearchParams();
    if (params?.grade_id) searchParams.set("grade_id", params.grade_id);
    if (params?.region_bucket) searchParams.set("region_bucket", params.region_bucket);

    const res = await fetch(`${API_BASE}/v1/inventory/blocks?${searchParams.toString()}`, {
      cache: "no-store",
    });
    return handleResponse<InventoryBlock[]>(res);
  },

  // RFQ
  async createRFQ(req: {
    buyer_id: string;
    grade_id: string;
    region_bucket: string;
    window_start: string;
    window_end: string;
  }): Promise<RFQ> {
    const res = await fetch(`${API_BASE}/v1/rfqs`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });
    return handleResponse<RFQ>(res);
  },

  async getRFQ(id: string): Promise<RFQ> {
    const res = await fetch(`${API_BASE}/v1/rfqs/${id}`, { cache: "no-store" });
    return handleResponse<RFQ>(res);
  },

  // Quotes
  async createQuote(
    rfqId: string,
    q: {
      seller_id: string;
      block_id: string;
      price_cents: number;
      currency?: string;
      expires_at: string;
    }
  ): Promise<Quote> {
    const res = await fetch(`${API_BASE}/v1/rfqs/${rfqId}/quotes`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(q),
    });
    return handleResponse<Quote>(res);
  },

  async listQuotes(rfqId: string): Promise<Quote[]> {
    const res = await fetch(`${API_BASE}/v1/rfqs/${rfqId}/quotes`, { cache: "no-store" });
    return handleResponse<Quote[]>(res);
  },

  async acceptQuote(
    rfqId: string,
    quoteId: string,
    buyerId: string
  ): Promise<{ status: string; contract_id: string; contract: Contract }> {
    const res = await fetch(`${API_BASE}/v1/rfqs/${rfqId}/quotes/${quoteId}/accept`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ buyer_id: buyerId }),
    });
    return handleResponse<{ status: string; contract_id: string; contract: Contract }>(res);
  },

  // Contracts
  async getContract(tradeId: string): Promise<Contract> {
    const res = await fetch(`${API_BASE}/v1/contracts/${tradeId}`, { cache: "no-store" });
    return handleResponse<Contract>(res);
  },

  async advanceContract(
    tradeId: string,
    nextState: ContractState,
    actor: string,
    reason: string,
    idempotencyKey: string
  ): Promise<{ trade_id: string; next_state: string; updated_at: string }> {
    const res = await fetch(`${API_BASE}/v1/contracts/${tradeId}/advance`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        next_state: nextState,
        actor,
        reason,
        idempotency_key: idempotencyKey,
      }),
    });
    return handleResponse<{ trade_id: string; next_state: string; updated_at: string }>(res);
  },

  // State Transition Validator
  async validateTransition(
    req: TransitionValidationRequest
  ): Promise<TransitionValidationResponse> {
    const res = await fetch(`${API_BASE}/v1/contracts/validate-transition`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });
    return handleResponse<TransitionValidationResponse>(res);
  },

  // Telemetry Attestation Verification
  async verifyTelemetry(
    req: TelemetryVerifyRequest
  ): Promise<TelemetryVerifyResponse> {
    const res = await fetch(`${API_BASE}/v1/telemetry/attestations/verify`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });
    return handleResponse<TelemetryVerifyResponse>(res);
  },
};
