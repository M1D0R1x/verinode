import { ContractState } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Check, ShieldAlert, Clock, Play } from "lucide-react";

interface StateTimelineProps {
  currentState: ContractState;
  className?: string;
}

const happyPathStages: { state: ContractState; label: string }[] = [
  { state: "open", label: "RFQ Open" },
  { state: "quoted", label: "Quoted" },
  { state: "accepted", label: "Accepted" },
  { state: "contract_pending", label: "Sign Pending" },
  { state: "funded_secured", label: "Funded & Escrow" },
  { state: "scheduled", label: "Scheduled" },
  { state: "delivery_test", label: "Canary Attestation" },
  { state: "live", label: "Live Delivery" },
  { state: "completed", label: "Completed" },
  { state: "settled", label: "Settled" },
];

export function StateTimeline({ currentState, className }: StateTimelineProps) {
  const isException = [
    "failed_delivery",
    "cure",
    "substituted",
    "claim_open",
    "disputed",
    "terminated",
    "cancelled",
  ].includes(currentState);

  // Find index in happy path
  const currentIndex = happyPathStages.findIndex((s) => s.state === currentState);

  return (
    <div className={cn("w-full py-4", className)}>
      {isException && (
        <div className="mb-4 flex items-center gap-2 p-3 rounded-lg border border-amber-800/80 bg-amber-950/40 text-amber-200 text-xs">
          <ShieldAlert className="h-4 w-4 text-amber-400 shrink-0" />
          <span>
            Contract is currently in exception workflow: <strong className="font-mono uppercase">{currentState.replace("_", " ")}</strong>. Active remedy or arbitration proceedings apply.
          </span>
        </div>
      )}

      <div className="relative">
        {/* Progress bar background */}
        <div className="absolute top-4 left-0 w-full h-0.5 bg-border -translate-y-1/2 z-0" />

        <div className="relative z-10 flex items-start justify-between overflow-x-auto pb-4 gap-2">
          {happyPathStages.map((stage, idx) => {
            const isCompleted = currentIndex > idx;
            const isCurrent = currentState === stage.state;
            const isPending = currentIndex < idx;

            return (
              <div key={stage.state} className="flex flex-col items-center min-w-[72px] sm:min-w-[90px] text-center">
                <div
                  className={cn(
                    "flex h-8 w-8 items-center justify-center rounded-full border text-xs font-mono transition-all duration-300",
                    isCompleted && "bg-accent border-accent text-white shadow-sm",
                    isCurrent && "bg-primary border-primary text-white ring-4 ring-primary/20",
                    isPending && "bg-surface border-border text-muted"
                  )}
                >
                  {isCompleted ? (
                    <Check className="h-4 w-4" />
                  ) : isCurrent ? (
                    <Play className="h-3.5 w-3.5 fill-current" />
                  ) : (
                    <Clock className="h-3.5 w-3.5 text-zinc-500" />
                  )}
                </div>

                <span
                  className={cn(
                    "mt-2 text-[11px] font-medium leading-tight",
                    isCurrent ? "text-white font-semibold" : isCompleted ? "text-zinc-300" : "text-muted"
                  )}
                >
                  {stage.label}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
