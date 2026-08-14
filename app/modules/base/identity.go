package base

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/annaselh/gorbio/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AuthServiceName   = "base.auth"
	SessionCookieName = "gorbio_session"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrTooManyAttempts    = errors.New("too many login attempts")
)

type Principal struct {
	UserID       uuid.UUID
	SessionID    uuid.UUID
	TenantID     uuid.UUID
	MembershipID uuid.UUID
	Permissions  map[string]struct{}
}

func (p Principal) HasPermission(code string) bool {
	_, ok := p.Permissions[code]
	return ok
}

type LoginInput struct {
	Email      string
	Password   string
	TenantSlug string
	IPAddress  string
	UserAgent  string
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	Principal Principal
}

type BootstrapOwnerInput struct {
	TenantSlug  string
	TenantName  string
	Email       string
	DisplayName string
	Password    string
}

// loginRateLimit bounds password attempts per client IP. The per-account
// lockout below is a second, independent layer: on its own it lets anyone who
// knows an email address lock that account out at will, so the IP window is what
// actually absorbs a spray.
const (
	loginRateLimit  = 10
	loginRateWindow = 15 * time.Minute
)

// timingEqualizerPassword feeds a throwaway Argon2 verification on the
// account-not-found path so a missing email costs the same wall-clock time as a
// wrong password. Without it, response latency enumerates valid accounts.
const timingEqualizerPassword = "gorbio-timing-equalizer-placeholder"

var timingEqualizerHash = sync.OnceValue(func() string {
	hash, err := hashPassword(timingEqualizerPassword)
	if err != nil {
		return ""
	}
	return hash
})

type AuthService struct {
	db           *gorm.DB
	sessionTTL   time.Duration
	cookieSecure bool
	loginLimiter *core.RateLimiter
}

func NewAuthService(db *gorm.DB, settings core.Settings) *AuthService {
	ttl := settings.SessionTTL
	if ttl <= 0 {
		ttl = core.DefaultSettings().SessionTTL
	}
	return &AuthService{
		db:           db,
		sessionTTL:   ttl,
		cookieSecure: settings.CookieSecure,
		loginLimiter: core.NewRateLimiter(loginRateLimit, loginRateWindow),
	}
}

func AuthFromApp(app *core.App) (*AuthService, error) {
	service, ok := app.Services.Get(AuthServiceName)
	if !ok {
		return nil, fmt.Errorf("auth service is not registered")
	}
	auth, ok := service.(*AuthService)
	if !ok {
		return nil, fmt.Errorf("registered auth service has invalid type")
	}
	return auth, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	if input.IPAddress != "" && !s.loginLimiter.Allow(input.IPAddress) {
		return nil, ErrTooManyAttempts
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	var user User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		// Burn equivalent CPU so a missing account is indistinguishable by time.
		verifyPassword(timingEqualizerHash(), input.Password)
		return nil, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	// Evaluate the password unconditionally rather than letting || short-circuit
	// on status or lockout: an early return would make a disabled or locked
	// account answer measurably faster than a merely wrong password.
	passwordValid := verifyPassword(user.PasswordHash, input.Password)
	accountActive := user.Status == UserStatusActive
	accountUnlocked := user.LockedUntil == nil || !user.LockedUntil.After(now)

	if !passwordValid || !accountActive || !accountUnlocked {
		s.recordFailedLogin(ctx, &user, now)
		return nil, ErrInvalidCredentials
	}

	membership, err := s.findMembership(ctx, user.ID, input.TenantSlug)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return nil, err
	}
	session := Session{
		ID:             uuid.New(),
		UserID:         user.ID,
		ActiveTenantID: &membership.TenantID,
		TokenHash:      tokenHash,
		ExpiresAt:      now.Add(s.sessionTTL),
		LastUsedAt:     now,
		IPAddress:      input.IPAddress,
		UserAgent:      input.UserAgent,
	}
	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&user).Updates(map[string]any{
		"last_login_at": now, "failed_login_attempts": 0, "locked_until": nil,
	}).Error; err != nil {
		return nil, fmt.Errorf("update login state: %w", err)
	}

	principal, err := s.principalForMembership(ctx, user.ID, session.ID, membership)
	if err != nil {
		return nil, err
	}
	// A legitimate sign-in forgives earlier failures from this address, so a
	// shared NAT egress cannot lock out its own users.
	if input.IPAddress != "" {
		s.loginLimiter.Reset(input.IPAddress)
	}
	s.recordAudit(ctx, &user.ID, &membership.TenantID, "auth.login", "session", session.ID.String(), input.IPAddress, input.UserAgent)
	return &LoginResult{Token: token, ExpiresAt: session.ExpiresAt, Principal: principal}, nil
}

