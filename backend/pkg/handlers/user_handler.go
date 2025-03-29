package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetUsers retrieves all users
func GetUsers(c *gin.Context) {
	// Example SQL query
	rows, err := database.DB.Query("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time, room_uuid FROM users")
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var users []models.UserRaw
	for rows.Next() {
		var user models.UserRaw
		if err := rows.Scan(&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber, &user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated, &user.PartitipationTime, &user.RoomUUID); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}

		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// GetUsersPagedAndSorted retrieves users with pagination and sorting
func GetUsersPagedAndSorted(c *gin.Context) {
	// Get pagination and sorting parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sortBy := c.DefaultQuery("sort_by", "id")
	sortOrder := c.DefaultQuery("sort_order", "asc")
	yearValues := c.QueryArray("year")
	minDrawNumber := c.Query("min_draw_number")
	maxDrawNumber := c.Query("max_draw_number")
	inDormValues := c.QueryArray("in_dorm")
	genderPreferenceValues := c.QueryArray("gender_preference")
	preplacedQuery := c.Query("preplaced")
	hasGenderPrefQuery := c.Query("has_gender_preference")

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Build base query
	baseQuery := "SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time, room_uuid, reslife_role, email, gender_preferences FROM users"

	// Build WHERE clause
	whereClause := ""
	args := []interface{}{}
	argCount := 1

	if len(yearValues) > 0 {
		if whereClause == "" {
			whereClause += " WHERE"
		} else {
			whereClause += " AND"
		}

		whereClause += " ("
		yearPlaceholders := make([]string, len(yearValues))
		for i, year := range yearValues {
			yearPlaceholders[i] = fmt.Sprintf("year = $%d", argCount)
			args = append(args, year)
			argCount++
		}
		whereClause += strings.Join(yearPlaceholders, " OR ")
		whereClause += ")"
	}

	if minDrawNumber != "" {
		minDraw, err := strconv.ParseFloat(minDrawNumber, 64)
		if err == nil {
			if whereClause == "" {
				whereClause += " WHERE"
			} else {
				whereClause += " AND"
			}
			whereClause += " draw_number >= $" + strconv.Itoa(argCount)
			args = append(args, minDraw)
			argCount++
		}
	}

	if maxDrawNumber != "" {
		maxDraw, err := strconv.ParseFloat(maxDrawNumber, 64)
		if err == nil {
			if whereClause == "" {
				whereClause += " WHERE"
			} else {
				whereClause += " AND"
			}
			whereClause += " draw_number <= $" + strconv.Itoa(argCount)
			args = append(args, maxDraw)
			argCount++
		}
	}

	if len(inDormValues) > 0 {
		if whereClause == "" {
			whereClause += " WHERE"
		} else {
			whereClause += " AND"
		}

		whereClause += " ("
		dormPlaceholders := make([]string, len(inDormValues))
		for i, dormID := range inDormValues {
			dormInt, err := strconv.Atoi(dormID)
			if err == nil {
				dormPlaceholders[i] = fmt.Sprintf("in_dorm = $%d", argCount)
				args = append(args, dormInt)
				argCount++
			} else {
				// If conversion fails, make sure we have a valid placeholder
				dormPlaceholders[i] = "FALSE"
			}
		}
		cleanedPlaceholders := []string{}
		for _, p := range dormPlaceholders {
			if p != "FALSE" {
				cleanedPlaceholders = append(cleanedPlaceholders, p)
			}
		}

		if len(cleanedPlaceholders) > 0 {
			whereClause += strings.Join(cleanedPlaceholders, " OR ")
			whereClause += ")"
		} else {
			// If there are no valid dorm IDs, add a condition that's always false
			whereClause += " FALSE)"
		}
	}

	if len(genderPreferenceValues) > 0 {
		if whereClause == "" {
			whereClause += " WHERE"
		} else {
			whereClause += " AND"
		}

		whereClause += " ("
		genderPrefPlaceholders := make([]string, len(genderPreferenceValues))
		for i, genderPref := range genderPreferenceValues {
			genderPrefPlaceholders[i] = fmt.Sprintf("$%d = ANY(gender_preferences)", argCount)
			args = append(args, genderPref)
			argCount++
		}
		whereClause += strings.Join(genderPrefPlaceholders, " OR ")
		whereClause += ")"
	}

	if hasGenderPrefQuery != "" {
		hasGenderPref, err := strconv.ParseBool(hasGenderPrefQuery)
		if err == nil {
			if whereClause == "" {
				whereClause += " WHERE"
			} else {
				whereClause += " AND"
			}

			if hasGenderPref {
				whereClause += " array_length(gender_preferences, 1) > 0"
			} else {
				whereClause += " (gender_preferences IS NULL OR array_length(gender_preferences, 1) IS NULL OR array_length(gender_preferences, 1) = 0)"
			}
		}
	}

	if preplacedQuery != "" {
		preplaced, err := strconv.ParseBool(preplacedQuery)
		if err == nil {
			if whereClause == "" {
				whereClause += " WHERE"
			} else {
				whereClause += " AND"
			}
			whereClause += " preplaced = $" + strconv.Itoa(argCount)
			args = append(args, preplaced)
			argCount++
		}
	}

	// Validate sort column to prevent SQL injection
	allowedSortColumns := map[string]string{
		"id":          "id",
		"year":        "year",
		"first_name":  "first_name",
		"last_name":   "last_name",
		"draw_number": "draw_number",
	}

	validSortColumn, exists := allowedSortColumns[sortBy]
	if !exists {
		validSortColumn = "id" // Default to id if invalid
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc" // Default to ascending if invalid
	}

	// Log the final query for debugging
	log.Printf("User search query: %s%s with args: %v", baseQuery, whereClause, args)

	// Count total records query
	countQuery := "SELECT COUNT(*) FROM users" + whereClause
	var totalRecords int
	err := database.DB.QueryRow(countQuery, args...).Scan(&totalRecords)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database count query failed"})
		return
	}

	// Calculate total pages
	totalPages := (totalRecords + limit - 1) / limit

	// Build final query with sorting and pagination
	query := baseQuery + whereClause + " ORDER BY " + validSortColumn + " " + sortOrder + " LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	// Execute the query
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var users []models.UserRaw
	for rows.Next() {
		var user models.UserRaw
		if err := rows.Scan(&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber,
			&user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated,
			&user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole, &user.Email, &user.GenderPreferences); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"page":        page,
		"limit":       limit,
		"total":       totalRecords,
		"total_pages": totalPages,
	})
}

