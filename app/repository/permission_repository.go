package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepository interface {
	ListRolePermissions(ctx context.Context) (map[string][]string, error)
}

type permissionPostgresRepository struct{ pool *pgxpool.Pool }

func NewPermissionRepository(pool *pgxpool.Pool) PermissionRepository {
	return &permissionPostgresRepository{pool: pool}
}

func (r *permissionPostgresRepository) ListRolePermissions(ctx context.Context) (map[string][]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT role_name, permission_name FROM role_permissions ORDER BY role_name, permission_name`)
	if err != nil {
		return nil, fmt.Errorf("memuat role permissions: %w", err)
	}
	defer rows.Close()
	result := map[string][]string{}
	for rows.Next() {
		var role, permission string
		if err := rows.Scan(&role, &permission); err != nil {
			return nil, fmt.Errorf("membaca role permissions: %w", err)
		}
		result[role] = append(result[role], permission)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi role permissions: %w", err)
	}
	return result, nil
}
