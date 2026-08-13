# Tutorial — Menambahkan Auth, User Management & Role Management (Go)

> Panduan menambahkan lapisan autentikasi, manajemen user, dan manajemen role (RBAC) ke backend ERP modular. Melanjutkan dari `Tutorial-Backend-Go-Modular-ERP.md`. Semua kode sudah diuji: build & vet bersih, unit test lolos, dan alur RBAC terbukti bekerja end-to-end (401/403/200 sesuai peran).

**Prasyarat:** Sudah menyelesaikan tutorial backend Go (punya `core/` dengan module registry, router, event bus, boot loader) dan tiga modul contoh.

**Hasil akhir:** login yang mengembalikan token, user & role tersimpan, endpoint yang dilindungi berdasarkan peran, dan user admin bawaan.

---

## Keputusan Arsitektur: Core vs Modul

Pertanyaan penting di awal: apakah auth masuk ke `core` atau jadi modul biasa? Jawabannya **keduanya, dengan pembagian yang tegas:**

| Masuk ke `core/auth` (mekanisme) | Masuk ke `modules/auth` (data & endpoint) |
|---|---|
| Hashing password | Tabel/store user & role |
| Pembuatan & verifikasi token | Endpoint login |
| Middleware auth | CRUD user |
| Guard `RequireAuth` / `RequireRole` | CRUD role & assign role |

Alasannya: mekanisme (hashing, token, guard) bersifat lintas-modul dan tidak berubah antar bisnis, jadi milik core. Data & endpoint (siapa user-nya, role apa saja) adalah domain bisnis yang bisa berbeda tiap deployment, jadi jadi modul yang memakai primitif core.

---

## Daftar Langkah

