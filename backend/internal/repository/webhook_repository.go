package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seymourrisey/payflow-simulator/internal/model"
	"github.com/seymourrisey/payflow-simulator/pkg/idgen"
)

type WebhookRepository struct {
	db *pgxpool.Pool
}

func NewWebhookRepository(db *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// InsertLog — simpan webhook log setelah pengiriman
func (r *WebhookRepository) InsertLog(ctx context.Context, log *model.WebhookLog) error {
	log.ID = idgen.NewWebhookLogID()
	return r.db.QueryRow(ctx, `
		INSERT INTO webhook_logs
			(id, merchant_id, transaction_id, event, payload,
			 response_status, response_body, retry_count, is_delivered)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING sent_at
	`, log.ID, log.MerchantID, log.TransactionID, log.Event,
		log.Payload, log.ResponseStatus, log.ResponseBody,
		log.RetryCount, log.IsDelivered).
		Scan(&log.SentAt)
}

// FindByMerchantID — log berdasarkan merchant
func (r *WebhookRepository) FindByMerchantID(ctx context.Context, merchantID string, limit, offset int) ([]model.WebhookLog, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, merchant_id, transaction_id, event, payload,
		       response_status, response_body, retry_count, is_delivered, sent_at
		FROM webhook_logs
		WHERE merchant_id = $1
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`, merchantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.WebhookLog
	for rows.Next() {
		var l model.WebhookLog
		if err := rows.Scan(
			&l.ID, &l.MerchantID, &l.TransactionID, &l.Event,
			&l.Payload, &l.ResponseStatus, &l.ResponseBody,
			&l.RetryCount, &l.IsDelivered, &l.SentAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM webhook_logs WHERE merchant_id = $1`, merchantID).Scan(&total)
	return logs, total, nil
}

// FindAll — semua log untuk panel global
func (r *WebhookRepository) FindAll(ctx context.Context, limit, offset int) ([]model.WebhookLogWithMerchant, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.id, w.merchant_id, m.merchant_name, w.transaction_id,
		       w.event, w.payload, w.response_status, w.response_body,
		       w.retry_count, w.is_delivered, w.sent_at
		FROM webhook_logs w
		JOIN merchants m ON m.id = w.merchant_id
		ORDER BY w.sent_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.WebhookLogWithMerchant
	for rows.Next() {
		var l model.WebhookLogWithMerchant
		if err := rows.Scan(
			&l.ID, &l.MerchantID, &l.MerchantName, &l.TransactionID,
			&l.Event, &l.Payload, &l.ResponseStatus, &l.ResponseBody,
			&l.RetryCount, &l.IsDelivered, &l.SentAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM webhook_logs`).Scan(&total)
	return logs, total, nil
}

// GetStats — statistik ringkasan webhook
func (r *WebhookRepository) GetStats(ctx context.Context) (*model.WebhookStats, error) {
	stats := &model.WebhookStats{}
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*)                                        AS total,
			COUNT(*) FILTER (WHERE is_delivered = TRUE)    AS delivered,
			COUNT(*) FILTER (WHERE is_delivered = FALSE)   AS failed,
			COUNT(*) FILTER (WHERE sent_at >= NOW() - INTERVAL '1 hour') AS last_hour
		FROM webhook_logs
	`).Scan(&stats.Total, &stats.Delivered, &stats.Failed, &stats.LastHour)
	if err != nil {
		return nil, err
	}
	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Delivered) / float64(stats.Total) * 100
	}
	return stats, nil
}

// FindMerchantWebhookURL — ambil webhook URL dari merchant
func (r *WebhookRepository) FindMerchantWebhookURL(ctx context.Context, merchantID string) (string, error) {
	var url string
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(webhook_url, '') FROM merchants WHERE id = $1
	`, merchantID).Scan(&url)
	return url, err
}

// GetAllMerchants — untuk dropdown di panel
func (r *WebhookRepository) GetAllMerchants(ctx context.Context) ([]model.Merchant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, merchant_name, COALESCE(webhook_url, ''), created_at FROM merchants ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var merchants []model.Merchant
	for rows.Next() {
		var m model.Merchant
		if err := rows.Scan(&m.ID, &m.MerchantName, &m.WebhookURL, &m.CreatedAt); err != nil {
			return nil, err
		}
		merchants = append(merchants, m)
	}
	return merchants, nil
}

// MarkDelivered — update status log setelah retry berhasil
func (r *WebhookRepository) MarkDelivered(ctx context.Context, logID string, responseStatus int, responseBody string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE webhook_logs
		SET is_delivered = TRUE, response_status = $1, response_body = $2
		WHERE id = $3
	`, responseStatus, responseBody, logID)
	return err
}

// GetRecentByTransaction — cek apakah webhook sudah pernah dikirim
func (r *WebhookRepository) GetRecentByTransaction(ctx context.Context, txID string) (*model.WebhookLog, error) {
	l := &model.WebhookLog{}
	err := r.db.QueryRow(ctx, `
		SELECT id, merchant_id, transaction_id, event, is_delivered, retry_count, sent_at
		FROM webhook_logs WHERE transaction_id = $1 ORDER BY sent_at DESC LIMIT 1
	`, txID).Scan(&l.ID, &l.MerchantID, &l.TransactionID, &l.Event,
		&l.IsDelivered, &l.RetryCount, &l.SentAt)
	if err != nil {
		return nil, err
	}
	return l, nil
}

var _ = time.Now
