package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"medicfo/app/model"
)

const doctorColumns = `id, user_id, name, specialization, phone, created_at`

type DoctorRepository interface {
	WithTx(tx DBTX) DoctorRepository
	Create(ctx context.Context, d model.Doctor) (model.Doctor, error)
	FindByID(ctx context.Context, id int64) (model.Doctor, error)
	Update(ctx context.Context, id int64, req model.UpdateDoctorRequest) (model.Doctor, error)
	// List mengambil sampai limit baris, terbaru dulu. after == nil berarti halaman pertama.
	List(ctx context.Context, limit int, afterCreatedAt *time.Time, afterID int64) ([]model.Doctor, error)
}

type doctorPostgresRepository struct{ db DBTX }

func NewDoctorRepository(db DBTX) DoctorRepository {
	return &doctorPostgresRepository{db: db}
}

func (r *doctorPostgresRepository) WithTx(tx DBTX) DoctorRepository {
	return &doctorPostgresRepository{db: tx}
}

func scanDoctor(row pgx.Row) (model.Doctor, error) {
	var d model.Doctor
	err := row.Scan(&d.ID, &d.UserID, &d.Name, &d.Specialization, &d.Phone, &d.CreatedAt)
	return d, err
}

func (r *doctorPostgresRepository) Create(ctx context.Context, d model.Doctor) (model.Doctor, error) {
	created, err := scanDoctor(r.db.QueryRow(ctx,
		`INSERT INTO doctors (user_id, name, specialization, phone)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+doctorColumns,
		d.UserID, d.Name, d.Specialization, d.Phone))
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 = unique_violation: user ini sudah punya profil dokter.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Doctor{}, ErrDuplicate
		}
		return model.Doctor{}, fmt.Errorf("menyimpan dokter: %w", err)
	}
	return created, nil
}

func (r *doctorPostgresRepository) FindByID(ctx context.Context, id int64) (model.Doctor, error) {
	d, err := scanDoctor(r.db.QueryRow(ctx,
		`SELECT `+doctorColumns+` FROM doctors WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Doctor{}, ErrNotFound
		}
		return model.Doctor{}, fmt.Errorf("mencari dokter by id: %w", err)
	}
	return d, nil
}

func (r *doctorPostgresRepository) Update(ctx context.Context, id int64, req model.UpdateDoctorRequest) (model.Doctor, error) {
	d, err := scanDoctor(r.db.QueryRow(ctx,
		`UPDATE doctors
		 SET name           = COALESCE($2::varchar, name),
		     specialization = COALESCE($3::varchar, specialization),
		     phone          = COALESCE($4::varchar, phone)
		 WHERE id = $1
		 RETURNING `+doctorColumns,
		id, req.Name, req.Specialization, req.Phone))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Doctor{}, ErrNotFound
		}
		return model.Doctor{}, fmt.Errorf("mengubah dokter: %w", err)
	}
	return d, nil
}

func (r *doctorPostgresRepository) List(ctx context.Context, limit int, afterCreatedAt *time.Time, afterID int64) ([]model.Doctor, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if afterCreatedAt == nil {
		rows, err = r.db.Query(ctx,
			`SELECT `+doctorColumns+` FROM doctors
			 ORDER BY created_at DESC, id DESC
			 LIMIT $1`, limit)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT `+doctorColumns+` FROM doctors
			 WHERE (created_at, id) < ($1, $2)
			 ORDER BY created_at DESC, id DESC
			 LIMIT $3`, *afterCreatedAt, afterID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar dokter: %w", err)
	}
	defer rows.Close()

	doctors := make([]model.Doctor, 0, limit)
	for rows.Next() {
		d, err := scanDoctor(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca baris dokter: %w", err)
		}
		doctors = append(doctors, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi daftar dokter: %w", err)
	}
	return doctors, nil
}