1. [Password hashing (core)](#langkah-1--password-hashing)
2. [Token bertanda tangan (core)](#langkah-2--token-bertanda-tangan)
3. [Identity pada Context (core)](#langkah-3--identity-pada-context)
4. [Middleware pada Router (core)](#langkah-4--middleware-pada-router)
5. [Middleware auth & guard (core)](#langkah-5--middleware-auth--guard)
6. [Modul auth: model & seed (modul)](#langkah-6--modul-auth-model--seed)
7. [Handler: login, user, role (modul)](#langkah-7--handler-login-user-role)
8. [Rakit di main & pasang middleware global](#langkah-8--rakit--pasang-middleware)
9. [Uji RBAC end-to-end](#langkah-9--uji-rbac-end-to-end)
10. [Langkah lanjut: produksi](#langkah-10--langkah-lanjut)

---

## Langkah 1 — Password Hashing

Password TIDAK BOLEH disimpan sebagai teks biasa. Kita simpan sebagai hash dengan salt.

Buat folder dan file `core/auth/password.go`:

```bash
mkdir -p core/auth
```

```go
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	saltLen    = 16
	keyLen     = 32
	iterations = 100_000
)

// HashPassword menghasilkan hash PBKDF2-HMAC-SHA256 dengan salt acak.
// Format: pbkdf2$<iter>$<salt_b64>$<hash_b64>
//
// CATATAN PRODUKSI: PBKDF2 dipilih agar bisa diimplementasikan murni dengan
// standard library (nol dependency). Untuk produksi, ganti dengan bcrypt
// (golang.org/x/crypto/bcrypt) atau argon2id — lebih tahan serangan hardware.
func HashPassword(plain string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := pbkdf2SHA256([]byte(plain), salt, iterations, keyLen)
	return fmt.Sprintf("pbkdf2$%d$%s$%s",
		iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword membandingkan password dengan hash (constant-time).
func VerifyPassword(plain, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2" {
		return false, errors.New("format hash tidak dikenal")
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, errors.New("iterasi tidak valid")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, err
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, err
	}
	got := pbkdf2SHA256([]byte(plain), salt, iter, len(want))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// pbkdf2SHA256 mengimplementasikan PBKDF2 (RFC 2898) dengan HMAC-SHA256.
func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	hashLen := sha256.Size
	numBlocks := (keyLen + hashLen - 1) / hashLen
	var dk []byte
	for block := 1; block <= numBlocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		var blockNum [4]byte
		binary.BigEndian.PutUint32(blockNum[:], uint32(block))
		mac.Write(blockNum[:])
		u := mac.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)
		for i := 1; i < iter; i++ {
			mac := hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		dk = append(dk, t...)
	}
	return dk[:keyLen]
}
```

> **Kenapa salt + banyak iterasi?** Salt acak membuat dua password identik menghasilkan hash berbeda (mencegah rainbow table). Iterasi tinggi (100.000) memperlambat brute-force. `subtle.ConstantTimeCompare` mencegah timing attack.

### Uji

Buat `core/auth/password_test.go`:

```go
package auth

import "testing"

func TestHashVerify(t *testing.T) {
	hash, err := HashPassword("rahasia123")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := VerifyPassword("rahasia123", hash)
	if err != nil || !ok {
		t.Fatalf("password benar harus terverifikasi, ok=%v err=%v", ok, err)
	}
	bad, _ := VerifyPassword("salah", hash)
	if bad {
		t.Fatal("password salah tidak boleh terverifikasi")
	}
}

func TestHashUnique(t *testing.T) {
	h1, _ := HashPassword("sama")
	h2, _ := HashPassword("sama")
	if h1 == h2 {
		t.Fatal("dua hash harus berbeda karena salt acak")
	}
}
```

Jalankan: `go test ./core/auth/ -v` — kedua test harus PASS.

---

## Langkah 2 — Token Bertanda Tangan

Setelah login, server memberi token yang dibawa client di setiap request. Token ditandatangani agar tidak bisa dipalsukan.

Buat `core/auth/token.go`:

```go
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Claims struct {
	UserID    int      `json:"uid"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
}

// TokenService menandatangani & memverifikasi token dengan HMAC-SHA256.
// Token signed bergaya JWT memakai hanya standard library.
type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), ttl: ttl}
}

func (s *TokenService) Issue(userID int, email string, roles []string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		ExpiresAt: time.Now().Add(s.ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	return body + "." + s.sign(body), nil
}

func (s *TokenService) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("format token tidak valid")
	}
	body, sig := parts[0], parts[1]
	if !hmac.Equal([]byte(sig), []byte(s.sign(body))) {
		return nil, errors.New("tanda tangan token tidak valid")
	}
	payload, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token kedaluwarsa")
	}
	return &claims, nil
}

func (s *TokenService) sign(body string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
```

> **Kenapa HMAC?** Server menandatangani token dengan secret yang hanya ia tahu. Client tidak bisa mengubah isi token (mis. menaikkan peran jadi admin) karena tanda tangan akan tidak cocok. `hmac.Equal` membandingkan constant-time.

Uji dengan `core/auth/token_test.go` (mencakup token valid, diutak-atik, kedaluwarsa, secret salah):

```go
package auth

import (
	"testing"
	"time"
)

func TestTokenRoundTrip(t *testing.T) {
	svc := NewTokenService("rahasia-super", time.Hour)
	tok, _ := svc.Issue(42, "user@x.com", []string{"admin"})
	claims, err := svc.Verify(tok)
	if err != nil {
		t.Fatalf("token valid harus lolos: %v", err)
	}
	if claims.UserID != 42 || claims.Roles[0] != "admin" {
		t.Fatalf("claims salah: %+v", claims)
	}
}

func TestTokenTampered(t *testing.T) {
	svc := NewTokenService("rahasia", time.Hour)
	tok, _ := svc.Issue(1, "a@b.com", nil)
	if _, err := svc.Verify("x" + tok[1:]); err == nil {
		t.Fatal("token diutak-atik harus ditolak")
	}
}

func TestTokenExpired(t *testing.T) {
	svc := NewTokenService("rahasia", -time.Minute)
	tok, _ := svc.Issue(1, "a@b.com", nil)
	if _, err := svc.Verify(tok); err == nil {
		t.Fatal("token kedaluwarsa harus ditolak")
	}
}
```

---

## Langkah 3 — Identity pada Context

Middleware auth perlu menaruh identitas user di suatu tempat yang bisa dibaca handler. Kita tambahkan ke `Context`.

Perbarui `core/context.go` — tambahkan tipe `Identity` dan field/method terkait:

```go
// Identity adalah identitas pengguna terautentikasi pada satu request.
type Identity struct {
	UserID        int
	Email         string
	Roles         []string
	Authenticated bool
}

func (id Identity) HasRole(role string) bool {
	for _, r := range id.Roles {
		if r == role {
			return true
		}
	}
	return false
}
```

Tambahkan field `identity Identity` ke struct `Context`, lalu method:

```go
func (c *Context) Identity() Identity      { return c.identity }
func (c *Context) SetIdentity(id Identity) { c.identity = id }
func (c *Context) Header(key string) string { return c.Request.Header.Get(key) }
```

> **Kenapa Identity di Context?** Setiap request punya Context sendiri, jadi identitas satu user tidak bocor ke request lain. `Authenticated: false` berarti anonim.

---

## Langkah 4 — Middleware pada Router

Auth diterapkan sebagai middleware — lapisan yang membungkus handler. Kita perlu router mendukungnya.

Perbarui interface `Router` di `core/module.go`:

```go
type Router interface {
	Handle(method, path string, h HandlerFunc)
	HandleWith(method, path string, h HandlerFunc, mw ...Middleware)
}

type HandlerFunc func(ctx *Context) error

// Middleware membungkus HandlerFunc (mis. auth, guard izin, logging).
type Middleware func(HandlerFunc) HandlerFunc
```

Perbarui `core/router.go` untuk mendukung middleware global (`Use`) dan per-rute (`HandleWith`). Tambahkan field `globalMW []Middleware` ke `httpRouter`, lalu:

```go
func (r *httpRouter) Use(mw Middleware) {
	r.globalMW = append(r.globalMW, mw)
}

func (r *httpRouter) Handle(method, path string, h HandlerFunc) {
	r.HandleWith(method, path, h)
}

func (r *httpRouter) HandleWith(method, path string, h HandlerFunc, mw ...Middleware) {
	wrapped := h
	for i := len(mw) - 1; i >= 0; i-- {
		wrapped = mw[i](wrapped)
	}
	r.routes = append(r.routes, route{method: method, pattern: splitPath(path), handler: wrapped})
}
```

Di `ServeHTTP`, sebelum memanggil handler, bungkus dengan middleware global:

```go
h := rt.handler
for i := len(r.globalMW) - 1; i >= 0; i-- {
	h = r.globalMW[i](h)
}
if err := h(ctx); err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
```

Tambahkan `Use` ke `App` di `core/app.go`:

```go
func (a *App) Use(mw Middleware) { a.router.Use(mw) }
```

> **Urutan pembungkusan.** Middleware dibungkus dari dalam ke luar, sehingga urutan eksekusi: global (terluar) → per-rute → handler. Auth global berjalan dulu (menetapkan identity), lalu guard per-rute memakainya.

---

## Langkah 5 — Middleware Auth & Guard

Sekarang middleware yang sesungguhnya. Buat `core/auth/middleware.go`:

```go
package auth

import (
	"strings"

	"myerp/core"
)

// Middleware membaca token Bearer, memverifikasi, menetapkan Identity.
// Request tanpa token tetap diteruskan sebagai anonim (guard yang menolak).
func Middleware(ts *TokenService) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			authHeader := c.Header("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if claims, err := ts.Verify(token); err == nil {
					c.SetIdentity(core.Identity{
						UserID:        claims.UserID,
						Email:         claims.Email,
						Roles:         claims.Roles,
						Authenticated: true,
					})
				}
			}
			return next(c)
		}
	}
}

// RequireAuth menolak request tidak terautentikasi (401).
func RequireAuth() core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			if !c.Identity().Authenticated {
				return c.JSON(401, map[string]string{"error": "autentikasi diperlukan"})
			}
			return next(c)
		}
	}
}

