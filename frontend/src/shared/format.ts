const idNumber = new Intl.NumberFormat("id-ID", { maximumFractionDigits: 0 });

/** "Rp 1.250.000.000" — id-ID uses "." as the thousands separator. */
export function formatIDR(value: number): string {
  return `Rp ${idNumber.format(value)}`;
}

export function formatNumber(value: number): string {
  return idNumber.format(value);
}

/** Compact axis / hero form: "Rp 200M", "Rp 1,25B". */
export function formatIDRCompact(value: number): string {
  const abs = Math.abs(value);
  if (abs >= 1_000_000_000) return `Rp ${trim(value / 1_000_000_000)}B`;
  if (abs >= 1_000_000) return `Rp ${trim(value / 1_000_000)}M`;
  if (abs >= 1_000) return `Rp ${trim(value / 1_000)}K`;
  return `Rp ${idNumber.format(value)}`;
}

function trim(n: number): string {
  // Two significant decimals at most, and no trailing ",00".
  const rounded = Math.round(n * 100) / 100;
  return new Intl.NumberFormat("id-ID", { maximumFractionDigits: 2 }).format(
    rounded,
  );
}

export function formatPercent(value: number, digits = 1): string {
  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(digits)}%`;
}

const dateFmt = new Intl.DateTimeFormat("en-GB", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

export function formatDate(iso: string): string {
  return dateFmt.format(new Date(iso));
}

const dayMonthFmt = new Intl.DateTimeFormat("en-GB", {
  day: "numeric",
  month: "short",
});

export function formatDayMonth(iso: string): string {
  return dayMonthFmt.format(new Date(iso));
}

/** "2m ago" / "3h ago" / "5d ago" — relative to now. */
export function formatRelative(iso: string, now = Date.now()): string {
  const diffMs = now - new Date(iso).getTime();
  const mins = Math.round(diffMs / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.round(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.round(hours / 24);
  return `${days}d ago`;
}
