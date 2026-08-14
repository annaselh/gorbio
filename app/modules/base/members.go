package base

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrMemberNotFound  = errors.New("member not found")
	ErrAlreadyMember   = errors.New("this address is already a member of the tenant")
	ErrUnknownRole     = errors.New("unknown role")
	ErrLastOwner       = errors.New("the last owner of a tenant cannot be removed or demoted")
	ErrCannotSelfAlter = errors.New("you cannot change your own membership")
)

const ownerRoleCode = "owner"

// Member is the tenant-scoped view of a user: who they are plus what they may
// do here. The same user may be a member of several tenants with different
// roles in each.
type Member struct {
	MembershipID  uuid.UUID        `json:"membership_id"`
	UserID        uuid.UUID        `json:"user_id"`
	Email         string           `json:"email"`
	DisplayName   string           `json:"display_name"`
	Status        MembershipStatus `json:"status"`
	Roles         []string         `json:"roles"`
	JoinedAt      time.Time        `json:"joined_at"`
	LastLoginAt   *time.Time       `json:"last_login_at,omitempty"`
	EmailVerified bool             `json:"email_verified"`
}

type InviteMemberInput struct {
	Email       string
	DisplayName string
	RoleCodes   []string
}

// ListMembers returns everyone with a membership in the tenant.
func (s *AuthService) ListMembers(ctx context.Context, tenantID uuid.UUID) ([]Member, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant scope is required")
	}

	type row struct {
		MembershipID    uuid.UUID
		UserID          uuid.UUID
		Email           string
		DisplayName     string
		Status          MembershipStatus
		JoinedAt        time.Time
		LastLoginAt     *time.Time
		EmailVerifiedAt *time.Time
	}

	var rows []row
	err := s.db.WithContext(ctx).
		Table("memberships").
		Select(`memberships.id AS membership_id, memberships.user_id, memberships.status,
			memberships.joined_at, users.email, users.display_name,
			users.last_login_at, users.email_verified_at`).
		Joins("JOIN users ON users.id = memberships.user_id").
		Where("memberships.tenant_id = ? AND users.deleted_at IS NULL", tenantID).
		Order("users.display_name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	if len(rows) == 0 {
		return []Member{}, nil
	}

	membershipIDs := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		membershipIDs = append(membershipIDs, r.MembershipID)
	}

	// One query for every membership's roles rather than one per member.
	type roleRow struct {
		MembershipID uuid.UUID
		Code         string
	}
	var roleRows []roleRow
	err = s.db.WithContext(ctx).
		Table("membership_roles").
		Select("membership_roles.membership_id, roles.code").
		Joins("JOIN roles ON roles.id = membership_roles.role_id").
		Where("membership_roles.membership_id IN ?", membershipIDs).
		Order("roles.code ASC").
		Scan(&roleRows).Error
	if err != nil {
		return nil, fmt.Errorf("load member roles: %w", err)
	}

	rolesByMembership := make(map[uuid.UUID][]string, len(rows))
	for _, r := range roleRows {
		rolesByMembership[r.MembershipID] = append(rolesByMembership[r.MembershipID], r.Code)
	}

	members := make([]Member, 0, len(rows))
	for _, r := range rows {
		roles := rolesByMembership[r.MembershipID]
		if roles == nil {
			roles = []string{}
		}
		members = append(members, Member{
			MembershipID: r.MembershipID, UserID: r.UserID,
			Email: r.Email, DisplayName: r.DisplayName, Status: r.Status,
			Roles: roles, JoinedAt: r.JoinedAt, LastLoginAt: r.LastLoginAt,
			EmailVerified: r.EmailVerifiedAt != nil,
		})
	}
	return members, nil
}