// RequireRole menolak request tanpa salah satu peran (403).
func RequireRole(roles ...string) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			id := c.Identity()
			if !id.Authenticated {
				return c.JSON(401, map[string]string{"error": "autentikasi diperlukan"})
			}
			for _, r := range roles {
				if id.HasRole(r) {
					return next(c)
				}
			}
			return c.JSON(403, map[string]string{"error": "izin tidak mencukupi"})
		}
	}
}
```

> **Tiga lapis, tugas berbeda.** `Middleware` hanya *menetapkan* identity — tidak pernah menolak, sehingga endpoint publik (login) tetap jalan. `RequireAuth` menolak anonim (401). `RequireRole` menolak peran kurang (403). Perbedaan 401 vs 403 penting: 401 = "kamu siapa?", 403 = "aku tahu kamu siapa, tapi kamu tidak boleh".

---

## Langkah 6 — Modul auth: Model & Seed

Sekarang data. Buat modul `auth` yang memakai primitif core di atas.

`modules/auth/models.go`:

```go
package auth

type Role struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type User struct {
	ID           int      `json:"id"`
	Email        string   `json:"email"`
	Name         string   `json:"name"`
	Roles        []string `json:"roles"`
	Active       bool     `json:"active"`
	passwordHash string   // huruf kecil → tidak diekspor ke JSON
}

