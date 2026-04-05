import { cn } from "@/lib/utils";

const roleColors = {
  admin: "bg-ash-600/40 text-ash-100 border-ash-500/50",
  manager: "bg-ash-700/40 text-ash-200 border-ash-600/40",
  viewer: "bg-ash-800/40 text-ash-300 border-ash-700/50",
  income: "bg-emerald-800/40 text-emerald-100 border-emerald-700/40",
  expense: "bg-rose-900/40 text-rose-100 border-rose-700/40",
};

export function Badge({ children, tone = "viewer", className }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-semibold capitalize",
        roleColors[tone] || roleColors.viewer,
        className
      )}
    >
      {children}
    </span>
  );
}