// ListRoles returns the assignable roles.
func (s *AuthService) ListRoles(ctx context.Context) ([]Role, error) {
	var roles []Role
	if err := s.db.WithContext(ctx).Order("code ASC").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

// InviteMember adds someone to the tenant, creating the user record when the
// address is new.
//
// An invited user has no password: they are created pending and must go through
// the password reset flow to set one. That avoids ever transmitting a
// credential, and reuses the token machinery that already exists.
func (s *AuthService) InviteMember(ctx context.Context, principal Principal, input InviteMemberInput) (*Member, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	displayName := strings.TrimSpace(input.DisplayName)
	if email == "" || displayName == "" {
		return nil, fmt.Errorf("%w: email and display name are required", ErrInvalidInput)
	}
	if len(input.RoleCodes) == 0 {
		input.RoleCodes = []string{"member"}
	}

	var membershipID uuid.UUID
	var userID uuid.UUID
	var resetToken string

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		roles, err := rolesByCodes(tx, input.RoleCodes)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		var user User
		switch err := tx.Where("email = ?", email).First(&user).Error; {
		case err == nil:
			// Existing user joining another tenant.
		case errors.Is(err, gorm.ErrRecordNotFound):
			user = User{
				ID: uuid.New(), Email: email, DisplayName: displayName,
				// No password hash: sign-in is impossible until the invitee
				// completes a reset, and verifyPassword rejects an empty hash.
				PasswordHash: "", Status: UserStatusPending,
			}
			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("create invited user: %w", err)
			}
		default:
			return fmt.Errorf("look up invited user: %w", err)
		}
		userID = user.ID

		var existing int64
		if err := tx.Model(&Membership{}).
			Where("user_id = ? AND tenant_id = ?", user.ID, principal.TenantID).
			Count(&existing).Error; err != nil {
			return fmt.Errorf("check existing membership: %w", err)
		}
		if existing > 0 {
			return ErrAlreadyMember
		}

		membership := Membership{
			ID: uuid.New(), UserID: user.ID, TenantID: principal.TenantID,
			Status: MembershipStatusActive, JoinedAt: now,
		}
		if err := tx.Create(&membership).Error; err != nil {
			return fmt.Errorf("create membership: %w", err)
		}
		membershipID = membership.ID

		assignments := make([]MembershipRole, 0, len(roles))
		for _, role := range roles {
			assignments = append(assignments, MembershipRole{MembershipID: membership.ID, RoleID: role.ID})
		}
		if err := tx.Create(&assignments).Error; err != nil {
			return fmt.Errorf("assign roles: %w", err)
		}

		// Issue a reset token so the invite email can carry a "set your
		// password" link even for a brand new account.
		token, tokenHash, err := newSessionToken()
		if err != nil {
			return err
		}
		resetToken = token
		if err := tx.Create(&PasswordResetToken{
			ID: uuid.New(), UserID: user.ID, TokenHash: tokenHash,
			// Invites get a longer window than a self-service reset: the
			// recipient may not be at their desk.
			ExpiresAt: now.Add(7 * 24 * time.Hour),
		}).Error; err != nil {
			return fmt.Errorf("create invite token: %w", err)
		}

		s.recordAuditTx(ctx, tx, &principal.UserID, &principal.TenantID,
			"membership.invited", "membership", membership.ID.String())
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Mail is sent after commit: a delivery failure must not roll back a
	// membership that already exists.
	var invited User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&invited).Error; err == nil {
		if mailErr := s.sendInviteEmail(ctx, invited, resetToken); mailErr != nil {
			return nil, fmt.Errorf("membership created but invite email failed: %w", mailErr)
		}
	}

	return s.memberByID(ctx, principal.TenantID, membershipID)
}