type store struct {
	users  map[int]*User
	roles  map[string]*Role
	byMail map[string]int
	nextID int
}

func newStore() *store {
	return &store{
		users:  map[int]*User{},
		roles:  map[string]*Role{},
		byMail: map[string]int{},
		nextID: 1,
	}
}
```

> **Perhatikan `passwordHash` huruf kecil.** Field tidak ter-ekspor tidak ikut di-serialize JSON, sehingga hash password tidak pernah bocor ke respons API. Detail keamanan kecil tapi krusial.

`modules/auth/auth.go` — manifest + migrasi yang menyemai role & admin:

```go
package auth

import (
	"context"
	"log"
	"time"

	"myerp/core"
	coreauth "myerp/core/auth"
)

type Module struct {
	store *store
	ts    *coreauth.TokenService
}

func New(secret string) *Module {
	return &Module{
		store: newStore(),
		ts:    coreauth.NewTokenService(secret, 24*time.Hour),
	}
}

func (m *Module) Manifest() core.Manifest {
	return core.Manifest{
		Name:        "auth",
		Version:     "1.0.0",
		Description: "Autentikasi, user & role management",
	}
}

func (m *Module) Migrations() []core.Migration {
	return []core.Migration{
		{ID: "001_seed_roles", Up: func() error {
			m.store.roles["admin"] = &Role{Name: "admin", Permissions: []string{"*"}}
			m.store.roles["user"] = &Role{Name: "user", Permissions: []string{"sales.read", "inventory.read"}}
			return nil
		}},
		{ID: "002_seed_admin", Up: func() error {
			hash, err := coreauth.HashPassword("admin123")
			if err != nil {
				return err
			}
			u := &User{ID: m.store.nextID, Email: "admin@erp.local", Name: "Administrator",
				Roles: []string{"admin"}, Active: true, passwordHash: hash}
			m.store.users[u.ID] = u
			m.store.byMail[u.Email] = u.ID
			m.store.nextID++
			log.Println("auth: user admin dibuat (admin@erp.local / admin123)")
			return nil
		}},
	}
}

// TokenService diekspos agar main bisa memasang middleware auth global.
func (m *Module) TokenService() *coreauth.TokenService { return m.ts }

func (m *Module) OnInstall(ctx context.Context) error   { return nil }
func (m *Module) OnUninstall(ctx context.Context) error { return nil }
```

> **Role `admin` punya permission `"*"`** — wildcard yang berarti semua izin. Role `user` hanya baca. Migrasi juga membuat satu admin bawaan agar sistem bisa langsung dipakai (ganti password-nya di produksi!).

---

## Langkah 7 — Handler: Login, User, Role

`modules/auth/handlers.go` — endpoint, dengan guard yang sesuai:

```go
package auth

import (
	"encoding/json"
	"strconv"

	"myerp/core"
	coreauth "myerp/core/auth"
)

func (m *Module) RegisterRoutes(r core.Router) {
	r.Handle("POST", "/auth/login", m.login)                                    // publik
	r.HandleWith("GET", "/auth/me", m.me, coreauth.RequireAuth())               // login saja
	r.HandleWith("GET", "/auth/users", m.listUsers, coreauth.RequireRole("admin"))
	r.HandleWith("POST", "/auth/users", m.createUser, coreauth.RequireRole("admin"))
	r.HandleWith("POST", "/auth/users/:id/roles", m.assignRole, coreauth.RequireRole("admin"))
	r.HandleWith("GET", "/auth/roles", m.listRoles, coreauth.RequireRole("admin"))
	r.HandleWith("POST", "/auth/roles", m.createRole, coreauth.RequireRole("admin"))
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (m *Module) login(c *core.Context) error {
	var in loginReq
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		return c.JSON(400, map[string]string{"error": "body tidak valid"})
	}
	uid, ok := m.store.byMail[in.Email]
	if !ok {
		return c.JSON(401, map[string]string{"error": "email atau password salah"})
	}
	u := m.store.users[uid]
	valid, _ := coreauth.VerifyPassword(in.Password, u.passwordHash)
	if !valid || !u.Active {
		return c.JSON(401, map[string]string{"error": "email atau password salah"})
	}
	token, err := m.ts.Issue(u.ID, u.Email, u.Roles)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "gagal membuat token"})
	}
	return c.JSON(200, map[string]any{"token": token, "user": u})
}

