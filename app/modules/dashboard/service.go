// Package dashboard aggregates figures that span several business modules.
//
// It depends on sales, inventory and procurement rather than the other way
// round. Putting these queries in base would invert the dependency - base is
// what every module depends on - so the aggregation lives in a module of its
// own that sits above them all.
package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const ServiceName = "dashboard.metrics"

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Period is a half-open interval [Start, End).
type Period struct {
	Start time.Time
	End   time.Time
}

// MonthToDate returns the current month and the same span of the month before,
// which is what a "vs last month" delta compares against.
func MonthToDate(now time.Time) (current, previous Period) {
	now = now.UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	current = Period{Start: startOfMonth, End: now}

	previousStart := startOfMonth.AddDate(0, -1, 0)
	// Compare like with like: the same number of elapsed days, so a comparison
	// made on the 3rd does not measure a full month against three days.
	previousEnd := previousStart.Add(now.Sub(startOfMonth))
	if limit := startOfMonth; previousEnd.After(limit) {
		previousEnd = limit
	}
	previous = Period{Start: previousStart, End: previousEnd}
	return current, previous
}

// Metric is one KPI: a figure, the same figure for the comparison period, and
// the percentage change between them.
type Metric struct {
	Value    int64   `json:"value"`
	Previous int64   `json:"previous"`
	Delta    float64 `json:"delta"`
}

func newMetric(current, previous int64) Metric {
	return Metric{Value: current, Previous: previous, Delta: percentChange(current, previous)}
}

// percentChange returns 0 when there is no baseline. Reporting an infinite or
// 100% jump from zero would read as real growth rather than "no prior data".
func percentChange(current, previous int64) float64 {
	if previous == 0 {
		return 0
	}
	return (float64(current-previous) / float64(previous)) * 100
}

