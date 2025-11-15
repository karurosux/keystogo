package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/karurosux/keystogo/pkg/keystogo"
	"github.com/karurosux/keystogo/pkg/models"
)

type PostgresConfig struct {
	TableName     string
	SchemaName    string
	ColumnMapping PostgresColumnMapping
}

type PostgresColumnMapping struct {
	ID          string
	Key         string
	Name        string
	Permissions string
	Metadata    string
	CreatedAt   string
	ExpiresAt   string
	LastUsedAt  string
	Active      string
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		TableName:  "api_keys",
		SchemaName: "public",
		ColumnMapping: PostgresColumnMapping{
			ID:          "id",
			Key:         "key",
			Name:        "name",
			Permissions: "permissions",
			Metadata:    "metadata",
			CreatedAt:   "created_at",
			ExpiresAt:   "expires_at",
			LastUsedAt:  "last_used_at",
			Active:      "active",
		},
	}
}

type PostgresStorage struct {
	db     *sql.DB
	config PostgresConfig
}

func NewPostgresStorage(db *sql.DB, config *PostgresConfig) keystogo.Storage {
	if config == nil {
		defaultConfig := DefaultPostgresConfig()
		config = &defaultConfig
	}

	return &PostgresStorage{
		db:     db,
		config: *config,
	}
}

func (p *PostgresStorage) tableName() string {
	if p.config.SchemaName != "" {
		return fmt.Sprintf("%s.%s", p.config.SchemaName, p.config.TableName)
	}
	return p.config.TableName
}

func (p *PostgresStorage) GetByID(ctx context.Context, id string) (*models.APIKey, error) {
	cm := p.config.ColumnMapping
	query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s
		FROM %s
		WHERE %s = $1
	`, cm.ID, cm.Key, cm.Name, cm.Permissions, cm.Metadata, cm.CreatedAt, cm.ExpiresAt, cm.LastUsedAt, cm.Active, p.tableName(), cm.ID)

	return p.scanAPIKey(p.db.QueryRowContext(ctx, query, id))
}

func (p *PostgresStorage) GetByHashedKey(ctx context.Context, hashedKey string) (*models.APIKey, error) {
	cm := p.config.ColumnMapping
	query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s
		FROM %s
		WHERE %s = $1
	`, cm.ID, cm.Key, cm.Name, cm.Permissions, cm.Metadata, cm.CreatedAt, cm.ExpiresAt, cm.LastUsedAt, cm.Active, p.tableName(), cm.Key)

	return p.scanAPIKey(p.db.QueryRowContext(ctx, query, hashedKey))
}

func (p *PostgresStorage) Create(ctx context.Context, apiKey *models.APIKey) error {
	cm := p.config.ColumnMapping

	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	metadataJSON, err := json.Marshal(apiKey.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s, %s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, p.tableName(), cm.ID, cm.Key, cm.Name, cm.Permissions, cm.Metadata, cm.CreatedAt, cm.ExpiresAt, cm.LastUsedAt, cm.Active)

	_, err = p.db.ExecContext(ctx, query,
		apiKey.ID,
		apiKey.Key,
		apiKey.Name,
		permissionsJSON,
		metadataJSON,
		apiKey.CreatedAt,
		apiKey.ExpiresAt,
		apiKey.LastUsedAt,
		apiKey.Active,
	)

	return err
}

