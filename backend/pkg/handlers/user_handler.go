package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"
	"strconv"

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
	year := c.Query("year")
	minDrawNumber := c.Query("min_draw_number")
	maxDrawNumber := c.Query("max_draw_number")
	inDorm := c.Query("in_dorm")

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
	baseQuery := "SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time, room_uuid, reslife_role, email FROM users"

	// Build WHERE clause
	whereClause := ""
	args := []interface{}{}
	argCount := 1

	if year != "" {
		whereClause += " WHERE year = $" + strconv.Itoa(argCount)
		args = append(args, year)
		argCount++
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

	if inDorm != "" {
		dormID, err := strconv.Atoi(inDorm)
		if err == nil {
			if whereClause == "" {
				whereClause += " WHERE"
			} else {
				whereClause += " AND"
			}
			whereClause += " in_dorm = $" + strconv.Itoa(argCount)
			args = append(args, dormID)
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
			&user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole, &user.Email); err != nil {
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
	rows, err := database.DB.Query("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time,room_uuid, reslife_role FROM users")
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	userMap := make(map[int]models.UserRaw)
	for rows.Next() {
		var user models.UserRaw
		if err := rows.Scan(&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber, &user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated, &user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole); err != nil {
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
	err := database.DB.QueryRow("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time, room_uuid, reslife_role, email FROM users WHERE id=$1", userid).Scan(
		&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber,
		&user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated,
		&user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole, &user.Email,
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
