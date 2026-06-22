package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jul2264/Flock/backend/internal/models"
)

type CommunityService struct {
	db     *sql.DB
	search *SearchService
}

func NewCommunityService(db *sql.DB, search *SearchService) *CommunityService {
	return &CommunityService{db: db, search: search}
}

func (s *CommunityService) Create(ownerClerkID string, req *models.CreateCommunityRequest) (*models.Community, error) {
	var ownerID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, ownerClerkID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("owner not found")
		}
		return nil, err
	}

	visibility := "public"
	if req.Visibility != nil {
		visibility = *req.Visibility
	}

	isRecurring := false
	if req.IsRecurring != nil {
		isRecurring = *req.IsRecurring
	}

	query := `
		INSERT INTO communities (
			owner_id, name, description, city, latitude, longitude,
			age_min, age_max, max_members, is_recurring, recurrence_rule, visibility
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12
		)
		RETURNING
			id, owner_id, name, description, city, latitude, longitude,
			age_min, age_max, max_members, is_recurring, recurrence_rule,
			visibility, member_count, created_at, updated_at
	`
	row := s.db.QueryRow(query,
		ownerID, req.Name, req.Description, req.City, req.Latitude, req.Longitude,
		req.AgeMin, req.AgeMax, req.MaxMembers, isRecurring, req.RecurrenceRule, visibility,
	)
	community, err := scanCommunity(row)
	if err != nil {
		return nil, err
	}

	if s.search != nil {
		go func(c *models.Community) {
			if err := s.search.SyncCommunity(c); err != nil {
				log.Printf("Error syncing community to Meilisearch: %v", err)
			}
		}(community)
	}

	return community, nil
}

func (s *CommunityService) GetByID(id string) (*models.Community, error) {
	query := `
		SELECT
			id, owner_id, name, description, city, latitude, longitude,
			age_min, age_max, max_members, is_recurring, recurrence_rule,
			visibility, member_count, created_at, updated_at
		FROM communities
		WHERE id = $1
	`
	row := s.db.QueryRow(query, id)
	return scanCommunity(row)
}

func (s *CommunityService) List() ([]models.Community, error) {
	query := `
		SELECT
			id, owner_id, name, description, city, latitude, longitude,
			age_min, age_max, max_members, is_recurring, recurrence_rule,
			visibility, member_count, created_at, updated_at
		FROM communities
		WHERE visibility = 'public'
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var communities []models.Community
	for rows.Next() {
		var c models.Community
		err := rows.Scan(
			&c.ID, &c.OwnerID, &c.Name, &c.Description, &c.City, &c.Latitude, &c.Longitude,
			&c.AgeMin, &c.AgeMax, &c.MaxMembers, &c.IsRecurring, &c.RecurrenceRule,
			&c.Visibility, &c.MemberCount, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		communities = append(communities, c)
	}
	return communities, nil
}

func scanCommunity(row *sql.Row) (*models.Community, error) {
	var c models.Community
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.Name, &c.Description, &c.City, &c.Latitude, &c.Longitude,
		&c.AgeMin, &c.AgeMax, &c.MaxMembers, &c.IsRecurring, &c.RecurrenceRule,
		&c.Visibility, &c.MemberCount, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
