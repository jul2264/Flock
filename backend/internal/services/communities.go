package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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
			age_min, age_max, max_members, is_recurring, recurrence_rule, visibility, image_url
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12, $13
		)
		RETURNING
			id, owner_id, name, description, city, latitude, longitude,
			age_min, age_max, max_members, is_recurring, recurrence_rule,
			visibility, member_count, status, image_url, created_at, updated_at
	`
	row := s.db.QueryRow(query,
		ownerID, req.Name, req.Description, req.City, req.Latitude, req.Longitude,
		req.AgeMin, req.AgeMax, req.MaxMembers, isRecurring, req.RecurrenceRule, visibility, req.ImageURL,
	)
	community, err := scanCommunity(row)
	if err != nil {
		return nil, err
	}

	// Make the creator a member (owner role) of the community automatically
	_, err = s.db.Exec(`
		INSERT INTO community_members (community_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, community.ID, ownerID)
	if err != nil {
		log.Printf("Warning: failed to add community creator as member: %v", err)
	} else {
		// Increment member count for the admin/creator
		_, _ = s.db.Exec(`UPDATE communities SET member_count = member_count + 1 WHERE id = $1`, community.ID)
		community.MemberCount = 1
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
			visibility, member_count, status, image_url, created_at, updated_at
		FROM communities
		WHERE id = $1
	`
	row := s.db.QueryRow(query, id)
	return scanCommunity(row)
}

func (s *CommunityService) List(filter *models.CommunityFilter) ([]models.Community, error) {
	var selectCols = `id, owner_id, name, description, city, latitude, longitude,
		age_min, age_max, max_members, is_recurring, recurrence_rule,
		visibility, member_count, status, image_url, created_at, updated_at`

	args := []interface{}{}
	idx := 1

	// Base conditions
	whereClauses := []string{"visibility = 'public'", "status = 'active'"}

	// Geo-filtering
	var selectQuery string
	if filter.Lat != nil && filter.Lng != nil && filter.RadiusKm != nil {
		selectQuery = fmt.Sprintf(`
			SELECT %s FROM (
				SELECT *, (6371 * acos(
					cos(radians($%d)) * cos(radians(latitude)) * cos(radians(longitude) - radians($%d)) +
					sin(radians($%d)) * sin(radians(latitude))
				)) AS distance
				FROM communities
			) sub
		`, selectCols, idx, idx+1, idx)
		args = append(args, *filter.Lat, *filter.Lng)
		idx += 2

		whereClauses = append(whereClauses, fmt.Sprintf("distance <= $%d", idx))
		args = append(args, *filter.RadiusKm)
		idx++

		whereClauses = append(whereClauses, "latitude IS NOT NULL", "longitude IS NOT NULL")
	} else {
		selectQuery = fmt.Sprintf("SELECT %s FROM communities", selectCols)
	}

	// Age filters
	if filter.AgeMin != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("age_min >= $%d", idx))
		args = append(args, *filter.AgeMin)
		idx++
	}
	if filter.AgeMax != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("age_max <= $%d", idx))
		args = append(args, *filter.AgeMax)
		idx++
	}

	// Construct WHERE clause
	var whereClause string
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Sort logic
	var orderClause string
	switch filter.Sort {
	case "distance":
		if filter.Lat != nil && filter.Lng != nil && filter.RadiusKm != nil {
			orderClause = " ORDER BY distance ASC"
		} else {
			orderClause = " ORDER BY created_at DESC"
		}
	case "trending":
		orderClause = " ORDER BY member_count DESC, created_at DESC"
	case "newest":
		fallthrough
	default:
		orderClause = " ORDER BY created_at DESC"
	}

	// Pagination limits
	limit := 20
	if filter.Limit > 0 {
		limit = filter.Limit
		if limit > 50 {
			limit = 50
		}
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	query := selectQuery + whereClause + orderClause + paginationClause

	rows, err := s.db.Query(query, args...)
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
			&c.Visibility, &c.MemberCount, &c.Status, &c.ImageURL, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		communities = append(communities, c)
	}

	return communities, nil
}

// GetCommunityOwner fetches the owner's database UUID for the given community ID.
func (s *CommunityService) GetCommunityOwner(communityID string) (string, error) {
	var ownerID string
	err := s.db.QueryRow(`SELECT owner_id FROM communities WHERE id = $1`, communityID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("community not found")
		}
		return "", err
	}
	return ownerID, nil
}

func (s *CommunityService) Update(communityID string, ownerClerkID string, userRole string, req *models.UpdateCommunityRequest) (*models.Community, error) {
	// 1. Get user UUID from Clerk ID
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, ownerClerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// 2. Fetch the community owner to check ownership
	ownerID, err := s.GetCommunityOwner(communityID)
	if err != nil {
		return nil, err
	}

	// 3. Ownership check: owner or admin
	if ownerID != userID && userRole != "admin" {
		return nil, fmt.Errorf("forbidden")
	}

	// 4. Construct dynamic UPDATE query
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	idx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", idx))
		args = append(args, *req.Name)
		idx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", idx))
		args = append(args, req.Description)
		idx++
	}
	if req.City != nil {
		setClauses = append(setClauses, fmt.Sprintf("city = $%d", idx))
		args = append(args, req.City)
		idx++
	}
	if req.Latitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("latitude = $%d", idx))
		args = append(args, req.Latitude)
		idx++
	}
	if req.Longitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("longitude = $%d", idx))
		args = append(args, req.Longitude)
		idx++
	}
	if req.AgeMin != nil {
		setClauses = append(setClauses, fmt.Sprintf("age_min = $%d", idx))
		args = append(args, req.AgeMin)
		idx++
	}
	if req.AgeMax != nil {
		setClauses = append(setClauses, fmt.Sprintf("age_max = $%d", idx))
		args = append(args, req.AgeMax)
		idx++
	}
	if req.MaxMembers != nil {
		setClauses = append(setClauses, fmt.Sprintf("max_members = $%d", idx))
		args = append(args, req.MaxMembers)
		idx++
	}
	if req.IsRecurring != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_recurring = $%d", idx))
		args = append(args, *req.IsRecurring)
		idx++
	}
	if req.RecurrenceRule != nil {
		setClauses = append(setClauses, fmt.Sprintf("recurrence_rule = $%d", idx))
		args = append(args, req.RecurrenceRule)
		idx++
	}
	if req.Visibility != nil {
		setClauses = append(setClauses, fmt.Sprintf("visibility = $%d", idx))
		args = append(args, *req.Visibility)
		idx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", idx))
		args = append(args, *req.Status)
		idx++
	}
	if req.ImageURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("image_url = $%d", idx))
		args = append(args, req.ImageURL)
		idx++
	}

	// Add communityID as final argument
	args = append(args, communityID)
	query := fmt.Sprintf(`
		UPDATE communities
		SET %s
		WHERE id = $%d
		RETURNING
			id, owner_id, name, description, city, latitude, longitude,
			age_min, age_max, max_members, is_recurring, recurrence_rule,
			visibility, member_count, status, image_url, created_at, updated_at
	`, strings.Join(setClauses, ", "), idx)

	row := s.db.QueryRow(query, args...)
	updatedCommunity, err := scanCommunity(row)
	if err != nil {
		return nil, err
	}

	if s.search != nil && updatedCommunity != nil {
		go func(c *models.Community) {
			if err := s.search.SyncCommunity(c); err != nil {
				log.Printf("Error syncing community to Meilisearch: %v", err)
			}
		}(updatedCommunity)
	}

	return updatedCommunity, nil
}

func (s *CommunityService) Delete(communityID string, ownerClerkID string, userRole string) error {
	// 1. Get user UUID from Clerk ID
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, ownerClerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return err
	}

	// 2. Fetch the community owner to check ownership
	ownerID, err := s.GetCommunityOwner(communityID)
	if err != nil {
		return err
	}

	// 3. Ownership check: owner or admin
	if ownerID != userID && userRole != "admin" {
		return fmt.Errorf("forbidden")
	}

	// 4. Soft delete by updating status to 'deactivated'
	_, err = s.db.Exec(`UPDATE communities SET status = 'deactivated', updated_at = NOW() WHERE id = $1`, communityID)
	if err != nil {
		return err
	}

	// 5. Remove community from Meilisearch
	if s.search != nil {
		go func() {
			if err := s.search.DeleteCommunity(communityID); err != nil {
				log.Printf("Error deleting community from Meilisearch: %v", err)
			}
		}()
	}

	return nil
}

func (s *CommunityService) Join(communityID string, userClerkID string) error {
	// 1. Get user UUID from Clerk ID
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, userClerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 2. Fetch the community (and lock it) to check max_members and status
	var maxMembers *int
	var memberCount int
	var status string
	err = tx.QueryRow(`
		SELECT max_members, member_count, status
		FROM communities
		WHERE id = $1
		FOR UPDATE
	`, communityID).Scan(&maxMembers, &memberCount, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("community not found")
		}
		return err
	}

	if status == "deactivated" {
		return fmt.Errorf("cannot join a deactivated community")
	}

	if maxMembers != nil && memberCount >= *maxMembers {
		return fmt.Errorf("community is full")
	}

	// 3. Insert record into community_members
	_, err = tx.Exec(`
		INSERT INTO community_members (community_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, communityID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("already a member")
		}
		return err
	}

	// 4. Increment member_count
	_, err = tx.Exec(`
		UPDATE communities
		SET member_count = member_count + 1
		WHERE id = $1
	`, communityID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *CommunityService) Leave(communityID string, userClerkID string) error {
	// 1. Get user UUID from Clerk ID
	var userID string
	err := s.db.QueryRow(`SELECT id FROM users WHERE clerk_id = $1`, userClerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 2. Delete record from community_members and check if it deleted anything
	res, err := tx.Exec(`
		DELETE FROM community_members
		WHERE community_id = $1 AND user_id = $2
	`, communityID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("not a member of this community")
	}

	// 3. Decrement member_count
	_, err = tx.Exec(`
		UPDATE communities
		SET member_count = GREATEST(0, member_count - 1)
		WHERE id = $1
	`, communityID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *CommunityService) ListMembers(communityID string) ([]models.CommunityMember, error) {
	query := `
		SELECT
			u.id, u.clerk_id, u.email, u.full_name, u.username, u.avatar_url,
			cm.role, cm.joined_at
		FROM community_members cm
		JOIN users u ON cm.user_id = u.id
		WHERE cm.community_id = $1
		ORDER BY cm.joined_at ASC
	`
	rows, err := s.db.Query(query, communityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.CommunityMember
	for rows.Next() {
		var m models.CommunityMember
		err := rows.Scan(
			&m.UserID, &m.ClerkID, &m.Email, &m.FullName, &m.Username, &m.AvatarURL,
			&m.Role, &m.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	// return empty slice instead of nil
	if members == nil {
		members = []models.CommunityMember{}
	}

	return members, nil
}

func scanCommunity(row *sql.Row) (*models.Community, error) {
	var c models.Community
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.Name, &c.Description, &c.City, &c.Latitude, &c.Longitude,
		&c.AgeMin, &c.AgeMax, &c.MaxMembers, &c.IsRecurring, &c.RecurrenceRule,
		&c.Visibility, &c.MemberCount, &c.Status, &c.ImageURL, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