// UpdateMemberRoles replaces a member's role set.
func (s *AuthService) UpdateMemberRoles(ctx context.Context, principal Principal, membershipID uuid.UUID, roleCodes []string) (*Member, error) {
	if len(roleCodes) == 0 {
		return nil, fmt.Errorf("%w: at least one role is required", ErrInvalidInput)
	}
	if membershipID == principal.MembershipID {
		return nil, ErrCannotSelfAlter
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var membership Membership
		if err := tx.Where("id = ? AND tenant_id = ?", membershipID, principal.TenantID).
			First(&membership).Error; err != nil {
			return ErrMemberNotFound
		}

		roles, err := rolesByCodes(tx, roleCodes)
		if err != nil {
			return err
		}

		losingOwner := !containsRoleCode(roles, ownerRoleCode)
		if losingOwner {
			if err := guardLastOwner(tx, principal.TenantID, membershipID); err != nil {
				return err
			}
		}

		if err := tx.Where("membership_id = ?", membershipID).
			Delete(&MembershipRole{}).Error; err != nil {
			return fmt.Errorf("clear roles: %w", err)
		}
		assignments := make([]MembershipRole, 0, len(roles))
		for _, role := range roles {
			assignments = append(assignments, MembershipRole{MembershipID: membershipID, RoleID: role.ID})
		}
		if err := tx.Create(&assignments).Error; err != nil {
			return fmt.Errorf("assign roles: %w", err)
		}

		s.recordAuditTx(ctx, tx, &principal.UserID, &principal.TenantID,
			"membership.roles_changed", "membership", membershipID.String())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.memberByID(ctx, principal.TenantID, membershipID)
}

// SetMemberStatus suspends or reactivates a membership. Suspension revokes the
// member's sessions immediately - otherwise they keep working until their
// current session expires, which defeats the point.
func (s *AuthService) SetMemberStatus(ctx context.Context, principal Principal, membershipID uuid.UUID, status MembershipStatus) (*Member, error) {
	if status != MembershipStatusActive && status != MembershipStatusSuspended {
		return nil, fmt.Errorf("%w: unknown membership status %q", ErrInvalidInput, status)
	}
	if membershipID == principal.MembershipID {
		return nil, ErrCannotSelfAlter
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var membership Membership
		if err := tx.Where("id = ? AND tenant_id = ?", membershipID, principal.TenantID).
			First(&membership).Error; err != nil {
			return ErrMemberNotFound
		}

		if status == MembershipStatusSuspended {
			if err := guardLastOwner(tx, principal.TenantID, membershipID); err != nil {
				return err
			}
		}

		if err := tx.Model(&membership).Update("status", status).Error; err != nil {
			return fmt.Errorf("update membership status: %w", err)
		}

		if status == MembershipStatusSuspended {
			now := time.Now().UTC()
			if err := tx.Model(&Session{}).
				Where("user_id = ? AND active_tenant_id = ? AND revoked_at IS NULL",
					membership.UserID, principal.TenantID).
				Update("revoked_at", now).Error; err != nil {
				return fmt.Errorf("revoke member sessions: %w", err)
			}
		}

		s.recordAuditTx(ctx, tx, &principal.UserID, &principal.TenantID,
			"membership.status_changed", "membership", membershipID.String())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.memberByID(ctx, principal.TenantID, membershipID)
}

func (s *AuthService) memberByID(ctx context.Context, tenantID, membershipID uuid.UUID) (*Member, error) {
	members, err := s.ListMembers(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for i := range members {
		if members[i].MembershipID == membershipID {
			return &members[i], nil
		}
	}
	return nil, ErrMemberNotFound
}

func rolesByCodes(tx *gorm.DB, codes []string) ([]Role, error) {
	normalized := make([]string, 0, len(codes))
	for _, code := range codes {
		if trimmed := strings.ToLower(strings.TrimSpace(code)); trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("%w: at least one role is required", ErrInvalidInput)
	}

	var roles []Role
	if err := tx.Where("code IN ?", normalized).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("load roles: %w", err)
	}
	if len(roles) != len(normalized) {
		return nil, fmt.Errorf("%w: %v", ErrUnknownRole, normalized)
	}
	return roles, nil
}

func containsRoleCode(roles []Role, code string) bool {
	for _, role := range roles {
		if role.Code == code {
			return true
		}
	}
	return false
}

// guardLastOwner refuses a change that would leave the tenant with no active
// owner, which would lock everyone out of their own tenant settings.
func guardLastOwner(tx *gorm.DB, tenantID, membershipID uuid.UUID) error {
	var isOwner int64
	if err := tx.Table("membership_roles").
		Joins("JOIN roles ON roles.id = membership_roles.role_id").
		Where("membership_roles.membership_id = ? AND roles.code = ?", membershipID, ownerRoleCode).
		Count(&isOwner).Error; err != nil {
		return fmt.Errorf("check owner role: %w", err)
	}
	if isOwner == 0 {
		return nil
	}

	var remaining int64
	if err := tx.Table("membership_roles").
		Joins("JOIN roles ON roles.id = membership_roles.role_id").
		Joins("JOIN memberships ON memberships.id = membership_roles.membership_id").
		Where(`roles.code = ? AND memberships.tenant_id = ? AND memberships.id <> ?
			AND memberships.status = ?`,
			ownerRoleCode, tenantID, membershipID, MembershipStatusActive).
		Count(&remaining).Error; err != nil {
		return fmt.Errorf("count remaining owners: %w", err)
	}
	if remaining == 0 {
		return ErrLastOwner
	}
	return nil
}
