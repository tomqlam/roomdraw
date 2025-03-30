package handlers

import (
	"log"
	"roomdraw/backend/pkg/models"
)

func isHigherPriority(user1 models.UserRaw, user2 models.UserRaw, dormId int) bool {
	if user1.Preplaced && !user2.Preplaced {
		return true
	} else if !user1.Preplaced && user2.Preplaced {
		return false
	}

	var user1YearNumber int

	switch user1.Year {
	case "sophomore":
		user1YearNumber = 2
	case "junior":
		user1YearNumber = 3
	case "senior":
		user1YearNumber = 4
		if dormId == user1.InDorm {
			user1YearNumber = 5
		}
	}

	var user2YearNumber int

	switch user2.Year {
	case "sophomore":
		user2YearNumber = 2
	case "junior":
		user2YearNumber = 3
	case "senior":
		user2YearNumber = 4
		if dormId == user2.InDorm {
			user2YearNumber = 5
		}
	}

	if user1YearNumber > user2YearNumber {
		return true
	} else if user1YearNumber < user2YearNumber {
		return false
	} else {
		return user1.DrawNumber < user2.DrawNumber
	}
}

// given a list of users, return the sorted list of users by priority highest to lowest
func sortUsersByPriority(users []models.UserRaw, dormId int) []models.UserRaw {
	if len(users) <= 1 {
		return users
	}

	pivot := users[0]
	var left []models.UserRaw
	var right []models.UserRaw

	for _, user := range users[1:] {
		if isHigherPriority(user, pivot, dormId) {
			left = append(left, user)
		} else {
			right = append(right, user)
		}
	}

	left = sortUsersByPriority(left, dormId)
	right = sortUsersByPriority(right, dormId)

	return append(append(left, pivot), right...)
}

func generateEmptyPriority() models.PullPriority {
	return models.PullPriority{
		Valid:       false,
		IsPreplaced: false,
		HasInDorm:   false,
		DrawNumber:  0,
		Year:        0,
		Inherited: models.InheritedPullPriority{
			HasInDorm:  false,
			DrawNumber: 0,
			Year:       0,
		},
	}
}

func generateUserPriority(user models.UserRaw, dormId int) models.PullPriority {
	if user.Preplaced {
		return models.PullPriority{
			IsPreplaced: true,
			HasInDorm:   false,
			DrawNumber:  0,
			Year:        0,
		}
	} else {
		var yearNumber int
		var hasInDorm bool
		switch user.Year {
		case "sophomore":
			yearNumber = 2
		case "junior":
			yearNumber = 3
		case "senior":
			yearNumber = 4
			if dormId == user.InDorm {
				hasInDorm = true
			}
		}

		return models.PullPriority{
			IsPreplaced: false,
			HasInDorm:   hasInDorm,
			DrawNumber:  user.DrawNumber,
			Year:        yearNumber,
		}
	}
}

// returns if the first priority is higher than the second
func comparePullPriority(priority1 models.PullPriority, priority2 models.PullPriority) bool {
	// log.Println("priority1", priority1)
	// log.Println("priority2", priority2)

	if !priority1.Valid {
		return false
	}

	if !priority2.Valid {
		return true
	}

	if priority1.IsPreplaced && !priority2.IsPreplaced {
		return true
	} else if !priority1.IsPreplaced && priority2.IsPreplaced {
		return false
	} else if priority1.IsPreplaced && priority2.IsPreplaced {
		return false
	}

	// check inherited to see if either has inherited priority
	p1EffectiveDrawNumber := priority1.DrawNumber
	p1EffectiveYear := priority1.Year

	p2EffectiveDrawNumber := priority2.DrawNumber
	p2EffectiveYear := priority2.Year

	if priority1.Inherited.Valid {
		p1EffectiveDrawNumber = priority1.Inherited.DrawNumber
		p1EffectiveYear = priority1.Inherited.Year
		if priority1.Inherited.HasInDorm && p1EffectiveYear == 4 {
			p1EffectiveYear = 5
		}
	} else {
		if priority1.HasInDorm && p1EffectiveYear == 4 {
			p1EffectiveYear = 5
		}
	}

	if priority2.Inherited.Valid {
		p2EffectiveDrawNumber = priority2.Inherited.DrawNumber
		p2EffectiveYear = priority2.Inherited.Year
		if priority2.Inherited.HasInDorm && p2EffectiveYear == 4 {
			p2EffectiveYear = 5
		}
	} else {
		if priority2.HasInDorm && p2EffectiveYear == 4 {
			p2EffectiveYear = 5
		}
	}

	if priority1.PullType == 3 {
		p1EffectiveYear = 6
	}

	if priority2.PullType == 3 {
		p2EffectiveYear = 6
	}

	log.Println("p1EffectiveDrawNumber", p1EffectiveDrawNumber)
	log.Println("p1EffectiveYear", p1EffectiveYear)
	log.Println("p2EffectiveDrawNumber", p2EffectiveDrawNumber)
	log.Println("p2EffectiveYear", p2EffectiveYear)

	if p1EffectiveYear > p2EffectiveYear {
		return true
	} else if p1EffectiveYear < p2EffectiveYear {
		return false
	} else {
		return p1EffectiveDrawNumber < p2EffectiveDrawNumber
	}
}

