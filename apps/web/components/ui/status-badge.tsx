import { ContractState } from "@/lib/types";
import { cn } from "@/lib/utils";

interface StatusBadgeProps {
  state: ContractState;
  className?: string;
}

export function StatusBadge({ state, className }: StatusBadgeProps) {
  let badgeStyles = "bg-zinc-800 text-zinc-300 border-zinc-700";
  let dotStyles = "bg-zinc-400";
  let label: string = state;

  switch (state) {
    case "draft_rfq":
    case "open":
      badgeStyles = "bg-blue-950/60 text-blue-300 border-blue-800/60";
      dotStyles = "bg-blue-400";
      label = state === "draft_rfq" ? "Draft RFQ" : "RFQ Open";
      break;
    case "quoted":
      badgeStyles = "bg-purple-950/60 text-purple-300 border-purple-800/60";
      dotStyles = "bg-purple-400";
      label = "Quoted";
      break;
    case "accepted":
    case "contract_pending":
      badgeStyles = "bg-amber-950/60 text-amber-300 border-amber-800/60";
      dotStyles = "bg-amber-400 animate-pulse";
      label = state === "accepted" ? "Accepted" : "Contract Pending Sign";
      break;
    case "funded_secured":
    case "scheduled":
      badgeStyles = "bg-cyan-950/60 text-cyan-300 border-cyan-800/60";
      dotStyles = "bg-cyan-400";
      label = state === "funded_secured" ? "Funded & Escrowed" : "Scheduled";
      break;
    case "delivery_test":
      badgeStyles = "bg-indigo-950/60 text-indigo-300 border-indigo-800/60";
      dotStyles = "bg-indigo-400";
      label = "Delivery Attestation / Canary";
      break;
    case "live":
      badgeStyles = "bg-emerald-950/70 text-emerald-300 border-emerald-800";
      dotStyles = "bg-emerald-400";
      label = "Live Delivered Capacity";
      break;
    case "completed":
    case "settled":
      badgeStyles = "bg-zinc-900 text-zinc-300 border-zinc-700";
      dotStyles = "bg-zinc-300";
      label = state === "completed" ? "Delivery Completed" : "Settled";
      break;
    case "failed_delivery":
    case "disputed":
    case "terminated":
      badgeStyles = "bg-rose-950/60 text-rose-300 border-rose-800/60";
      dotStyles = "bg-rose-400";
      label = state.replace("_", " ").toUpperCase();
      break;
    case "cure":
    case "substituted":
      badgeStyles = "bg-orange-950/60 text-orange-300 border-orange-800/60";
      dotStyles = "bg-orange-400";
      label = state === "cure" ? "In Cure Window" : "Substituted Node";
      break;
    case "claim_open":
      badgeStyles = "bg-amber-950/80 text-amber-200 border-amber-700";
      dotStyles = "bg-amber-400";
      label = "Claim Open";
      break;
    case "cancelled":
      badgeStyles = "bg-zinc-900 text-zinc-500 border-zinc-800";
      dotStyles = "bg-zinc-600";
      label = "Cancelled";
      break;
  }

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-mono font-medium border transition-colors",
        badgeStyles,
        className
      )}
    >
      <span className={cn("h-1.5 w-1.5 rounded-full", dotStyles)} />
      <span>{label}</span>
    </span>
  );
}