type Summary struct {
	Revenue     Metric `json:"revenue"`
	Orders      Metric `json:"orders"`
	Customers   Metric `json:"customers"`
	Purchases   Metric `json:"purchases"`
	GrossMargin Metric `json:"gross_margin"`
	Period      struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"period"`
}

// Summary computes the KPI row.
//
// Only Confirmed sales orders count as revenue and only Confirmed or Received
// purchase orders count as spend: a draft is a proposal, and a cancelled order
// never happened. Gross margin here is revenue minus purchase spend, which is
// the closest honest figure available without cost-of-goods tracking.
func (s *Service) Summary(ctx context.Context, tenantID uuid.UUID, now time.Time) (*Summary, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant scope is required")
	}
	current, previous := MonthToDate(now)

	revenueNow, ordersNow, customersNow, err := s.salesTotals(ctx, tenantID, current)
	if err != nil {
		return nil, err
	}
	revenuePrev, ordersPrev, customersPrev, err := s.salesTotals(ctx, tenantID, previous)
	if err != nil {
		return nil, err
	}
	purchasesNow, err := s.purchaseTotal(ctx, tenantID, current)
	if err != nil {
		return nil, err
	}
	purchasesPrev, err := s.purchaseTotal(ctx, tenantID, previous)
	if err != nil {
		return nil, err
	}

	summary := &Summary{
		Revenue:     newMetric(revenueNow, revenuePrev),
		Orders:      newMetric(ordersNow, ordersPrev),
		Customers:   newMetric(customersNow, customersPrev),
		Purchases:   newMetric(purchasesNow, purchasesPrev),
		GrossMargin: newMetric(revenueNow-purchasesNow, revenuePrev-purchasesPrev),
	}
	summary.Period.Start = current.Start
	summary.Period.End = current.End
	return summary, nil
}

func (s *Service) salesTotals(ctx context.Context, tenantID uuid.UUID, period Period) (revenue, orders, customers int64, err error) {
	var row struct {
		Revenue   int64
		Orders    int64
		Customers int64
	}
	err = s.db.WithContext(ctx).
		Table("sales_orders").
		Select(`COALESCE(SUM(total), 0) AS revenue,
			COUNT(*) AS orders,
			COUNT(DISTINCT customer_name) AS customers`).
		Where(`tenant_id = ? AND status = 'Confirmed' AND deleted_at IS NULL
			AND order_date >= ? AND order_date < ?`, tenantID, period.Start, period.End).
		Scan(&row).Error
	if err != nil {
		return 0, 0, 0, fmt.Errorf("aggregate sales: %w", err)
	}
	return row.Revenue, row.Orders, row.Customers, nil
}

func (s *Service) purchaseTotal(ctx context.Context, tenantID uuid.UUID, period Period) (int64, error) {
	var total int64
	err := s.db.WithContext(ctx).
		Table("procurement_purchase_orders").
		Select("COALESCE(SUM(total), 0)").
		Where(`tenant_id = ? AND status IN ('Confirmed', 'Received') AND deleted_at IS NULL
			AND order_date >= ? AND order_date < ?`, tenantID, period.Start, period.End).
		Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("aggregate purchases: %w", err)
	}
	return total, nil
}

// SeriesPoint is one day across every headline figure, so a single request
// backs all four sparklines instead of one round trip per card.
type SeriesPoint struct {
	Date      string `json:"date"`
	Revenue   int64  `json:"revenue"`
	Orders    int64  `json:"orders"`
	Purchases int64  `json:"purchases"`
}

// SalesSeries returns daily figures for the current and previous month. Days
// with no activity are emitted as zero so a chart draws a continuous line
// rather than silently closing the gap between two distant dates.
func (s *Service) SalesSeries(ctx context.Context, tenantID uuid.UUID, now time.Time) (currentSeries, previousSeries []SeriesPoint, err error) {
	if tenantID == uuid.Nil {
		return nil, nil, fmt.Errorf("tenant scope is required")
	}

	now = now.UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfPrevious := startOfMonth.AddDate(0, -1, 0)

	currentSeries, err = s.dailySeries(ctx, tenantID, startOfMonth, startOfMonth.AddDate(0, 1, 0))
	if err != nil {
		return nil, nil, err
	}
	previousSeries, err = s.dailySeries(ctx, tenantID, startOfPrevious, startOfMonth)
	if err != nil {
		return nil, nil, err
	}
	return currentSeries, previousSeries, nil
}

func (s *Service) dailySeries(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]SeriesPoint, error) {
	type dayRow struct {
		Day     time.Time
		Value   int64
		Records int64
	}

	var salesRows []dayRow
	err := s.db.WithContext(ctx).
		Table("sales_orders").
		Select(`DATE_TRUNC('day', order_date) AS day,
			COALESCE(SUM(total), 0) AS value, COUNT(*) AS records`).
		Where(`tenant_id = ? AND status = 'Confirmed' AND deleted_at IS NULL
			AND order_date >= ? AND order_date < ?`, tenantID, start, end).
		Group("DATE_TRUNC('day', order_date)").
		Scan(&salesRows).Error
	if err != nil {
		return nil, fmt.Errorf("daily sales: %w", err)
	}

	var purchaseRows []dayRow
	err = s.db.WithContext(ctx).
		Table("procurement_purchase_orders").
		Select(`DATE_TRUNC('day', order_date) AS day,
			COALESCE(SUM(total), 0) AS value, COUNT(*) AS records`).
		Where(`tenant_id = ? AND status IN ('Confirmed', 'Received') AND deleted_at IS NULL
			AND order_date >= ? AND order_date < ?`, tenantID, start, end).
		Group("DATE_TRUNC('day', order_date)").
		Scan(&purchaseRows).Error
	if err != nil {
		return nil, fmt.Errorf("daily purchases: %w", err)
	}

	revenueByDay := make(map[string]int64, len(salesRows))
	ordersByDay := make(map[string]int64, len(salesRows))
	for _, r := range salesRows {
		key := r.Day.UTC().Format("2006-01-02")
		revenueByDay[key] = r.Value
		ordersByDay[key] = r.Records
	}
	purchasesByDay := make(map[string]int64, len(purchaseRows))
	for _, r := range purchaseRows {
		purchasesByDay[r.Day.UTC().Format("2006-01-02")] = r.Value
	}

	points := make([]SeriesPoint, 0, 31)
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		points = append(points, SeriesPoint{
			Date:      key,
			Revenue:   revenueByDay[key],
			Orders:    ordersByDay[key],
			Purchases: purchasesByDay[key],
		})
	}
	return points, nil
}

// TopProduct is one row of the best-sellers table.
type TopProduct struct {
	SKU         string  `json:"sku"`
	Description string  `json:"description"`
	Revenue     int64   `json:"revenue"`
	Quantity    int64   `json:"quantity"`
	Delta       float64 `json:"delta"`
}

// TopProducts ranks SKUs by confirmed revenue this month, with the change
// against the same span last month.
func (s *Service) TopProducts(ctx context.Context, tenantID uuid.UUID, now time.Time, limit int) ([]TopProduct, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant scope is required")
	}
	if limit <= 0 || limit > 50 {
		limit = 5
	}
	current, previous := MonthToDate(now)

	rows, err := s.productRevenue(ctx, tenantID, current, limit)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []TopProduct{}, nil
	}

	// One extra query for the comparison period rather than one per SKU.
	priorBySKU, err := s.productRevenueBySKU(ctx, tenantID, previous)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Delta = percentChange(rows[i].Revenue, priorBySKU[rows[i].SKU])
	}
	return rows, nil
}

func (s *Service) productRevenue(ctx context.Context, tenantID uuid.UUID, period Period, limit int) ([]TopProduct, error) {
	var rows []TopProduct
	err := s.db.WithContext(ctx).
		Table("sales_order_lines").
		Select(`sales_order_lines.sku,
			MAX(sales_order_lines.description) AS description,
			COALESCE(SUM(sales_order_lines.line_total), 0) AS revenue,
			COALESCE(SUM(sales_order_lines.quantity), 0) AS quantity`).
		Joins("JOIN sales_orders ON sales_orders.id = sales_order_lines.order_id").
		Where(`sales_order_lines.tenant_id = ? AND sales_orders.status = 'Confirmed'
			AND sales_orders.deleted_at IS NULL
			AND sales_orders.order_date >= ? AND sales_orders.order_date < ?`,
			tenantID, period.Start, period.End).
		Group("sales_order_lines.sku").
		Order("revenue DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("rank products: %w", err)
	}
	return rows, nil
}

func (s *Service) productRevenueBySKU(ctx context.Context, tenantID uuid.UUID, period Period) (map[string]int64, error) {
	type row struct {
		SKU     string
		Revenue int64
	}
	var rows []row
	err := s.db.WithContext(ctx).
		Table("sales_order_lines").
		Select("sales_order_lines.sku, COALESCE(SUM(sales_order_lines.line_total), 0) AS revenue").
		Joins("JOIN sales_orders ON sales_orders.id = sales_order_lines.order_id").
		Where(`sales_order_lines.tenant_id = ? AND sales_orders.status = 'Confirmed'
			AND sales_orders.deleted_at IS NULL
			AND sales_orders.order_date >= ? AND sales_orders.order_date < ?`,
			tenantID, period.Start, period.End).
		Group("sales_order_lines.sku").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("prior product revenue: %w", err)
	}

	result := make(map[string]int64, len(rows))
	for _, r := range rows {
		result[r.SKU] = r.Revenue
	}
	return result, nil
}

// CashFlowSlice is one segment of the income/expense breakdown.
type CashFlowSlice struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// CashFlow reports confirmed sales revenue against confirmed purchase spend.
//
// It is deliberately two slices, not three: the mockup's "Other" category has
// no source in the data model, and inventing one would put a fabricated figure
// on a finance chart.
func (s *Service) CashFlow(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]CashFlowSlice, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant scope is required")
	}
	current, _ := MonthToDate(now)

	revenue, _, _, err := s.salesTotals(ctx, tenantID, current)
	if err != nil {
		return nil, err
	}
	purchases, err := s.purchaseTotal(ctx, tenantID, current)
	if err != nil {
		return nil, err
	}

	return []CashFlowSlice{
		{Name: "Income", Value: revenue},
		{Name: "Expense", Value: purchases},
	}, nil
}
