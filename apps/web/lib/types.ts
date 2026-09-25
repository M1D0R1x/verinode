export type ContractState =
  | "draft_rfq"
  | "open"
  | "quoted"
  | "accepted"
  | "contract_pending"
  | "funded_secured"
  | "scheduled"
  | "delivery_test"
  | "live"
  | "completed"
  | "settled"
  | "cancelled"
  | "failed_delivery"
  | "cure"
  | "substituted"
  | "claim_open"
  | "disputed"
  | "terminated";

export type DeliveryGrade = Grade;

export interface Grade {
  id: string; // e.g. "H100-SXM-8XNV"
  gpu_sku: string;
  min_memory_gb: number;
  topology: string;
  min_healthy_gpu_count: number;
  benchmark_floor: {
    nccl_allreduce_gb_per_sec: number;
    min_cuda_driver?: string;
    max_ecc_unrecovered_errors?: number;
  };
  min_cpu_cores: number;
  min_ram_gb: number;
  min_nvme_perf: number;
}

export interface RFQ {
  id: string;
  buyer_id: string;
  grade_id: string;
  region_bucket: string;
  window_start: string;
  window_end: string;
  status: "open" | "quoted" | "accepted" | "cancelled" | "expired";
  created_at: string;
  quotes?: Quote[];
}

export interface Quote {
  id: string;
  rfq_id: string;
  seller_id: string;
  seller_name?: string;
  block_id: string;
  price_cents: number;
  currency: string;
  expires_at: string;
  status: "active" | "accepted" | "expired" | "withdrawn";
  created_at: string;
}

export interface Contract {
  id: string; // canonical trade_id
  rfq_id: string;
  quote_id: string;
  buyer_id: string;
  seller_id: string;
  grade_id: string;
  template_version: string;
  state: ContractState;
  signed_pdf_hash?: string;
  signed_pdf_ref?: string;
  created_at: string;
  updated_at: string;
  grade?: Grade;
  events?: ContractEvent[];
}

export interface ContractEvent {
  id: string;
  contract_id: string;
  prior_state: ContractState;
  new_state: ContractState;
  actor: string;
  reason: string;
  idempotency_key: string;
  evidence_hash?: string;
  created_at: string;
}

export interface TelemetryHealth {
  agent_id: string;
  block_id: string;
  last_heartbeat_at: string;
  gpu_count: number;
  ecc_errors: number;
  allreduce_gbps: number;
  status: "healthy" | "degraded" | "failing";
}