func (m *Module) me(c *core.Context) error {
	u, ok := m.store.users[c.Identity().UserID]
	if !ok {
		return c.JSON(404, map[string]string{"error": "user tidak ditemukan"})
	}
	return c.JSON(200, u)
}

func (m *Module) listUsers(c *core.Context) error {
	list := make([]*User, 0, len(m.store.users))
	for _, u := range m.store.users {
		list = append(list, u)
	}
	return c.JSON(200, list)
}

type createUserReq struct {
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

func (m *Module) createUser(c *core.Context) error {
	var in createUserReq
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		return c.JSON(400, map[string]string{"error": "body tidak valid"})
	}
	if _, dup := m.store.byMail[in.Email]; dup {
		return c.JSON(409, map[string]string{"error": "email sudah terdaftar"})
	}
	for _, rn := range in.Roles {
		if _, ok := m.store.roles[rn]; !ok {
			return c.JSON(400, map[string]string{"error": "role tidak dikenal: " + rn})
		}
	}
	hash, err := coreauth.HashPassword(in.Password)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "gagal hash password"})
	}
	u := &User{ID: m.store.nextID, Email: in.Email, Name: in.Name,
		Roles: in.Roles, Active: true, passwordHash: hash}
	m.store.users[u.ID] = u
	m.store.byMail[u.Email] = u.ID
	m.store.nextID++
	c.Bus().Publish(core.Event{Name: "auth.user.created", Payload: u})
	return c.JSON(201, u)
}

type assignRoleReq struct {
	Role string `json:"role"`
}

func (m *Module) assignRole(c *core.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	u, ok := m.store.users[id]
	if !ok {
		return c.JSON(404, map[string]string{"error": "user tidak ditemukan"})
	}
	var in assignRoleReq
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		return c.JSON(400, map[string]string{"error": "body tidak valid"})
	}
	if _, ok := m.store.roles[in.Role]; !ok {
		return c.JSON(400, map[string]string{"error": "role tidak dikenal"})
	}
	for _, r := range u.Roles {
		if r == in.Role {
			return c.JSON(200, u)
		}
	}
	u.Roles = append(u.Roles, in.Role)
	return c.JSON(200, u)
}

func (m *Module) listRoles(c *core.Context) error {
	list := make([]*Role, 0, len(m.store.roles))
	for _, r := range m.store.roles {
		list = append(list, r)
	}
	return c.JSON(200, list)
}

func (m *Module) createRole(c *core.Context) error {
	var in Role
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		return c.JSON(400, map[string]string{"error": "body tidak valid"})
	}
	if in.Name == "" {
		return c.JSON(400, map[string]string{"error": "nama role wajib"})
	}
	if _, dup := m.store.roles[in.Name]; dup {
		return c.JSON(409, map[string]string{"error": "role sudah ada"})
	}
	m.store.roles[in.Name] = &in
	return c.JSON(201, in)
}
```

> **Login mengembalikan pesan seragam** ("email atau password salah") baik email tidak ada maupun password salah — mencegah penyerang menebak email mana yang terdaftar. `createUser` memvalidasi role benar-benar ada sebelum menyimpan.

---

## Langkah 8 — Rakit & Pasang Middleware

Perbarui `cmd/server/main.go`:

```go
package main

import (
	"log"

	"myerp/core"
	coreauth "myerp/core/auth"
	"myerp/modules/auth"
	"myerp/modules/base"
	"myerp/modules/inventory"
	"myerp/modules/sales"
)

