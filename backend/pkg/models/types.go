package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type IntArray []int

func (a *IntArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}

	switch src := src.(type) {
	case []byte:
		// Convert byte array to string
		str := string(src)

		// Trim the curly braces and split the string
		trimmed := strings.Trim(str, "{}")
		if trimmed == "" {
			return nil // Empty array
		}
		elements := strings.Split(trimmed, ",")

		// Convert the split strings to integers
		result := make(IntArray, len(elements))
		for i, s := range elements {
			intVal, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			result[i] = intVal
		}

		*a = result
		return nil

	default:
		return fmt.Errorf("unsupported type for IntArray: %T", src)
	}
}

type UUIDArray []uuid.UUID

func (a *UUIDArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}

	switch src := src.(type) {
	case []byte:
		str := string(src)
		trimmed := strings.Trim(str, "{}")
		if trimmed == "" {
			return nil // Empty array
		}
		elements := strings.Split(trimmed, ",")

		result := make(UUIDArray, len(elements))
		for i, s := range elements {
			// Trim any whitespace around the UUID
			uuidStr := strings.TrimSpace(s)
			uuidVal, err := uuid.Parse(uuidStr)
			if err != nil {
				return fmt.Errorf("invalid UUID: %v", err)
			}
			result[i] = uuidVal
		}

		*a = result
		return nil

	default:
		return fmt.Errorf("unsupported type for UUIDArray: %T", src)
	}
}

type RoomRaw struct {
	RoomUUID         uuid.UUID    `db:"room_uuid"`
	Dorm             int          `db:"dorm"`
	DormName         string       `db:"dorm_name"`
	RoomID           string       `db:"room_id"`
	SuiteUUID        uuid.UUID    `db:"suite_uuid"` // Use sql.NullString for nullable fields
	MaxOccupancy     int          `db:"max_occupancy"`
	CurrentOccupancy int          `db:"current_occupancy"`
	Occupants        IntArray     `db:"occupants"`
	PullPriority     PullPriority `db:"pull_priority"`
	SGroupUUID       uuid.UUID    `db:"sgroup_uuid"`
	HasFrosh         bool         `db:"has_frosh"`
	FroshRoomType    int          `db:"frosh_room_type"`
}

type SuiteRaw struct {
	SuiteUUID              uuid.UUID      `db:"suite_uuid"`
	Dorm                   int            `db:"dorm"`
	DormName               string         `db:"dorm_name"`
	Floor                  int            `db:"floor"`
	RoomCount              int            `db:"room_count"`
	Rooms                  UUIDArray      `db:"rooms"`
	AlternativePull        bool           `db:"alternative_pull"`
	SuiteDesign            string         `db:"suite_design"`
	CanLockPull            bool           `db:"can_lock_pull"`
	LockPulledRoom         uuid.UUID      `db:"lock_pulled_room"`
	ReslifeRoom            uuid.UUID      `db:"reslife_room"`
	GenderPreferences      pq.StringArray `db:"gender_preferences"`
	CanBeGenderPreferenced bool           `db:"can_be_gender_preferenced"`
}

type DormSimple struct {
	DormName    string        `json:"dormName"`
	Description string        `json:"description"`
	Floors      []FloorSimple `json:"floors"`
}

type DormSimpler struct {
	Floors []FloorSimpler `json:"floors"`
}

type FloorSimple struct {
	FloorNumber int           `json:"floorNumber"`
	Suites      []SuiteSimple `json:"suites"`
}

type FloorSimpler struct {
	Suites []SuiteSimpler `json:"suites"`
}

type SuiteSimple struct {
	Rooms             []RoomSimple   `json:"rooms"`
	SuiteDesign       string         `json:"suiteDesign"`
	SuiteUUID         uuid.UUID      `json:"suiteUUID"`
	GenderPreferences pq.StringArray `json:"genderPreferences"`
	AlternativePull   bool           `json:"alternative_pull"`
	CanLockPull       bool           `json:"can_lock_pull"`
}

type SuiteSimpler struct {
	Rooms           []RoomSimpler `json:"rooms"`
	AlternativePull bool          `json:"alternative_pull"`
}

type RoomSimple struct {
	RoomNumber    string       `json:"roomNumber"`
	MaxOccupancy  int          `json:"maxOccupancy"`
	PullPriority  PullPriority `json:"pullPriority"`
	RoomUUID      uuid.UUID    `json:"roomUUID"`
	Occupant1     int          `json:"occupant1"`
	Occupant2     int          `json:"occupant2"`
	Occupant3     int          `json:"occupant3"`
	Occupant4     int          `json:"occupant4"`
	HasFrosh      bool         `json:"hasFrosh"`
	FroshRoomType int          `json:"froshRoomType"`
	Dorm          int          `json:"dorm"`
}

