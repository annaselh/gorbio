import {
  Activity,
  Boxes,
  BriefcaseBusiness,
  Building2,
  Calculator,
  Calendar,
  CircleCheck,
  ClipboardList,
  Factory,
  FolderKanban,
  IdCard,
  LayoutDashboard,
  LifeBuoy,
  MessageSquare,
  Package,
  Plug,
  ScrollText,
  Settings,
  ShoppingBag,
  ShoppingCart,
  Users,
  Wallet,
  type LucideIcon,
} from "lucide-react";

/**
 * Name -> component. Menu manifests carry the *name* so that a manifest
 * delivered over the wire (federation / server-driven menu) stays plain JSON.
 */
export const ICONS: Record<string, LucideIcon> = {
  activity: Activity,
  accounting: Calculator,
  audit: ScrollText,
  boxes: Boxes,
  calendar: Calendar,
  company: Building2,
  customers: Users,
  dashboard: LayoutDashboard,
  discuss: MessageSquare,
  helpdesk: LifeBuoy,
  hr: IdCard,
  integrations: Plug,
  inventory: Package,
  manufacturing: Factory,
  preferences: Settings,
  projects: FolderKanban,
  purchase: ShoppingBag,
  sales: ShoppingCart,
  tasks: CircleCheck,
  users: BriefcaseBusiness,
  wallet: Wallet,
  list: ClipboardList,
};

export function resolveIcon(name?: string): LucideIcon {
  return (name && ICONS[name]) || ClipboardList;
}
