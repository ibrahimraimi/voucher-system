package voucher

import (
	"database/sql"
	"fmt"
	"time"
)

type Repository interface {
	Save(v *Voucher) error
	SaveBatch(vouchers []*Voucher) error
	FindByHash(hash string) (*Voucher, error)
	UpdateStatus(id string, status VoucherStatus) error
	Redeem(hash string, user string) (int, error)
}

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Redeem(hash string, user string) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Check current status and value
	var value int
	var status string
	err = tx.QueryRow("SELECT value, status FROM vouchers WHERE code_hash = ?", hash).Scan(&value, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("voucher not found")
		}
		return 0, fmt.Errorf("failed to query voucher: %w", err)
	}

	if VoucherStatus(status) != StatusUnused {
		return 0, fmt.Errorf("voucher is %s", status)
	}

	// 2. Mark as used
	res, err := tx.Exec("UPDATE vouchers SET status = ?, redeemed_at = ?, redeemed_by = ? WHERE code_hash = ? AND status = ?",
		StatusUsed, time.Now(), user, hash, StatusUnused)
	if err != nil {
		return 0, fmt.Errorf("failed to update voucher: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return 0, fmt.Errorf("concurrent redemption attempt detected")
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return value, nil
}

func (r *SQLiteRepository) SaveBatch(vouchers []*Voucher) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO vouchers (id, code_hash, value, status, batch_id, created_at) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range vouchers {
		_, err := stmt.Exec(v.ID, v.CodeHash, v.Value, v.Status, v.BatchID, v.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) Save(v *Voucher) error {
	query := `
		INSERT INTO vouchers (id, code_hash, value, status, batch_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, v.ID, v.CodeHash, v.Value, v.Status, v.BatchID, v.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save voucher: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) FindByHash(hash string) (*Voucher, error) {
	query := `
		SELECT id, code_hash, value, status, batch_id, created_at, redeemed_at, redeemed_by
		FROM vouchers
		WHERE code_hash = ?
	`
	row := r.db.QueryRow(query, hash)

	v := &Voucher{}
	var status string
	err := row.Scan(
		&v.ID,
		&v.CodeHash,
		&v.Value,
		&status,
		&v.BatchID,
		&v.CreatedAt,
		&v.RedeemedAt,
		&v.RedeemedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find voucher: %w", err)
	}
	v.Status = VoucherStatus(status)
	return v, nil
}

func (r *SQLiteRepository) UpdateStatus(id string, status VoucherStatus) error {
	query := `UPDATE vouchers SET status = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}
