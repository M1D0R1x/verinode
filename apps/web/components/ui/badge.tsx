import * as React from "react";
import { cn } from "@/lib/utils";

export type BadgeVariant =
  | "default"
  | "secondary"
  | "destructive"
  | "outline"
  | "success"
  | "warning";

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: BadgeVariant;
}

const variantStyles: Record<BadgeVariant, string> = {
  default: "border-transparent bg-primary/20 text-primary hover:bg-primary/30",
  secondary: "border-transparent bg-zinc-800 text-zinc-200 hover:bg-zinc-700",
  destructive: "border-transparent bg-rose-500/20 text-rose-400 hover:bg-rose-500/30",
  outline: "text-zinc-300 border-border bg-transparent",
  success: "border-transparent bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30",
  warning: "border-transparent bg-amber-500/20 text-amber-400 hover:bg-amber-500/30",
};

export function Badge({ className, variant = "default", ...props }: BadgeProps) {
  return (
    <div
      className={cn(
        "inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none",
        variantStyles[variant],
        className
      )}
      {...props}
    />
  );
}
