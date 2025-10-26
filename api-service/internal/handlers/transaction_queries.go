package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// TransactionQueries handles database queries for transactions
type TransactionQueries struct {
	DB *pgxpool.Pool
}

// NewTransactionQueries creates a new TransactionQueries instance
func NewTransactionQueries(db *pgxpool.Pool) *TransactionQueries {
	return &TransactionQueries{DB: db}
}

// GetUserGroupIDs returns group IDs for a user
func (q *TransactionQueries) GetUserGroupIDs(ctx context.Context, userID int64) ([]int64, error) {
	var groupIDs []int64
	rows, err := q.DB.Query(ctx, "SELECT group_id FROM group_members WHERE user_id = $1", userID)
	if err != nil {
		log.Warn().Err(err).Int64("user_id", userID).Msg("failed to query group_members")
		return groupIDs, err
	}
	defer rows.Close()

	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			log.Warn().Err(err).Int64("user_id", userID).Msg("failed to scan group_id")
			continue
		}
		groupIDs = append(groupIDs, groupID)
	}

	if err := rows.Err(); err != nil {
		log.Warn().Err(err).Int64("user_id", userID).Msg("error during group_members iteration")
	}

	return groupIDs, nil
}

// BuildTransactionQuery builds the SQL query for fetching transactions
func (q *TransactionQueries) BuildTransactionQuery(
	whereConditions []string,
	args []interface{},
	argIndex int,
	limit int,
) (string, []interface{}) {
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.user_id, e.amount_cents, e.category_id, e.subcategory_id, 
			   e.operation_type, e.timestamp, e.is_shared, u.username,
			   c.name as category_name, s.name as subcategory_name
		FROM expenses e
		LEFT JOIN users u ON u.telegram_id = e.user_id
		LEFT JOIN categories c ON e.category_id = c.id
		LEFT JOIN subcategories s ON e.subcategory_id = s.id
		%s
		ORDER BY e.timestamp DESC, e.id DESC
		LIMIT $%d
	`, whereClause, argIndex)

	args = append(args, limit+1) // Get one extra to check if there are more
	return query, args
}

// ValidateCategory checks if category exists
func (q *TransactionQueries) ValidateCategory(ctx context.Context, categoryID int) (bool, error) {
	var exists bool
	err := q.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", categoryID).Scan(&exists)
	return exists, err
}

// ValidateSubcategory checks if subcategory exists
func (q *TransactionQueries) ValidateSubcategory(ctx context.Context, subcategoryID int) (bool, error) {
	var exists bool
	err := q.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM subcategories WHERE id = $1)", subcategoryID).Scan(&exists)
	return exists, err
}
