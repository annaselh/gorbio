package integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annaselh/gorbio/internal/testdb"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/annaselh/gorbio/modules/sales"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// The last-owner guard is two counting queries across a three-table join. It
// cannot be exercised without rows in all three, and getting it wrong locks a
// tenant out of its own administration permanently.

func TestTheLastOwnerCannotBeDemoted(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	auth, err := base.AuthFromApp(app)
	if err != nil {
		t.Fatalf("resolve auth service: %v", err)
	}

	owner := membership(t, app.DB, tenantID, "owner")
	// The acting principal is a different member: demoting yourself is refused
	// earlier, by a different rule, and would not reach the guard.
	actor := membership(t, app.DB, tenantID, "admin")

	_, err = auth.UpdateMemberRoles(ctx, principalFor(tenantID, actor.ID), owner.ID, []string{"member"})
	if !errors.Is(err, base.ErrLastOwner) {
		t.Fatalf("expected ErrLastOwner, got %v", err)
	}
}

func TestAnOwnerCanBeDemotedWhileAnotherRemains(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	auth, _ := base.AuthFromApp(app)

	first := membership(t, app.DB, tenantID, "owner")
	membership(t, app.DB, tenantID, "owner")
	actor := membership(t, app.DB, tenantID, "admin")

	if _, err := auth.UpdateMemberRoles(ctx, principalFor(tenantID, actor.ID), first.ID, []string{"member"}); err != nil {
		t.Fatalf("demoting one of two owners must be allowed: %v", err)
	}
}

func TestAnOwnerInAnotherTenantDoesNotCount(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	mine := testdb.Tenant(t, app.DB)
	theirs := testdb.Tenant(t, app.DB)

	auth, _ := base.AuthFromApp(app)

	owner := membership(t, app.DB, mine, "owner")
	actor := membership(t, app.DB, mine, "admin")
	// Another tenant's owner must not satisfy this tenant's requirement to keep
	// one, or the guard could be defeated by an unrelated signup.
	membership(t, app.DB, theirs, "owner")

	_, err := auth.UpdateMemberRoles(ctx, principalFor(mine, actor.ID), owner.ID, []string{"member"})
	if !errors.Is(err, base.ErrLastOwner) {
		t.Fatalf("expected ErrLastOwner, got %v", err)
	}
}

func TestASuspendedOwnerDoesNotHoldTheTenant(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	auth, _ := base.AuthFromApp(app)

	active := membership(t, app.DB, tenantID, "owner")
	suspended := membership(t, app.DB, tenantID, "owner")
	actor := membership(t, app.DB, tenantID, "admin")

	if err := app.DB.Model(&base.Membership{}).Where("id = ?", suspended.ID).
		Update("status", base.MembershipStatusSuspended).Error; err != nil {
		t.Fatalf("suspend membership: %v", err)
	}

	// A suspended owner cannot administer anything, so they cannot be the
	// reason the last active owner is allowed to step down.
	_, err := auth.UpdateMemberRoles(ctx, principalFor(tenantID, actor.ID), active.ID, []string{"member"})
	if !errors.Is(err, base.ErrLastOwner) {
		t.Fatalf("expected ErrLastOwner, got %v", err)
	}
}

// TestAnOrderWithNoCustomerLinkSurvives is the migration concern stated as the
// thing it protects. CRM added customer_id to a table that already had rows, so
// the column has to be nullable: an order raised before the module existed
// carries no link, and must still read back and still aggregate.
func TestAnOrderWithNoCustomerLinkSurvives(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)
	orders, _ := sales.ServiceFromApp(app)

	order, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
		CustomerName: "Pre-CRM Walk-in",
		OrderDate:    dashboardOrderDate,
		Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 75_000}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if order.CustomerID != nil {
		t.Fatalf("a walk-in must carry no link, got %v", order.CustomerID)
	}

	var customerID *uuid.UUID
	if err := app.DB.Table("sales_orders").
		Where("id = ?", order.ID).Pluck("customer_id", &customerID).Error; err != nil {
		t.Fatalf("read customer_id back: %v", err)
	}
	if customerID != nil {
		t.Fatalf("customer_id should be NULL in the row, got %v", customerID)
	}

	// The dashboard casts this column to text; a NULL has to survive the cast
	// and fall back to the name rather than dropping the order from the count.
	if _, err := orders.UpdateStatus(ctx, tenantID, order.ID, sales.OrderStatusConfirmed); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	summary, err := dashboardService(t, app).Summary(ctx, tenantID, dashboardNow)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Customers.Value != 1 || summary.Revenue.Value != 75_000 {
		t.Fatalf("an unlinked order fell out of the aggregate: %+v", summary)
	}
}

// ------------------------------------------------------------------ helpers

func principalFor(tenantID, membershipID uuid.UUID) base.Principal {
	return base.Principal{
		UserID:       uuid.New(),
		TenantID:     tenantID,
		MembershipID: membershipID,
		Permissions:  map[string]struct{}{"membership.manage": {}},
	}
}

// membership creates an active member of the tenant holding one role.
func membership(t *testing.T, db *gorm.DB, tenantID uuid.UUID, roleCode string) base.Membership {
	t.Helper()

	userID := testdb.User(t, db)
	member := base.Membership{
		ID:       uuid.New(),
		UserID:   userID,
		TenantID: tenantID,
		Status:   base.MembershipStatusActive,
		JoinedAt: time.Now().UTC(),
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("create membership: %v", err)
	}

	var role base.Role
	if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
		t.Fatalf("load %s role: %v", roleCode, err)
	}
	if err := db.Create(&base.MembershipRole{MembershipID: member.ID, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("assign %s role: %v", roleCode, err)
	}
	return member
}
