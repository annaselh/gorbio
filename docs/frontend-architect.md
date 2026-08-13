# Frontend Architecture — Modular ERP Platform

> Arsitektur frontend modular berbasis React untuk platform ERP yang dapat dikustomisasi pihak ketiga

| | |
|---|---|
| **Versi Dokumen** | 1.0 (Draft) |
| **Status** | For Review |
| **Tanggal** | Agustus 2026 |
| **Referensi** | PRD v1.0, TSD v1.0, ERD v1.0 |
| **Stack Utama** | React + Vite + TypeScript |

---

## Daftar Isi

1. [Keputusan: React vs Go](#1-keputusan-react-vs-go)
2. [Prinsip Arsitektur](#2-prinsip-arsitektur)
3. [Tumpukan Teknologi](#3-tumpukan-teknologi)
4. [Arsitektur Shell + Modul](#4-arsitektur-shell--modul)
5. [Strategi 1 — Plugin Registry (Satu Codebase)](#5-strategi-1--plugin-registry-satu-codebase)
6. [Strategi 2 — Module Federation (Runtime)](#6-strategi-2--module-federation-runtime)
7. [Strategi 3 — Server-Driven UI (Custom Field)](#7-strategi-3--server-driven-ui-custom-field)
8. [Struktur Direktori](#8-struktur-direktori)
9. [Kontrak Modul Frontend](#9-kontrak-modul-frontend)
10. [Keselarasan Frontend–Backend](#10-keselarasan-frontendbackend)
11. [Roadmap Implementasi](#11-roadmap-implementasi)

---

## 1. Keputusan: React vs Go

Untuk ERP modular ini, **React (atau framework JS sejenis) adalah pilihan yang lebih tepat daripada Go untuk frontend.**

| Aspek | Go (WASM / html/template) | React / JS |
|---|---|---|
| Ekosistem komponen UI | Terbatas | Sangat matang (tabel, form, dsb) |
| Interaktivitas kompleks | Melawan arus | Natural |
| State management | Minim | TanStack Query, Redux, Zustand |
| Modularitas UI runtime | Sangat sulit | Module Federation, dynamic import |
| Server-driven / custom field | Sulit | Idiomatik |
| Tooling & hiring | Sempit untuk frontend | Luas |

Faktor penentu utama: kebutuhan modularitas. Karena backend memungkinkan modul pihak ketiga menambah fitur, frontend juga harus bisa "menerima" UI dari modul secara dinamis — dan ini jauh lebih mudah di ekosistem React/JS.

> Vue dengan pola serupa juga valid bila tim lebih nyaman. Semua konsep di dokumen ini (registry, federation, server-driven UI) berlaku sama.

---

## 2. Prinsip Arsitektur

Frontend mengikuti prinsip yang sama dengan backend agar selaras:

- **Shell stabil, modul dinamis** — sebuah host app menyediakan layout, routing, autentikasi, dan registry; modul mendaftarkan diri.
- **Modul tidak saling import langsung** — komunikasi melalui registry, slot, dan event, bukan dependency keras antar-modul.
- **Kontrak berbasis manifest** — setiap modul mengekspor manifest (route, menu, slot).
- **Dependency resolution deterministik** — modul diurutkan via topological sort (sama seperti backend).
- **Server-driven untuk data dinamis** — custom field dirender dari skema yang dikirim backend.

---

## 3. Tumpukan Teknologi

| Kategori | Pilihan | Alasan |
|---|---|---|
| Framework | React 18+ | Ekosistem, modularitas runtime |
| Build tool | Vite | Cepat, mendukung federation |
| Bahasa | TypeScript | Type-safety pada kontrak modul |
| Data fetching | TanStack Query | Cache, sinkronisasi ke REST API |
| Routing | TanStack Router / React Router | Route dinamis dari registry |
| Component library | Ant Design / Mantine / shadcn/ui | Komponen tabel & form kaya untuk ERP |
| State ringan | Zustand | State lintas modul yang sederhana |
| Federation (fase lanjut) | @originjs/vite-plugin-federation | Modul pihak ketiga tanpa rebuild |

> Ant Design sangat cocok untuk ERP karena komponen `Table`, `Form`, dan `ProComponents`-nya matang untuk data-entry berat.

---

## 4. Arsitektur Shell + Modul

```
┌─────────────────────────────────────────────────────────┐
│  Shell (host React app)                                  │
│  Layout · Routing · Auth · Module Registry               │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐  │
│  │ Module       │ │ Router +     │ │ Shared UI +      │  │
│  │ registry     │ │ menu builder │ │ API client       │  │
│  └──────▲───────┘ └──────▲───────┘ └────────▲─────────┘  │
└─────────┼────────────────┼──────────────────┼───────────┘
          │ register       │ register         │ use
   ┌──────┴─────┐   ┌───────┴──────┐   ┌───────┴────────────┐
   │ sales      │   │ inventory    │   │ modul pihak ketiga │
   │ route/menu │   │ route/menu   │   │ (load dinamis)     │
   └────────────┘   └──────────────┘   └────────────────────┘
```

Shell mengumpulkan manifest semua modul, mengurutkannya berdasarkan dependency, lalu merakit router dan menu secara dinamis. Modul pihak ketiga dapat di-load saat runtime (lihat Strategi 2).

---

## 5. Strategi 1 — Plugin Registry (Satu Codebase)

**Direkomendasikan untuk memulai.** Semua modul dalam satu build React, terstruktur sebagai plugin yang mendaftarkan diri ke registry. Setara dengan monolith modular di backend.

### 5.1 Manifest Modul

```typescript
// modules/sales/index.tsx
import type { ModuleManifest } from "@/core/types";
import { lazy } from "react";

const SalesOrders = lazy(() => import("./pages/SalesOrders"));

export const salesModule: ModuleManifest = {
  name: "sales",
  dependencies: ["base", "inventory"],
  routes: [
    { path: "/sales/orders", element: <SalesOrders /> },
  ],
  menu: [
    { label: "Penjualan", icon: "shopping-cart", path: "/sales/orders", sequence: 20 },
  ],
  // opsional: komponen yang bisa "disuntik" modul lain (mirip _inherit backend)
  slots: {
    "partner.form.extra": lazy(() => import("./slots/PartnerCreditField")),
  },
};
```

### 5.2 Tipe Kontrak

```typescript
// core/types.ts
import type { ReactNode, ComponentType } from "react";

export interface RouteDef {
  path: string;
  element: ReactNode;
}

export interface MenuDef {
  label: string;
  icon?: string;
  path: string;
  sequence: number;
  parent?: string;
}

export interface ModuleManifest {
  name: string;
  dependencies?: string[];
  routes?: RouteDef[];
  menu?: MenuDef[];
  slots?: Record<string, ComponentType<any>>;
}
```

### 5.3 Registry + Dependency Resolution

```typescript
// core/registry.ts
import type { ModuleManifest } from "./types";

export function resolveDependencyOrder(mods: ModuleManifest[]): ModuleManifest[] {
  const byName = new Map(mods.map((m) => [m.name, m]));
  const visited = new Set<string>();
  const ordered: ModuleManifest[] = [];

  function visit(m: ModuleManifest, stack: Set<string>) {
    if (visited.has(m.name)) return;
    if (stack.has(m.name)) throw new Error(`Cyclic dependency: ${m.name}`);
    stack.add(m.name);
    for (const dep of m.dependencies ?? []) {
      const depMod = byName.get(dep);
      if (!depMod) throw new Error(`Missing dependency: ${dep} (butuh oleh ${m.name})`);
      visit(depMod, stack);
    }
    stack.delete(m.name);
    visited.add(m.name);
    ordered.push(m);
  }

  for (const m of mods) visit(m, new Set());
  return ordered;
}

export function buildRegistry(mods: ModuleManifest[]) {
  const ordered = resolveDependencyOrder(mods);
  const routes = ordered.flatMap((m) => m.routes ?? []);
  const menuItems = ordered
    .flatMap((m) => m.menu ?? [])
    .sort((a, b) => a.sequence - b.sequence);

  // Slot registry: banyak modul bisa mengisi slot yang sama
  const slots: Record<string, React.ComponentType<any>[]> = {};
  for (const m of ordered) {
    for (const [slotName, Comp] of Object.entries(m.slots ?? {})) {
      (slots[slotName] ??= []).push(Comp);
    }
  }
  return { ordered, routes, menuItems, slots };
}
```

### 5.4 Daftarkan & Rakit Shell

```typescript
// core/modules.ts
import { baseModule } from "@/modules/base";
import { inventoryModule } from "@/modules/inventory";
import { salesModule } from "@/modules/sales";
import { buildRegistry } from "./registry";

export const registry = buildRegistry([
  baseModule,
  inventoryModule,
  salesModule,
]);
```

```tsx
// App.tsx
import { Suspense } from "react";
import { registry } from "@/core/modules";
import { Shell } from "@/core/Shell";
// Router dari TanStack/React Router dibangun dari registry.routes

export default function App() {
  return (
    <Shell menu={registry.menuItems}>
      <Suspense fallback={<div>Memuat…</div>}>
        <AppRouter routes={registry.routes} />
      </Suspense>
    </Shell>
  );
}
```

### 5.5 Slot: Extension UI Antar-Modul

Slot adalah cerminan pola `_inherit` di backend — modul lain menyisipkan UI ke titik yang sudah ditentukan modul pemilik.

```tsx
// core/Slot.tsx — komponen pemanggil slot
import { registry } from "@/core/modules";

export function Slot({ name, ctx }: { name: string; ctx?: any }) {
  const comps = registry.slots[name] ?? [];
  return (
    <>
      {comps.map((C, i) => <C key={i} ctx={ctx} />)}
    </>
  );
}
```

```tsx
// modules/base/pages/PartnerForm.tsx — modul base menyediakan slot
import { Slot } from "@/core/Slot";

export function PartnerForm({ partner }) {
  return (
    <form>
      <input name="name" defaultValue={partner.name} />
      <input name="email" defaultValue={partner.email} />
      {/* Modul sales menyisipkan field credit_limit di sini */}
      <Slot name="partner.form.extra" ctx={{ partner }} />
    </form>
  );
}
```

**Kelebihan:** sederhana, type-safe penuh, satu build, mudah di-debug.
**Kekurangan:** menambah modul butuh rebuild frontend (cocok bila target pengguna adalah developer).

---

## 6. Strategi 2 — Module Federation (Runtime)

**Untuk modul pihak ketiga tanpa rebuild shell.** Setara dengan gRPC/WASM di backend: modul di-build & di-deploy terpisah, di-load saat runtime.

### 6.1 Konfigurasi Host

```typescript
// vite.config.ts (host)
import federation from "@originjs/vite-plugin-federation";

export default {
  plugins: [
    federation({
      name: "host",
      remotes: {
        // URL bisa dibaca dari konfigurasi/DB saat runtime, tidak harus hardcode
        sales: "https://cdn.example.com/modules/sales/remoteEntry.js",
      },
      shared: ["react", "react-dom"], // dependency bersama tidak diduplikasi
    }),
  ],
};
```

### 6.2 Konfigurasi Remote (Modul)

```typescript
// vite.config.ts (remote: modul sales)
import federation from "@originjs/vite-plugin-federation";

export default {
  plugins: [
    federation({
      name: "sales",
      filename: "remoteEntry.js",
      exposes: {
        "./manifest": "./src/index.tsx", // ekspor manifest modul
      },
      shared: ["react", "react-dom"],
    }),
  ],
};
```

### 6.3 Load Dinamis di Host

```typescript
// core/loadRemote.ts
export async function loadRemoteModule(remoteName: string) {
  // @ts-ignore — resolusi remote ditangani plugin federation
  const mod = await import(`${remoteName}/manifest`);
  return mod.default as ModuleManifest;
}

// Saat boot: baca daftar modul aktif dari backend, lalu load
const active = await fetch("/api/v1/modules/frontend").then((r) => r.json());
for (const remote of active) {
  const manifest = await loadRemoteModule(remote.name);
  registry.register(manifest);
}
```

**Kelebihan:** modul pihak ketiga dipasang/diperbarui tanpa menyentuh shell.
**Kekurangan:** kompleksitas build & deploy naik; harus disiplin mengelola versi dependency bersama. Jangan dipakai sampai kebutuhan marketplace benar-benar ada.

---

## 7. Strategi 3 — Server-Driven UI (Custom Field)

**Pelengkap penting, bukan pengganti.** Backend punya custom field runtime (JSONB) yang bisa ditambah admin tanpa coding. Frontend harus merender field itu tanpa tahu sebelumnya: backend mengirim skema, frontend merender form dari skema.

### 7.1 Skema dari Backend

```jsonc
// GET /api/v1/partner/schema
{
  "fields": [
    { "name": "name",         "type": "string", "required": true },
    { "name": "credit_limit", "type": "float",  "source": "sales" },
    { "name": "loyalty_tier", "type": "string", "custom": true }   // ditambah admin
  ]
}
```

### 7.2 Renderer Generik

```tsx
// core/DynamicForm.tsx
interface FieldSchema {
  name: string;
  type: "string" | "int" | "float" | "bool" | "date";
  required?: boolean;
  custom?: boolean;
}

const FIELD_COMPONENTS = {
  string: (f: FieldSchema) => <input type="text" name={f.name} required={f.required} />,
  int:    (f: FieldSchema) => <input type="number" step="1" name={f.name} />,
  float:  (f: FieldSchema) => <input type="number" step="any" name={f.name} />,
  bool:   (f: FieldSchema) => <input type="checkbox" name={f.name} />,
  date:   (f: FieldSchema) => <input type="date" name={f.name} />,
};

export function DynamicForm({ schema }: { schema: { fields: FieldSchema[] } }) {
  return (
    <form>
      {schema.fields.map((f) => (
        <div key={f.name} className="field">
          <label>{f.name}</label>
          {FIELD_COMPONENTS[f.type](f)}
        </div>
      ))}
    </form>
  );
}
```

Saat admin menambah custom field di backend, form otomatis menampilkannya tanpa perubahan kode frontend. Inilah yang membuat ERP terasa dapat dikustomisasi oleh end-user, bukan hanya developer.

---

## 8. Struktur Direktori

```
erp-frontend/
├── src/
│   ├── core/                  # SHELL — tidak tahu logika bisnis modul
│   │   ├── types.ts           # kontrak ModuleManifest, RouteDef, MenuDef
│   │   ├── registry.ts        # resolveDependencyOrder + buildRegistry
│   │   ├── modules.ts         # daftar modul aktif (satu-satunya yang tahu daftar)
│   │   ├── Shell.tsx          # layout: sidebar, header, konten
│   │   ├── Slot.tsx           # penyisipan UI antar-modul
│   │   ├── DynamicForm.tsx    # server-driven form
│   │   ├── apiClient.ts       # wrapper fetch + auth ke REST API backend
│   │   └── AppRouter.tsx      # router dibangun dari registry.routes
│   │
│   ├── modules/               # BISNIS — boleh pakai core, tidak saling import
│   │   ├── base/
│   │   │   ├── index.tsx      # manifest
│   │   │   ├── pages/
│   │   │   └── slots/
│   │   ├── inventory/
│   │   └── sales/
│   │       ├── index.tsx
│   │       ├── pages/SalesOrders.tsx
│   │       └── slots/PartnerCreditField.tsx
│   │
│   ├── shared/                # komponen & util lintas modul
│   └── App.tsx
├── vite.config.ts
├── tsconfig.json
└── package.json
```

Aturan dependency (cermin backend):

```
App.tsx  →  modules  →  core
                │
                └── modul TIDAK saling import langsung;
                    komunikasi lewat registry, Slot, dan event
```

---

## 9. Kontrak Modul Frontend

Ringkasan apa yang boleh didaftarkan sebuah modul ke shell:

| Elemen | Wajib | Keterangan |
|---|---|---|
| `name` | Ya | Nama unik modul (selaras dengan backend). |
| `dependencies` | Tidak | Modul lain yang harus di-boot lebih dulu. |
| `routes` | Tidak | Halaman yang ditambahkan ke router. |
| `menu` | Tidak | Item menu (dengan `sequence` untuk urutan). |
| `slots` | Tidak | Komponen yang mengisi titik ekstensi modul lain. |

Kapabilitas bersifat opsional — modul cukup mendaftarkan yang relevan, mirip interface opsional di backend Go.

---

## 10. Keselarasan Frontend–Backend

Usahakan kontrak modul frontend mencerminkan kontrak modul backend:

| Backend (Go) | Frontend (React) |
|---|---|
| `Module` interface + manifest | `ModuleManifest` |
| Dependency resolution (topo-sort) | `resolveDependencyOrder` |
| `Extend` model (_inherit) | `Slot` (penyisipan UI) |
| Event bus antar-modul | Event/emitter atau state store lintas modul |
| Custom field runtime (JSONB) | Server-driven `DynamicForm` |
| Modul pihak ketiga (gRPC/WASM) | Module Federation |

Idealnya satu tim memiliki modul `sales` secara utuh: dari tabel database, model, event, hingga UI. Keselarasan ini membuat pengembangan per-tim rapi dan mengurangi koordinasi lintas-tim.

---

## 11. Roadmap Implementasi

| Fase | Fokus | Keluaran |
|---|---|---|
| **FE-1** | Shell + registry | Layout, routing dinamis, dependency resolution |
| **FE-2** | Modul inti | base, inventory, sales (Strategi 1) |
| **FE-3** | Server-driven UI | DynamicForm untuk custom field (Strategi 3) |
| **FE-4** | Slot & extension | Penyisipan UI antar-modul |
| **FE-5** | Module Federation | Modul pihak ketiga tanpa rebuild (Strategi 2) |

**Urutan yang disarankan:** mulai FE-1 dan FE-2 (Strategi 1) untuk 90% manfaat modularitas dengan kompleksitas terendah. Tambahkan FE-3 (server-driven UI) lebih awal karena custom field adalah inti nilai jual ERP. Tunda FE-5 (Module Federation) sampai kebutuhan marketplace pihak ketiga benar-benar muncul.