type RoomSimpler struct {
	RoomNumber    string `json:"room_number"`
	MaxOccupancy  int    `json:"capacity"`
	FroshRoomType int    `json:"frosh_room_type"`
}

type UserRaw struct {
	Id                    int            `db:"id"`
	Year                  string         `db:"year"`
	FirstName             string         `db:"first_name"`
	LastName              string         `db:"last_name"`
	Email                 string         `db:"email"`
	DrawNumber            float64        `db:"draw_number"`
	Preplaced             bool           `db:"preplaced"`
	InDorm                int            `db:"in_dorm"`
	SGroupUUID            uuid.UUID      `db:"sgroup_uuid"`
	Participated          bool           `db:"participated"`
	PartitipationTime     pq.NullTime    `db:"participation_time"`
	RoomUUID              uuid.UUID      `db:"room_uuid"`
	ReslifeRole           string         `db:"reslife_role"`
	NotificationsEnabled  bool           `db:"notifications_enabled"`
	NotificationCreatedAt time.Time      `db:"notification_created_at"`
	NotificationUpdatedAt time.Time      `db:"notification_updated_at"`
	GenderPreferences     pq.StringArray `db:"gender_preferences"`
}

type SuiteGroupRaw struct {
	SGroupUUID     uuid.UUID `db:"sgroup_uuid"`
	SGroupSize     int       `db:"sgroup_size"`
	SGroupName     string    `db:"sgroup_name"`
	SGroupSuite    uuid.UUID `db:"sgroup_suite"`
	SGroupPriority string    `db:"sgroup_priority"`
	Rooms          UUIDArray `db:"rooms"`
	Disbanded      bool      `db:"disbanded"`
	LockPulledRoom uuid.UUID `db:"contains_lock_pull"`
	ReslifeRoom    uuid.UUID `db:"reslife_room"`
}

func (pp *PullPriority) Scan(src interface{}) error {
	// src is a JSON/JSONB value from PostgreSQL, so it should be a byte slice or string.
	var source []byte
	switch src := src.(type) {
	case []byte:
		source = src
	case string:
		source = []byte(src)
	default:
		return errors.New("incompatible type for PullPriority")
	}

	// Unmarshal JSON to the PullPriority struct
	err := json.Unmarshal(source, pp)
	if err != nil {
		return err
	}

	return nil
}

type PullPriority struct {
	Valid       bool                  `json:"valid"`
	IsPreplaced bool                  `json:"isPreplaced"`
	HasInDorm   bool                  `json:"hasInDorm"`
	DrawNumber  float64               `json:"drawNumber"`
	Year        int                   `json:"year"`     // 0 = undefined, 1 = freshman, 2 = sophomore, 3 = junior, 4 = senior
	PullType    int                   `json:"pullType"` // 0 = undefined, 1 = self, 2 = normal pull, 3 = lock pull, 4 = alternative pull
	Inherited   InheritedPullPriority `json:"inherited"`
}

type InheritedPullPriority struct {
	Valid      bool    `json:"valid"`
	HasInDorm  bool    `json:"hasInDorm"`
	DrawNumber float64 `json:"drawNumber"`
	Year       int     `json:"year"` // 1 = freshman, 2 = sophomore, 3 = junior, 4 = senior
}

type OccupantUpdateRequest struct {
	ProposedOccupants IntArray  `json:"proposedOccupants"`
	PullType          int       `json:"pullType"` // 0 = undefined, 1 = self, 2 = normal pull, 3 = lock pull, 4 = alternative pull
	PullLeaderRoom    uuid.UUID `json:"pullLeaderRoom"`
	UserJWT           string    `json:"userJWT"`
}

type SuiteDesignUpdateRequest struct {
	SuiteDesign string `json:"suiteDesign"`
	UserJWT     string `json:"userJWT"`
}

type BumpFroshRequest struct {
	TargetRoomUUID uuid.UUID `json:"targetRoomUUID"`
	UserJWT        string    `json:"userJWT"`
}

type GeneralRequest struct {
	UserJWT string `json:"userJWT"`
}

type PreplacedRequest struct {
	ProposedOccupants IntArray `json:"proposedOccupants"`
	UserJWT           string   `json:"userJWT"`
}

// UserRateLimit represents an entry in the user_rate_limits table
type UserRateLimit struct {
	Email             string         `db:"email"`
	ClearRoomCount    int            `db:"clear_room_count"`
	ClearRoomDate     sql.NullTime   `db:"clear_room_date"`
	IsBlacklisted     bool           `db:"is_blacklisted"`
	BlacklistedAt     sql.NullTime   `db:"blacklisted_at"`
	BlacklistedReason sql.NullString `db:"blacklisted_reason"`
}

const (
	EntityTypeRoom  = "ROOM"
	EntityTypeUser  = "USER"
	EntityTypeSuite = "SUITE"
)