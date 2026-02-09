package repositories

import (
	"context"
	"database/sql"
	"kasir-api/models"
	"time"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// FetchReport menghitung total_revenue, total_transaksi, dan produk_terlaris
// pada rentang [start, end) — end eksklusif.
func (r *ReportRepository) FetchReport(ctx context.Context, start, end time.Time) (models.ReportResponse, error) {
	// Tambahkan timeout supaya query tidak hang
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var resp models.ReportResponse

	// 1) Total revenue
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(t.total_amount), 0)
         FROM transactions t
         WHERE t.created_at >= $1 AND t.created_at < $2`,
		start, end,
	).Scan(&resp.TotalRevenue); err != nil {
		return resp, err
	}

	// 2) Total transaksi
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*)
         FROM transactions t
         WHERE t.created_at >= $1 AND t.created_at < $2`,
		start, end,
	).Scan(&resp.TotalTransaksi); err != nil {
		return resp, err
	}

	// 3) Produk terlaris
	var nama sql.NullString
	var qty sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT p.name AS nama, SUM(d.quantity) AS qty_terlaris
         FROM transaction_details d
         JOIN transactions t ON t.id = d.transaction_id
         JOIN products p ON p.id = d.product_id
         WHERE t.created_at >= $1 AND t.created_at < $2
         GROUP BY p.name
         ORDER BY qty_terlaris DESC
         LIMIT 1`,
		start, end,
	).Scan(&nama, &qty)

	if err != nil {
		if err == sql.ErrNoRows {
			resp.ProdukTerlaris = nil
		} else {
			return resp, err
		}
	} else if nama.Valid && qty.Valid {
		resp.ProdukTerlaris = &models.BestSeller{
			Nama:       nama.String,
			QtyTerjual: int(qty.Int64),
		}
	}

	return resp, nil
}
