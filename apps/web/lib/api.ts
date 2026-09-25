import { DeliveryGrade } from "./types";

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
  async getHealth(): Promise<GatewayHealth> {
    const res = await fetch(`${API_BASE}/healthz`, { cache: "no-store" });
    return handleResponse<GatewayHealth>(res);
  },

  async getGrades(): Promise<DeliveryGrade[]> {
    const res = await fetch(`${API_BASE}/v1/grades`, { cache: "no-store" });
    return handleResponse<DeliveryGrade[]>(res);
  },

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
