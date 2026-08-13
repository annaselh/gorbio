export interface StockAlertItem {
  id: string;
  name: string;
  qty: number;
  emoji: string;
}

export const stockAlerts: StockAlertItem[] = [
  { id: "s1", name: "iPhone 15 Pro Max", qty: 12, emoji: "📱" },
  { id: "s2", name: "Samsung Galaxy S24", qty: 18, emoji: "📱" },
  { id: "s3", name: "AirPods Pro", qty: 0, emoji: "🎧" },
  { id: "s4", name: "MacBook Air M3", qty: 8, emoji: "💻" },
  { id: "s5", name: "Logitech MX Master 3S", qty: 15, emoji: "🖱️" },
];