func (p *PostgresStorage) Update(ctx context.Context, id string, update models.ApiKeyUpdate) error {
	cm := p.config.ColumnMapping

	existing, err := p.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if update.Active != nil {
		existing.Active = *update.Active
	}
	if update.Name != nil {
		existing.Name = *update.Name
	}
	if update.ExpiresAt != nil {
		existing.ExpiresAt = update.ExpiresAt
	}
	if update.Metadata != nil {
		existing.Metadata = update.Metadata
	}
	if update.Permissions != nil {
		existing.Permissions = update.Permissions
	}
	if update.LastUsedAt != nil {
		existing.LastUsedAt = update.LastUsedAt
	}

	permissionsJSON, err := json.Marshal(existing.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	metadataJSON, err := json.Marshal(existing.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s = $1, %s = $2, %s = $3, %s = $4, %s = $5, %s = $6
		WHERE %s = $7
	`, p.tableName(), cm.Name, cm.Permissions, cm.Metadata, cm.ExpiresAt, cm.LastUsedAt, cm.Active, cm.ID)

	_, err = p.db.ExecContext(ctx, query,
		existing.Name,
		permissionsJSON,
		metadataJSON,
		existing.ExpiresAt,
		existing.LastUsedAt,
		existing.Active,
		id,
	)

	return err
}

func (p *PostgresStorage) Delete(ctx context.Context, id string) error {
	cm := p.config.ColumnMapping
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, p.tableName(), cm.ID)
	_, err := p.db.ExecContext(ctx, query, id)
	return err
}

func (p *PostgresStorage) List(ctx context.Context, page models.Page, filter models.Filter) ([]models.APIKey, int64, error) {
	cm := p.config.ColumnMapping

	whereClause := fmt.Sprintf(" WHERE (%s IS NULL OR %s > NOW())", cm.ExpiresAt, cm.ExpiresAt)
	args := []interface{}{}
	argCount := 1

	if filter.Name != nil && *filter.Name != "" {
		whereClause += fmt.Sprintf(" AND %s ILIKE $%d", cm.Name, argCount)
		args = append(args, "%"+*filter.Name+"%")
		argCount++
	}

	if filter.Active != nil {
		whereClause += fmt.Sprintf(" AND %s = $%d", cm.Active, argCount)
		args = append(args, *filter.Active)
		argCount++
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s%s`, p.tableName(), whereClause)
	var total int64
	err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s
		FROM %s%s
		ORDER BY %s DESC
		LIMIT $%d OFFSET $%d
	`, cm.ID, cm.Key, cm.Name, cm.Permissions, cm.Metadata, cm.CreatedAt, cm.ExpiresAt, cm.LastUsedAt, cm.Active,
		p.tableName(), whereClause, cm.CreatedAt, argCount, argCount+1)

	limit := page.Limit
	if limit == 0 {
		limit = 100
	}

	args = append(args, limit, page.Offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := make([]models.APIKey, 0, limit)
	for rows.Next() {
		apiKey, err := p.scanAPIKeyFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *apiKey)
	}

	return result, total, rows.Err()
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresStorage) Clear(ctx context.Context) error {
	query := fmt.Sprintf(`DELETE FROM %s`, p.tableName())
	_, err := p.db.ExecContext(ctx, query)
	return err
}

func (p *PostgresStorage) CleanupExpired(ctx context.Context) (int64, error) {
	cm := p.config.ColumnMapping
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s IS NOT NULL AND %s <= NOW()`, p.tableName(), cm.ExpiresAt, cm.ExpiresAt)

	result, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (p *PostgresStorage) scanAPIKey(row *sql.Row) (*models.APIKey, error) {
	var apiKey models.APIKey
	var permissionsJSON, metadataJSON []byte

	err := row.Scan(
		&apiKey.ID,
		&apiKey.Key,
		&apiKey.Name,
		&permissionsJSON,
		&metadataJSON,
		&apiKey.CreatedAt,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.Active,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrKeyNotFound()
	}
	if err != nil {
		return nil, err
	}

	if permissionsJSON != nil {
		if err := json.Unmarshal(permissionsJSON, &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &apiKey.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &apiKey, nil
}

func (p *PostgresStorage) scanAPIKeyFromRows(rows *sql.Rows) (*models.APIKey, error) {
	var apiKey models.APIKey
	var permissionsJSON, metadataJSON []byte

	err := rows.Scan(
		&apiKey.ID,
		&apiKey.Key,
		&apiKey.Name,
		&permissionsJSON,
		&metadataJSON,
		&apiKey.CreatedAt,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.Active,
	)

	if err != nil {
		return nil, err
	}

	if permissionsJSON != nil {
		if err := json.Unmarshal(permissionsJSON, &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &apiKey.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &apiKey, nil
}