func GetUsersIdMap(c *gin.Context) {
	// Example SQL query
	rows, err := database.DB.Query("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time, room_uuid, reslife_role, gender_preferences FROM users")
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	userMap := make(map[int]models.UserRaw)
	for rows.Next() {
		var user models.UserRaw
		if err := rows.Scan(&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber, &user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated, &user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole, &user.GenderPreferences); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}
		userMap[user.Id] = user
	}

	c.JSON(http.StatusOK, userMap)
}

func GetUser(c *gin.Context) {
	// Get the user id from the URL
	userid := c.Param("userid")

	// Query for a single user
	var user models.UserRaw
	err := database.DB.QueryRow("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time, room_uuid, reslife_role, email, gender_preferences FROM users WHERE id=$1", userid).Scan(
		&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber,
		&user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated,
		&user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole, &user.Email,
		&user.GenderPreferences,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetUserByEmail(c *gin.Context) {
	// Get the email from the query
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email query parameter is required"})
		return
	}

	// Query for a single user by email
	var user models.UserRaw
	err := database.DB.QueryRow("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time, room_uuid, reslife_role, email FROM users WHERE email=$1", email).Scan(
		&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber,
		&user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated,
		&user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole, &user.Email,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// User not found - return empty object with status 200 (not a 404, because this is expected for guests)
			c.JSON(http.StatusOK, gin.H{"found": false})
			return
		}
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	// User found
	c.JSON(http.StatusOK, gin.H{
		"found": true,
		"user":  user,
	})
}