// BootstrapOwner creates the first tenant and its owner. It is intentionally
// not exposed as a public HTTP endpoint; invoke it from a protected setup
// command or deployment job exactly once.
func (s *AuthService) BootstrapOwner(ctx context.Context, input BootstrapOwnerInput) error {
	slug := strings.ToLower(strings.TrimSpace(input.TenantSlug))
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if slug == "" || input.TenantName == "" || input.DisplayName == "" || email == "" {
		return fmt.Errorf("tenant slug, tenant name, display name, and email are required")
	}
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Tenant{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("tenant %q already exists", slug)
		}

		now := time.Now().UTC()
		user := User{ID: uuid.New(), Email: email, DisplayName: input.DisplayName, PasswordHash: passwordHash, Status: UserStatusActive, EmailVerifiedAt: &now, PasswordChangedAt: &now}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("create owner user: %w", err)
		}
		tenant := Tenant{ID: uuid.New(), Slug: slug, Name: input.TenantName, Status: TenantStatusActive, Settings: []byte("{}")}
		if err := tx.Create(&tenant).Error; err != nil {
			return fmt.Errorf("create tenant: %w", err)
		}
		membership := Membership{ID: uuid.New(), UserID: user.ID, TenantID: tenant.ID, Status: MembershipStatusActive, JoinedAt: now}
		if err := tx.Create(&membership).Error; err != nil {
			return fmt.Errorf("create owner membership: %w", err)
		}
		var owner Role
		if err := tx.Where("code = ?", "owner").First(&owner).Error; err != nil {
			return fmt.Errorf("load owner role: %w", err)
		}
		if err := tx.Create(&MembershipRole{MembershipID: membership.ID, RoleID: owner.ID}).Error; err != nil {
			return fmt.Errorf("assign owner role: %w", err)
		}
		return nil
	})
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrUnauthorized
	}
	hash := hashToken(token)
	now := time.Now().UTC()
	var session Session
	if err := s.db.WithContext(ctx).Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, now).First(&session).Error; err != nil {
		return Principal{}, ErrUnauthorized
	}
	if session.ActiveTenantID == nil {
		return Principal{}, ErrUnauthorized
	}
	var membership Membership
	if err := s.db.WithContext(ctx).Where("user_id = ? AND tenant_id = ? AND status = ?", session.UserID, *session.ActiveTenantID, MembershipStatusActive).First(&membership).Error; err != nil {
		return Principal{}, ErrUnauthorized
	}
	if err := s.db.WithContext(ctx).Model(&session).Update("last_used_at", now).Error; err != nil {
		return Principal{}, fmt.Errorf("update session activity: %w", err)
	}
	return s.principalForMembership(ctx, session.UserID, session.ID, membership)
}

func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()
	return s.db.WithContext(ctx).Model(&Session{}).Where("id = ? AND revoked_at IS NULL", sessionID).Update("revoked_at", now).Error
}

func (s *AuthService) findMembership(ctx context.Context, userID uuid.UUID, tenantSlug string) (Membership, error) {
	query := s.db.WithContext(ctx).Model(&Membership{}).Where("memberships.user_id = ? AND memberships.status = ?", userID, MembershipStatusActive)
	if tenantSlug != "" {
		query = query.Joins("JOIN tenants ON tenants.id = memberships.tenant_id").Where("tenants.slug = ? AND tenants.status = ?", strings.ToLower(strings.TrimSpace(tenantSlug)), TenantStatusActive)
	}
	var membership Membership
	if err := query.Order("memberships.joined_at ASC").First(&membership).Error; err != nil {
		return Membership{}, err
	}
	return membership, nil
}

func (s *AuthService) principalForMembership(ctx context.Context, userID, sessionID uuid.UUID, membership Membership) (Principal, error) {
	var codes []string
	err := s.db.WithContext(ctx).Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN membership_roles ON membership_roles.role_id = role_permissions.role_id").
		Where("membership_roles.membership_id = ?", membership.ID).
		Pluck("permissions.code", &codes).Error
	if err != nil {
		return Principal{}, fmt.Errorf("load permissions: %w", err)
	}
	permissions := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		permissions[code] = struct{}{}
	}
	return Principal{UserID: userID, SessionID: sessionID, TenantID: membership.TenantID, MembershipID: membership.ID, Permissions: permissions}, nil
}

func (s *AuthService) recordFailedLogin(ctx context.Context, user *User, now time.Time) {
	updates := map[string]any{"failed_login_attempts": user.FailedLoginAttempts + 1}
	if user.FailedLoginAttempts+1 >= 5 {
		updates["locked_until"] = now.Add(15 * time.Minute)
	}
	_ = s.db.WithContext(ctx).Model(user).Updates(updates).Error
}

func (s *AuthService) recordAudit(ctx context.Context, actorID, tenantID *uuid.UUID, action, resourceType, resourceID, ipAddress, userAgent string) {
	_ = s.db.WithContext(ctx).Create(&AuditLog{ID: uuid.New(), ActorUserID: actorID, TenantID: tenantID, Action: action, ResourceType: resourceType, ResourceID: resourceID, Metadata: []byte("{}"), IPAddress: ipAddress, UserAgent: userAgent}).Error
}

func newSessionToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}
