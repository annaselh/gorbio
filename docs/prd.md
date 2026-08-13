# Product Requirements Document — Modular ERP Platform

Platform ERP modular yang dapat dikustomisasi pihak ketiga (Go)

| | |
|---|---|
| **Versi Dokumen** | 1.0 (Draft) |
| **Status** | For Review |
| **Tanggal** | Agustus 2026 |
| **Pemilik Produk** | ________________ |
| **Klasifikasi** | Internal |

---

## 1. Ringkasan Eksekutif

Dokumen ini mendefinisikan kebutuhan produk untuk sebuah platform ERP modular berbasis Go yang memungkinkan pihak ketiga membangun, memasang, dan melepas modul fungsional tanpa mengubah inti sistem. Konsepnya serupa dengan Odoo, namun dibangun di atas ekosistem Go untuk memperoleh performa, kemudahan deployment (binary tunggal), dan konkurensi yang kuat.

Nilai inti produk terletak pada arsitektur modular: sebuah core yang stabil menyediakan kontrak (interface), sementara modul-modul bisnis mengimplementasikan kontrak tersebut dan didaftarkan secara dinamis. Model ini memungkinkan pengembangan paralel oleh banyak tim, ekosistem modul pihak ketiga, serta kustomisasi mendalam per pelanggan tanpa fork kode inti.

## 2. Latar Belakang & Masalah

### 2.1 Pernyataan Masalah

Banyak organisasi membutuhkan sistem ERP yang dapat disesuaikan dengan proses bisnis mereka yang unik. Solusi ERP komersial umumnya mahal, sulit dikustomisasi, dan mengunci pelanggan pada vendor. Solusi open-source yang ada (seperti Odoo) fleksibel namun terikat pada ekosistem Python yang memiliki keterbatasan performa dan kompleksitas deployment untuk skala tertentu.

### 2.2 Peluang

Terdapat peluang membangun platform ERP dengan karakteristik berikut:

- Performa tinggi dan footprint rendah berkat Go (binary tunggal, konkurensi native).
- Arsitektur modular yang memungkinkan ekosistem modul pihak ketiga.
- Kemudahan deployment dan operasional dibandingkan stack berbasis interpreter.
- Kustomisasi tanpa perlu fork kode inti (extension antar-modul dan custom field runtime).

## 3. Tujuan & Metrik Keberhasilan

### 3.1 Tujuan Produk

1. Menyediakan core ERP yang stabil dengan mekanisme modul yang jelas dan terdokumentasi.
2. Memungkinkan pihak ketiga membangun modul dengan effort minimal melalui SDK dan scaffolding.
3. Mendukung multi-tenant dan multi-company sejak awal.
4. Menyediakan jalur integrasi standar (REST, gRPC, webhook) untuk sistem eksternal.

### 3.2 Metrik Keberhasilan (KPI)

| Metrik | Target | Periode |
|---|---|---|
| Waktu membangun modul sederhana (developer baru) | < 1 hari | Setelah rilis SDK |
| Jumlah modul inti tersedia saat GA | ≥ 4 modul | General Availability |
| Latensi P95 request CRUD | < 200 ms | Produksi |
| Waktu boot sistem dengan 10 modul | < 5 detik | Produksi |
| Modul pihak ketiga terpasang tanpa recompile core | Didukung | Fase 2 |

## 4. Target Pengguna & Persona

| Persona | Deskripsi | Kebutuhan Utama |
|---|---|---|
| Admin Sistem | Mengelola instalasi, modul, dan konfigurasi tenant. | Install/uninstall modul, kelola user & hak akses. |
| Developer Modul | Membangun modul bisnis di atas core. | SDK jelas, dokumentasi, scaffolding, isolasi. |
| End User Bisnis | Menggunakan fitur ERP sehari-hari. | UI konsisten, cepat, sesuai proses kerja. |
| Integrator | Menghubungkan ERP dengan sistem lain. | API standar, webhook, autentikasi mesin. |

## 5. Lingkup Produk

### 5.1 Dalam Lingkup (In Scope)

- Core framework: module registry, dependency resolution, boot loader.
- ORM dengan dukungan extension antar-modul dan custom field runtime (JSONB).
- Sistem autentikasi, otorisasi (RBAC/ACL), dan multi-tenant.
- Event bus internal untuk komunikasi antar-modul (loose coupling).
- Migrasi database versioned per-modul.
- Modul inti: base, inventory, sales, accounting.
- Lapisan integrasi: REST API otomatis, webhook keluar, connector async.
- Mekanisme modul pihak ketiga via gRPC dan WASM.

