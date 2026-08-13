# Tutorial Step-by-Step — Membangun Backend Go Modular ERP

> Panduan koding dari nol sampai server jalan. Setiap langkah menambah satu konsep, dengan penjelasan *kenapa*. Semua kode di tutorial ini sudah diuji: build bersih, `go vet` lolos, server jalan, dan endpoint merespons.

**Prasyarat:** Go 1.22+ terpasang (`go version`), editor, pemahaman dasar Go & HTTP.

**Hasil akhir:** Server ERP dengan module registry, dependency resolution (topological sort), boot loader, router HTTP, dan event bus antar-modul — semua dari tangan sendiri, tanpa framework eksternal.

> **Catatan desain:** Tutorial ini memakai penyimpanan **in-memory** (bukan database) agar fokus pada arsitektur modular tanpa terganggu setup DB. Di akhir ada panduan mengganti ke PostgreSQL. Konsep modularitasnya identik apa pun penyimpanannya.

---

## Daftar Langkah

1. [Inisialisasi proyek & struktur folder](#langkah-1--inisialisasi-proyek)
2. [Kontrak modul (`Module` interface)](#langkah-2--kontrak-modul)
3. [Tipe pendukung: Context & EventBus](#langkah-3--context--eventbus)
4. [Registry + dependency resolution](#langkah-4--registry--dependency-resolution)
5. [Router HTTP + boot loader](#langkah-5--router--boot-loader)
6. [Modul pertama: `base`](#langkah-6--modul-pertama-base)
7. [Modul `inventory` & `sales` + event bus](#langkah-7--inventory--sales--event-bus)
8. [Rakit `main.go` & jalankan](#langkah-8--rakit--jalankan)
9. [Uji endpoint & buktikan event bus](#langkah-9--uji--buktikan)
10. [Langkah lanjut: PostgreSQL & produksi](#langkah-10--langkah-lanjut)

---

## Langkah 1 — Inisialisasi Proyek

Buat folder proyek dan inisialisasi Go module:

```bash
mkdir myerp && cd myerp
go mod init myerp
```

Buat struktur folder:

```bash
mkdir -p cmd/server core modules/base modules/inventory modules/sales
```

Struktur akhir yang akan kita bangun:

```
myerp/
├── go.mod
├── cmd/server/main.go        # entry point, registrasi modul
├── core/                     # framework (tidak tahu logika bisnis)
│   ├── module.go             # interface Module + opsional
│   ├── context.go            # Context request/response
│   ├── eventbus.go           # publish/subscribe
│   ├── registry.go           # registry + ResolveOrder
│   ├── router.go             # router HTTP
│   └── app.go                # boot loader + serve
└── modules/                  # bisnis (boleh import core)
    ├── base/
    ├── inventory/
    └── sales/
```

> **Aturan dependency:** `cmd/server` → `modules` → `core`. Modul boleh import `core`, tapi **tidak** saling import langsung. Komunikasi antar-modul lewat event bus.

---

## Langkah 2 — Kontrak Modul

Ini fondasi semuanya. Buat `core/module.go`:

```go
package core

import "context"

// Manifest adalah metadata identitas modul.
type Manifest struct {
	Name         string
	Version      string
	Dependencies []string
	Description  string
}

// Module adalah kontrak WAJIB yang dipenuhi setiap modul.
type Module interface {
	Manifest() Manifest
}

// ── Interface OPSIONAL: modul mengimplementasikan sesuai kebutuhan ──

// Migratable: modul yang punya tabel/skema sendiri.
type Migratable interface {
	Migrations() []Migration
}

// Routable: modul yang punya endpoint HTTP.
type Routable interface {
	RegisterRoutes(r Router)
}

// EventSubscriber: modul yang bereaksi terhadap event modul lain.
type EventSubscriber interface {
	RegisterHooks(bus *EventBus)
}

// Lifecycle: hook saat install/uninstall.
type Lifecycle interface {
	OnInstall(ctx context.Context) error
	OnUninstall(ctx context.Context) error
}

// Migration adalah satu perubahan skema yang versioned.
type Migration struct {
	ID string
	Up func() error
}

// Router adalah abstraksi minimal (diisi di Langkah 5).
type Router interface {
	Handle(method, path string, h HandlerFunc)
}

type HandlerFunc func(ctx *Context) error
```

> **Konsep kunci — interface opsional.** Hanya `Module` (dengan `Manifest()`) yang wajib. Sisanya opsional: sebuah modul cukup mengimplementasikan yang relevan. Modul tanpa DB tak perlu `Migratable`; modul tanpa endpoint tak perlu `Routable`. Boot loader mendeteksi kapabilitas via *type assertion* (Langkah 5). Ini padanan Go untuk "modul cukup daftarkan yang dibutuhkan".

Jika kamu build sekarang (`go build ./core/`), akan **error**: `undefined: EventBus`, `undefined: Context`. Itu wajar — kontrak memerlukan tipe pendukung. Kita buat berikutnya.

---

## Langkah 3 — Context & EventBus

### 3.1 Context

Buat `core/context.go` — pembungkus request/response HTTP:

```go
package core

import (
	"encoding/json"
	"net/http"
)

// Context membungkus request/response HTTP + akses ke layanan core.
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	bus     *EventBus
	params  map[string]string
}

// JSON menulis respons JSON dengan status tertentu.
func (c *Context) JSON(status int, v any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	return json.NewEncoder(c.Writer).Encode(v)
}

// Bus memberi akses ke event bus (untuk publish dari handler).
func (c *Context) Bus() *EventBus { return c.bus }

// Param membaca path parameter (mis. :id).
func (c *Context) Param(key string) string { return c.params[key] }
```

### 3.2 EventBus

Buat `core/eventbus.go` — publish/subscribe in-memory:

```go
package core

import "sync"

// Event adalah pesan yang dipublish antar-modul.
type Event struct {
	Name    string
	Payload any
}

// EventHandler menerima event.
type EventHandler func(e Event)

// EventBus: publish/subscribe untuk komunikasi antar-modul.
type EventBus struct {
	mu   sync.RWMutex
	subs map[string][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{subs: map[string][]EventHandler{}}
}

// Subscribe mendaftarkan handler untuk sebuah nama event.
func (b *EventBus) Subscribe(name string, h EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[name] = append(b.subs[name], h)
}

// Publish mengirim event ke semua subscriber.
func (b *EventBus) Publish(e Event) {
	b.mu.RLock()
	handlers := append([]EventHandler(nil), b.subs[e.Name]...)
	b.mu.RUnlock()
	for _, h := range handlers {
		h(e)
	}
}
```

> **Kenapa `RWMutex`?** Subscribe terjadi saat boot (jarang), Publish terjadi saat request (sering). RWMutex membiarkan banyak Publish berjalan paralel. Kita menyalin slice handler sebelum memanggilnya agar lock dilepas cepat dan tidak deadlock bila handler mem-publish lagi.

Sekarang `go build ./core/` harus lolos.

---

## Langkah 4 — Registry & Dependency Resolution

Inti dari "modular": mengurutkan modul agar dependency selalu di-boot lebih dulu. Buat `core/registry.go`:

```go
package core

import "fmt"

// Registry menampung modul dan mengurutkannya berdasarkan dependency.
type Registry struct {
	modules []Module
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Register(m Module) {
	r.modules = append(r.modules, m)
}

// ResolveOrder mengurutkan modul via topological sort.
func (r *Registry) ResolveOrder() ([]Module, error) {
	byName := map[string]Module{}
	for _, m := range r.modules {
		name := m.Manifest().Name
		if _, dup := byName[name]; dup {
			return nil, fmt.Errorf("modul duplikat: %q", name)
		}
		byName[name] = m
	}

	var ordered []Module
	visited := map[string]bool{} // selesai
	inStack := map[string]bool{} // sedang dikunjungi (deteksi siklus)

	var visit func(m Module) error
	visit = func(m Module) error {
		name := m.Manifest().Name
		if visited[name] {
			return nil
		}
		if inStack[name] {
			return fmt.Errorf("dependency melingkar terdeteksi pada: %q", name)
		}
		inStack[name] = true
		for _, dep := range m.Manifest().Dependencies {
			depMod, ok := byName[dep]
			if !ok {
				return fmt.Errorf("dependency hilang: %q dibutuhkan oleh %q", dep, name)
			}
			if err := visit(depMod); err != nil {
				return err
			}
		}
		inStack[name] = false
		visited[name] = true
		ordered = append(ordered, m)
		return nil
	}

	for _, m := range r.modules {
		if err := visit(m); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}
```

### Uji dengan unit test

Buat `core/registry_test.go`:

```go
package core

import "testing"

type fakeModule struct {
	name string
	deps []string
}

func (f fakeModule) Manifest() Manifest {
	return Manifest{Name: f.name, Dependencies: f.deps}
}

func TestResolveOrder(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeModule{name: "sales", deps: []string{"base", "inventory"}})
	r.Register(fakeModule{name: "base"})
	r.Register(fakeModule{name: "inventory", deps: []string{"base"}})

	ordered, err := r.ResolveOrder()
	if err != nil {
		t.Fatalf("tak terduga error: %v", err)
	}
	pos := map[string]int{}
	for i, m := range ordered {
		pos[m.Manifest().Name] = i
	}
	if pos["base"] > pos["inventory"] {
		t.Errorf("base harus sebelum inventory")
	}
	if pos["inventory"] > pos["sales"] {
		t.Errorf("inventory harus sebelum sales")
	}
}

func TestMissingDependency(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeModule{name: "sales", deps: []string{"tidakada"}})
	if _, err := r.ResolveOrder(); err == nil {
		t.Fatal("harus error karena dependency hilang")
	}
}

func TestCyclicDependency(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeModule{name: "a", deps: []string{"b"}})
	r.Register(fakeModule{name: "b", deps: []string{"a"}})
	if _, err := r.ResolveOrder(); err == nil {
		t.Fatal("harus error karena siklus")
	}
}
```

Jalankan:

```bash
go test ./core/ -v
```

Ketiga test harus **PASS**: sorting benar, dependency hilang tertangkap, siklus tertangkap.

> **Kenapa dua map (`visited` + `inStack`)?** `visited` mencegah memproses modul dua kali. `inStack` melacak modul yang sedang dalam jalur rekursi saat ini — jika kita bertemu modul yang masih `inStack`, berarti ada siklus (A→B→A).

---

## Langkah 5 — Router & Boot Loader

### 5.1 Router HTTP

Buat `core/router.go` — router minimal di atas `net/http`, mendukung path param (`:id`):

```go
package core

import (
	"net/http"
	"strings"
)

type route struct {
	method  string
	pattern []string
	handler HandlerFunc
}

type httpRouter struct {
	routes []route
	bus    *EventBus
}

func newRouter(bus *EventBus) *httpRouter {
	return &httpRouter{bus: bus}
}

func (r *httpRouter) Handle(method, path string, h HandlerFunc) {
	r.routes = append(r.routes, route{
		method:  method,
		pattern: splitPath(path),
		handler: h,
	})
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

func (r *httpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	segs := splitPath(req.URL.Path)
	for _, rt := range r.routes {
		if rt.method != req.Method || len(rt.pattern) != len(segs) {
			continue
		}
		params := map[string]string{}
		matched := true
		for i, pat := range rt.pattern {
			if strings.HasPrefix(pat, ":") {
				params[pat[1:]] = segs[i]
			} else if pat != segs[i] {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		ctx := &Context{Writer: w, Request: req, bus: r.bus, params: params}
		if err := rt.handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.NotFound(w, req)
}
```

> **Catatan:** Ini router edukatif agar kamu paham cara kerjanya. Di produksi ganti dengan `chi`, `gin`, atau `echo` yang menangani radix tree, middleware, dan kasus tepi lebih baik. Karena `Router` adalah interface, penggantian ini tidak menyentuh kode modul.

### 5.2 Boot Loader

Buat `core/app.go` — menyatukan registry, bus, router, dan mendeteksi kapabilitas modul:

```go
package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type App struct {
	registry *Registry
	bus      *EventBus
	router   *httpRouter
}

func New() *App {
	bus := NewEventBus()
	return &App{
		registry: NewRegistry(),
		bus:      bus,
		router:   newRouter(bus),
	}
}

func (a *App) Register(m Module) { a.registry.Register(m) }

// Boot mengurutkan modul dan mendaftarkan kapabilitas masing-masing.
func (a *App) Boot() error {
	ordered, err := a.registry.ResolveOrder()
	if err != nil {
		return err
	}
	for _, m := range ordered {
		name := m.Manifest().Name

		if mig, ok := m.(Migratable); ok {
			for _, mg := range mig.Migrations() {
				if err := mg.Up(); err != nil {
					return fmt.Errorf("migrasi %s pada modul %s gagal: %w", mg.ID, name, err)
				}
			}
		}
		if rt, ok := m.(Routable); ok {
			rt.RegisterRoutes(a.router)
		}
		if es, ok := m.(EventSubscriber); ok {
			es.RegisterHooks(a.bus)
		}
		if lc, ok := m.(Lifecycle); ok {
			if err := lc.OnInstall(context.Background()); err != nil {
				return fmt.Errorf("OnInstall modul %s gagal: %w", name, err)
			}
		}
		log.Printf("modul di-boot: %s", name)
	}
	return nil
}

func (a *App) Serve(addr string) error {
	log.Printf("server berjalan di %s", addr)
	return http.ListenAndServe(addr, a.router)
}

func (a *App) Bus() *EventBus { return a.bus }
```

> **Ini jantung sistem.** Pola `if x, ok := m.(Interface); ok { ... }` adalah cara Go menanyakan "apakah modul ini punya kapabilitas X?". Boot berjalan dalam urutan hasil topological sort, jadi migrasi & registrasi modul induk selalu jalan sebelum modul yang bergantung padanya.

`go build ./core/` harus lolos.

---

## Langkah 6 — Modul Pertama: base

Sekarang kita tulis modul bisnis. `base` punya model `Partner`, migrasi (seed data), dan endpoint.

### 6.1 Definisi modul & migrasi

Buat `modules/base/base.go`:

```go
package base

import (
	"context"
	"log"

	"myerp/core"
)

type Partner struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type store struct {
	partners map[int]Partner
	nextID   int
}

type Module struct {
	store *store
}

func New() *Module {
	return &Module{store: &store{partners: map[int]Partner{}, nextID: 1}}
}

// Manifest — kontrak WAJIB.
func (m *Module) Manifest() core.Manifest {
	return core.Manifest{
		Name:        "base",
		Version:     "1.0.0",
		Description: "Users, company, partner",
	}
}

// Migrations — OPSIONAL (Migratable). Di sini seed satu data contoh.
func (m *Module) Migrations() []core.Migration {
	return []core.Migration{
		{ID: "001_seed_partner", Up: func() error {
			m.store.partners[m.store.nextID] = Partner{
				ID: m.store.nextID, Name: "PT Contoh", Email: "contoh@example.com",
			}
			m.store.nextID++
			return nil
		}},
	}
}

// RegisterRoutes — OPSIONAL (Routable).
func (m *Module) RegisterRoutes(r core.Router) {
	r.Handle("GET", "/base/partners", m.listPartners)
	r.Handle("GET", "/base/partners/:id", m.getPartner)
	r.Handle("POST", "/base/partners", m.createPartner)
}

// OnInstall — OPSIONAL (Lifecycle).
func (m *Module) OnInstall(ctx context.Context) error {
	log.Println("base: OnInstall dipanggil")
	return nil
}
func (m *Module) OnUninstall(ctx context.Context) error { return nil }
```

### 6.2 Handler

Buat `modules/base/handlers.go`:

```go
package base

import (
	"encoding/json"
	"strconv"

	"myerp/core"
)

func (m *Module) listPartners(c *core.Context) error {
	list := make([]Partner, 0, len(m.store.partners))
	for _, p := range m.store.partners {
		list = append(list, p)
	}
	return c.JSON(200, list)
}

func (m *Module) getPartner(c *core.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	p, ok := m.store.partners[id]
	if !ok {
		return c.JSON(404, map[string]string{"error": "partner tidak ditemukan"})
	}
	return c.JSON(200, p)
}

func (m *Module) createPartner(c *core.Context) error {
	var in Partner
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		return c.JSON(400, map[string]string{"error": "body tidak valid"})
	}
	in.ID = m.store.nextID
	m.store.nextID++
	m.store.partners[in.ID] = in

	// Publish event agar modul lain (mis. sales) bisa bereaksi.
	c.Bus().Publish(core.Event{Name: "base.partner.created", Payload: in})

	return c.JSON(201, in)
}
```

`go build ./...` harus lolos.

> **Perhatikan `createPartner`:** setelah menyimpan, ia mem-*publish* event `base.partner.created`. Modul `base` tidak tahu siapa yang mendengarkan — ia hanya mengumumkan. Ini kunci loose coupling.

---

## Langkah 7 — inventory & sales + Event Bus

### 7.1 inventory (bergantung pada base)

Buat `modules/inventory/inventory.go`:

```go
package inventory

import "myerp/core"

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Manifest() core.Manifest {
	return core.Manifest{
		Name:         "inventory",
		Version:      "1.0.0",
		Dependencies: []string{"base"},
		Description:  "Produk & stok",
	}
}

func (m *Module) RegisterRoutes(r core.Router) {
	r.Handle("GET", "/inventory/products", func(c *core.Context) error {
		return c.JSON(200, []map[string]any{
			{"id": 1, "sku": "SKU-001", "name": "Contoh Produk"},
		})
	})
}
```

### 7.2 sales (bergantung base+inventory, dan MENDENGARKAN event base)

Buat `modules/sales/sales.go`:

```go
package sales

import (
	"log"

	"myerp/core"
)

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Manifest() core.Manifest {
	return core.Manifest{
		Name:         "sales",
		Version:      "1.0.0",
		Dependencies: []string{"base", "inventory"},
		Description:  "Pesanan penjualan",
	}
}

func (m *Module) RegisterRoutes(r core.Router) {
	r.Handle("GET", "/sales/orders", func(c *core.Context) error {
		return c.JSON(200, []map[string]any{
			{"id": 1, "partner_id": 1, "total": 150000, "state": "draft"},
		})
	})
}

// RegisterHooks — sales BEREAKSI terhadap event base TANPA mengimpor base.
func (m *Module) RegisterHooks(bus *core.EventBus) {
	bus.Subscribe("base.partner.created", func(e core.Event) {
		log.Printf("sales mendengar partner baru dibuat: %+v", e.Payload)
	})
}
```

> **Ini inti modularitas.** Buka `import` di `sales.go` — **tidak ada** `myerp/modules/base`. Kedua modul hanya berbagi *nama event* (string `"base.partner.created"`) lewat `core`. Kamu bisa menghapus modul `sales` dan `base` tetap jalan; kamu bisa menambah modul lain yang mendengarkan event yang sama tanpa menyentuh `base`.

`go build ./...` harus lolos.

---

## Langkah 8 — Rakit & Jalankan

Buat `cmd/server/main.go`:

```go
package main

import (
	"log"

	"myerp/core"
	"myerp/modules/base"
	"myerp/modules/inventory"
	"myerp/modules/sales"
)

func main() {
	app := core.New()

	// Daftarkan modul (urutan acak — boot loader yang mengurutkan).
	app.Register(sales.New())
	app.Register(base.New())
	app.Register(inventory.New())

	if err := app.Boot(); err != nil {
		log.Fatalf("boot gagal: %v", err)
	}

	if err := app.Serve(":8080"); err != nil {
		log.Fatal(err)
	}
}
```

Verifikasi:

```bash
go build ./...
go vet ./...
```

Keduanya harus bersih.

> **Perhatikan urutan registrasi:** `sales, base, inventory` — sengaja acak. Boot loader tetap mem-boot dalam urutan `base → inventory → sales`. Inilah gunanya `ResolveOrder`.

---

## Langkah 9 — Uji & Buktikan

Jalankan server:

```bash
go run ./cmd/server
```

Kamu akan melihat log boot (perhatikan urutannya):

```
base: OnInstall dipanggil
modul di-boot: base
modul di-boot: inventory
modul di-boot: sales
server berjalan di :8080
```

Di terminal lain, uji endpoint:

```bash
# Seed dari migrasi base harus muncul
curl http://localhost:8080/base/partners
# → [{"id":1,"name":"PT Contoh","email":"contoh@example.com"}]

# Buat partner baru — ini akan MEMICU event ke modul sales
curl -X POST http://localhost:8080/base/partners \
  -d '{"name":"PT Baru","email":"baru@x.com"}'
# → {"id":2,"name":"PT Baru","email":"baru@x.com"}

# Endpoint modul lain
curl http://localhost:8080/sales/orders
curl http://localhost:8080/inventory/products
```

Setelah POST, **cek log server** — akan muncul baris dari modul sales:

```
sales mendengar partner baru dibuat: {ID:2 Name:PT Baru Email:baru@x.com}
```

**Selamat.** Kamu baru saja membuktikan seluruh rantai: boot berurutan, migrasi jalan, routing per-modul, dan event bus yang menghubungkan modul tanpa coupling.

### Checklist pemahaman

- [ ] Menu boot terurut `base → inventory → sales` meski registrasi acak → *topological sort bekerja*.
- [ ] `GET /base/partners` mengembalikan seed → *migrasi jalan saat boot*.
- [ ] POST partner memunculkan log di sales → *event bus bekerja tanpa import langsung*.
- [ ] Menghapus `app.Register(sales.New())` dari main → base & inventory tetap jalan → *modul benar-benar pluggable*.

---

## Langkah 10 — Langkah Lanjut

Kerangka belajar ini menunjukkan semua konsep inti. Untuk menuju produksi, tingkatkan bertahap.

### 10.1 Ganti in-memory ke PostgreSQL

Tambahkan driver:

```bash
go get github.com/jackc/pgx/v5
```

Ubah `Migration.Up` agar menerima transaksi DB dan jalankan `CREATE TABLE`/`ALTER TABLE`. Simpan riwayat migrasi di tabel `migration_log` (mana yang sudah jalan) agar tidak dijalankan dua kali. Ganti `store` in-memory dengan query SQL. Untuk custom field runtime, tambahkan kolom `JSONB` dan tabel `ir_model_field` (lihat dokumen TSD & ERD).

### 10.2 Peningkatan yang disarankan

| Area | Sekarang (belajar) | Produksi |
|---|---|---|
| Router | `core/router.go` edukatif | chi / gin / echo |
| Penyimpanan | in-memory `map` | PostgreSQL (pgx) + migrasi versioned |
| Event bus | in-memory sinkron | NATS / outbox pattern untuk durability |
| Migrasi | fungsi `Up()` sederhana | transaksi + `migration_log` + rollback |
| Auth | belum ada | middleware API key / OAuth2 + multi-tenant |
| Extension model | belum ada | ModelRegistry + `Extend()` (pola `_inherit`) |
| Config | hardcode `:8080` | env / file config |
| Observability | `log` standar | logging terstruktur + metrik + tracing |

### 10.3 Menambah kapabilitas ModelRegistry (extension antar-modul)

Langkah berikut yang paling menarik: buat `core/orm` dengan `ModelRegistry` yang punya `Define()` dan `Extend()`, sehingga modul `sales` bisa menambah field `credit_limit` ke model `partner` milik `base` — padanan backend dari "slot" di frontend. Polanya: `base` memanggil `reg.Define("partner", ...)`, `sales` memanggil `reg.Extend("partner", Field{Name:"credit_limit", ...})`. Karena boot terurut, `base` selalu mendefinisikan sebelum `sales` meng-extend.

---

## Ringkasan Alur Belajar

Kamu telah membangun, dari nol:

1. **Kontrak modul** (`module.go`) — interface wajib + opsional.
2. **Context & EventBus** — pembungkus request + komunikasi antar-modul.
3. **Registry + dependency resolution** — topological sort dengan deteksi siklus & dependency hilang.
4. **Router + boot loader** — deteksi kapabilitas via type assertion, boot berurutan.
5. **Tiga modul** (`base`, `inventory`, `sales`) — dengan migrasi, route, dan event.
6. **Bukti loose coupling** — `sales` bereaksi ke event `base` tanpa import langsung.

Setiap konsep sejajar dengan tutorial frontend: `Module` interface ↔ `ModuleManifest`, `ResolveOrder` ↔ `resolveDependencyOrder`, event bus ↔ slot/emitter. Dengan begitu, satu tim bisa memiliki satu modul utuh dari database sampai UI — inti dari ERP modular ala Odoo, dengan stack Go + React.
