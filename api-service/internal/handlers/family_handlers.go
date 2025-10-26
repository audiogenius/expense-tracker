package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// FamilyHandlers handles family/group-related endpoints
type FamilyHandlers struct {
	DB   *pgxpool.Pool
	Auth *auth.Auth
}

// NewFamilyHandlers creates a new FamilyHandlers instance
func NewFamilyHandlers(db *pgxpool.Pool, auth *auth.Auth) *FamilyHandlers {
	return &FamilyHandlers{
		DB:   db,
		Auth: auth,
	}
}

type groupResponse struct {
	GroupID   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
	ChatType  string `json:"chat_type"`
}

type groupMemberResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

type familyGroupsResponse struct {
	Groups []groupWithMembers `json:"groups"`
}

type groupWithMembers struct {
	GroupID   int64                 `json:"group_id"`
	GroupName string                `json:"group_name"`
	ChatType  string                `json:"chat_type"`
	Members   []groupMemberResponse `json:"members"`
}

// GetFamilyGroups returns all groups the user is a member of, with member lists
func (h *FamilyHandlers) GetFamilyGroups(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(auth.UserIDKey)
	if uid == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all groups the user is a member of
	groupRows, err := h.DB.Query(r.Context(), `
		SELECT tg.group_id, tg.group_name, tg.chat_type
		FROM telegram_groups tg
		INNER JOIN group_members gm ON tg.group_id = gm.group_id
		WHERE gm.user_id = $1
		ORDER BY tg.group_name
	`, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("failed to query user groups")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer groupRows.Close()

	var groups []groupWithMembers
	for groupRows.Next() {
		var g groupWithMembers
		if err := groupRows.Scan(&g.GroupID, &g.GroupName, &g.ChatType); err != nil {
			log.Warn().Err(err).Msg("failed to scan group row")
			continue
		}

		// Get members for this group
		memberRows, err := h.DB.Query(r.Context(), `
			SELECT u.telegram_id, u.username
			FROM users u
			INNER JOIN group_members gm ON u.telegram_id = gm.user_id
			WHERE gm.group_id = $1
			ORDER BY u.username
		`, g.GroupID)
		if err != nil {
			log.Warn().Err(err).Int64("group_id", g.GroupID).Msg("failed to query group members")
			// Continue with empty members list
			g.Members = []groupMemberResponse{}
		} else {
			defer memberRows.Close()
			for memberRows.Next() {
				var m groupMemberResponse
				if err := memberRows.Scan(&m.UserID, &m.Username); err != nil {
					log.Warn().Err(err).Msg("failed to scan member row")
					continue
				}
				g.Members = append(g.Members, m)
			}
			if err := memberRows.Err(); err != nil {
				log.Warn().Err(err).Int64("group_id", g.GroupID).Msg("error during members iteration")
			}
		}

		groups = append(groups, g)
	}

	if err := groupRows.Err(); err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("error during groups iteration")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := familyGroupsResponse{
		Groups: groups,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("failed to encode family groups response")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Info().Int64("user_id", userID).Int("groups_count", len(groups)).Msg("returned family groups")
}