### 5.2 Di Luar Lingkup (Out of Scope) — Rilis Awal

- Aplikasi mobile native (frontend web menjadi prioritas awal).
- Marketplace modul publik dengan sistem pembayaran (fase lanjutan).
- Modul vertikal khusus industri (dibangun sebagai modul terpisah nanti).
- Business Intelligence / analytics lanjutan (integrasi tool eksternal dulu).

## 6. Kebutuhan Fungsional (Tingkat Tinggi)

Kebutuhan fungsional rinci beserta kriteria penerimaan diuraikan dalam dokumen FSD. Berikut ringkasan kapabilitas utama yang harus disediakan produk:

| ID | Kapabilitas | Prioritas |
|---|---|---|
| PRD-F-01 | Registrasi & lifecycle modul (install, enable, disable, uninstall) | Must |
| PRD-F-02 | Dependency resolution antar-modul (topological sort) | Must |
| PRD-F-03 | Model data yang dapat di-extend oleh modul lain | Must |
| PRD-F-04 | Custom field runtime tanpa restart | Should |
| PRD-F-05 | Autentikasi user & mesin (session, API key, OAuth2) | Must |
| PRD-F-06 | Otorisasi berbasis peran per-modul (RBAC/ACL) | Must |
| PRD-F-07 | Multi-tenant & multi-company | Must |
| PRD-F-08 | Event bus untuk komunikasi antar-modul | Must |
| PRD-F-09 | Migrasi database versioned per-modul | Must |
| PRD-F-10 | REST API otomatis dari definisi model | Should |
| PRD-F-11 | Webhook keluar & connector integrasi async | Should |
| PRD-F-12 | Modul pihak ketiga via gRPC / WASM | Could |

## 7. Kebutuhan Non-Fungsional

| Kategori | Kebutuhan |
|---|---|
| Performa | P95 latensi CRUD < 200 ms; boot 10 modul < 5 detik. |
| Skalabilitas | Mendukung scaling horizontal core dan worker. |
| Keamanan | Isolasi tenant ketat; audit log; enkripsi kredensial. |
| Keandalan | Retry, dead-letter queue, idempotency untuk integrasi. |
| Maintainability | Pemisahan tegas core vs modul; modul tidak saling import langsung. |
| Observability | Logging terstruktur, metrik, dan tracing per-modul. |
| Portabilitas | Deployment sebagai binary tunggal dan container. |

## 8. Asumsi, Ketergantungan & Risiko

### 8.1 Asumsi

- PostgreSQL digunakan sebagai basis data utama (dukungan JSONB & schema fleksibel).
- Frontend dikembangkan terpisah dan mengonsumsi REST/GraphQL.

### 8.2 Ketergantungan

- Runtime WASM (mis. wazero) untuk modul pihak ketiga ringan.
- Message broker (mis. NATS) untuk event bus terdistribusi dan worker.

### 8.3 Risiko Utama

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Go tidak mendukung dynamic loading andal | Modul pihak ketiga sulit dipasang tanpa recompile | Gunakan strategi hybrid gRPC + WASM |
| Kompleksitas extension model antar-modul | Bug data & migrasi | Kontrak jelas, urutan boot deterministik, uji migrasi |
| Overhead integrasi eksternal | Latensi & kegagalan request user | Pola async, retry, circuit breaker, outbox |

## 9. Rencana Rilis (Roadmap Ringkas)

| Fase | Fokus | Keluaran Utama |
|---|---|---|
| Fase 1 | Fondasi core | Module registry, boot, ORM dasar, RBAC, migrasi |
| Fase 2 | Extensibility | Custom field, event bus, extension model |
| Fase 3 | Modul standar | base, inventory, sales, accounting |
| Fase 4 | Developer experience | SDK, scaffolding CLI, dokumentasi |
| Fase 5 | Integrasi & deployment | REST/webhook/connector, UI dinamis, multi-tenant |

## 10. Pertanyaan Terbuka

- Apakah GraphQL akan disediakan di rilis awal atau ditunda?
- Model lisensi & tata kelola untuk marketplace modul pihak ketiga?
- Batas dukungan bahasa untuk modul WASM (Rust, TinyGo, AssemblyScript)?