// hasIntersection checks if two string arrays have any elements in common
func hasIntersection(arr1, arr2 []string) bool {
	for _, item1 := range arr1 {
		for _, item2 := range arr2 {
			if item1 == item2 {
				return true
			}
		}
	}
	return false
}

// findIntersection returns the intersection of two string arrays
func findIntersection(arr1, arr2 []string) []string {
	result := []string{}
	for _, item1 := range arr1 {
		for _, item2 := range arr2 {
			if item1 == item2 {
				result = append(result, item1)
				break
			}
		}
	}
	return result
}

// findIntersectionOfPreferences returns the intersection of multiple string arrays
func findIntersectionOfPreferences(preferencesArray [][]string) []string {
	if len(preferencesArray) == 0 {
		return []string{}
	}

	result := preferencesArray[0]
	for i := 1; i < len(preferencesArray); i++ {
		result = findIntersection(result, preferencesArray[i])
	}
	return result
}

// GetSuiteGenderPreference determines the gender preference for a suite based on the priority of its occupants
// Returns the gender preferences array and a boolean indicating if a preference was found
func GetSuiteGenderPreference(users []models.UserRaw, dormId int) ([]string, bool) {
	if len(users) == 0 {
		return []string{}, false
	}

	var preplacedUsersPreferences [][]string
	anyPreplacedExist := false // Flag to track if Rule 3 applies at all

	for _, user := range users {
		if user.Preplaced {
			anyPreplacedExist = true // Found at least one preplaced user
			if len(user.GenderPreferences) > 0 {
				preplacedUsersPreferences = append(preplacedUsersPreferences, user.GenderPreferences)
			}
		}
	}

	// --- Rule 3 Logic ---
	if anyPreplacedExist {
		// Subcase 3a: At least one preplaced user HAS preferences
		if len(preplacedUsersPreferences) > 0 {
			intersection := findIntersectionOfPreferences(preplacedUsersPreferences)
			if len(intersection) > 0 {
				return intersection, true // Found valid intersection
			} else {
				// Conflict among preplaced users with preferences
				log.Printf("Warning: No intersection found between preplaced users' gender preferences in suite (conflict)")
				return []string{}, false // Error state
			}
		} else {
			// Subcase 3b: Preplaced users exist, but NONE have preferences
			log.Printf("Info: Preplaced users exist, but none have specified preferences. Suite preference is empty.")
			return []string{}, false // Suite preference is empty, indicate no specific pref found
			// Note: Returning 'false' here signifies "no specific preference determined by rule",
			// which aligns with returning empty prefs due to conflict or lack of input.
			// The CALLER needs to know that {} + false means "set suite to {}"
		}
	}

	// --- Rule 2 Logic (Only reached if !anyPreplacedExist) ---
	log.Printf("Info: No preplaced users found. Determining preference by highest priority.")
	sortedUsers := sortUsersByPriority(users, dormId) // Sort only non-preplaced users

	for _, user := range sortedUsers {
		// Since we already know no preplaced users exist, we just check preferences
		if len(user.GenderPreferences) > 0 {
			return user.GenderPreferences, true // Found highest priority user with preferences
		}
	}

	// No user (in the non-preplaced group) had preferences, or list was empty after filtering
	log.Printf("Info: No preplaced users, and no non-preplaced users had preferences. Suite preference is empty.")
	return []string{}, false // No preference found based on Rule 2
}