func main() {
	app := core.New()

	// Modul auth dibuat lebih dulu agar token service-nya bisa dipakai middleware.
	authModule := auth.New("ganti-dengan-secret-yang-aman")

	// Middleware auth GLOBAL: setiap request dicek token-nya (opsional).
	app.Use(coreauth.Middleware(authModule.TokenService()))

	app.Register(sales.New())
	app.Register(authModule)
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

Verifikasi: `go build ./...` dan `go vet ./...` harus bersih.

> **Kenapa auth dibuat sebelum Register?** Middleware global butuh `TokenService` milik modul auth. Jadi kita buat modul dulu, ambil token service-nya untuk `app.Use`, baru daftarkan. Ini satu-satunya modul yang perlu perlakuan khusus karena menyediakan layanan lintas-modul.

---

## Langkah 9 — Uji RBAC End-to-End

Jalankan: `go run ./cmd/server`. Log boot akan menampilkan pembuatan admin.

Uji seluruh alur (butuh `curl` + `python3` untuk ekstrak token):

```bash
# 1. Tanpa token → 401
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/auth/users
# → 401

# 2. Login admin → dapat token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -d '{"email":"admin@erp.local","password":"admin123"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

# 3. Password salah → 401
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/auth/login \
  -d '{"email":"admin@erp.local","password":"salah"}'
# → 401

# 4. Admin akses /auth/users → 200
curl -s -o /dev/null -w "%{http_code}\n" -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/auth/users
# → 200

# 5. Admin buat user staff (role: user)
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/auth/users \
  -d '{"email":"staff@erp.local","name":"Staff","password":"staff123","roles":["user"]}'

# 6. Login staff, coba akses /auth/users → 403 (bukan admin!)
STAFF=$(curl -s -X POST http://localhost:8080/auth/login \
  -d '{"email":"staff@erp.local","password":"staff123"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
curl -s -o /dev/null -w "%{http_code}\n" -H "Authorization: Bearer $STAFF" \
  http://localhost:8080/auth/users
# → 403

# 7. Tapi staff boleh akses /auth/me → 200
curl -s -o /dev/null -w "%{http_code}\n" -H "Authorization: Bearer $STAFF" \
  http://localhost:8080/auth/me
# → 200
```

### Checklist pemahaman

- [ ] Tanpa token → 401 → *guard RequireRole menolak anonim*.
- [ ] Password salah → 401, dengan pesan seragam → *tidak membocorkan email terdaftar*.
- [ ] Admin → 200, staff → 403 untuk endpoint sama → *RBAC bekerja*.
- [ ] Staff → 200 untuk /auth/me → *RequireAuth cukup, tak perlu peran khusus*.
- [ ] `passwordHash` tak pernah muncul di JSON → *field tak ter-ekspor aman*.

---

## Langkah 10 — Langkah Lanjut

| Area | Sekarang (belajar) | Produksi |
|---|---|---|
| Hashing | PBKDF2 (stdlib murni) | bcrypt / argon2id |
| Token | HMAC signed sederhana | JWT penuh (jika perlu interop) + refresh token |
| Penyimpanan | in-memory map | PostgreSQL: tabel users, roles, user_roles |
| Permission check | role → wildcard/exact | ACL granular per resource+action, cache |
| Multi-tenant | belum | company_id di token & filter semua query |
| Session | token stateless | opsi revocation list / short-lived + refresh |
| Rate limit | belum | batasi percobaan login (anti brute-force) |
| Audit | belum | catat login, perubahan role di audit log |

### Integrasi dengan modul lain

Kini modul lain bisa melindungi endpoint-nya dengan guard yang sama. Contoh, modul `sales` membatasi penghapusan order hanya untuk admin:

```go
func (m *Module) RegisterRoutes(r core.Router) {
	r.Handle("GET", "/sales/orders", m.list) // publik/terautentikasi
	r.HandleWith("DELETE", "/sales/orders/:id", m.delete,
		coreauth.RequireRole("admin", "manager"))
}
```

Karena guard ada di `core/auth`, semua modul memakainya tanpa duplikasi. Inilah keuntungan menaruh mekanisme di core: satu implementasi, dipakai seluruh modul.

### Menghubungkan permission ke ACL granular

Struct `Role` sudah punya field `Permissions` (mis. `"sales.write"`). Langkah berikutnya: buat guard `RequirePermission("sales.write")` yang mengecek permission (bukan sekadar nama role), dengan mengekspos fungsi pengecekan permission dari modul auth. Ini memberi kontrol lebih halus daripada berbasis nama role saja.

---

## Ringkasan

Kamu telah menambahkan, dengan pembagian core vs modul yang tegas:

Di `core/auth`: hashing password (PBKDF2 + salt, constant-time verify), token bertanda tangan (HMAC, anti-tamper & expiry), serta middleware `Middleware`/`RequireAuth`/`RequireRole`. Di `core`: `Identity` pada Context dan dukungan middleware pada router. Di `modules/auth`: model user & role, seed admin, dan endpoint login + CRUD user/role yang dilindungi guard.

Seluruh alur terbukti: 401 untuk anonim, 403 untuk peran kurang, 200 untuk yang berhak — dengan password tersimpan aman sebagai hash dan tidak pernah bocor ke respons.
