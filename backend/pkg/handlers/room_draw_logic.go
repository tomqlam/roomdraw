package handlers

import (
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
// use quicksort
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

		switch user.Year {
		case "sophomore":
			yearNumber = 2
		case "junior":
			yearNumber = 3
		case "senior":
			yearNumber = 4
			if dormId == user.InDorm {
				yearNumber = 5
			}
		}

		return models.PullPriority{
			IsPreplaced: false,
			HasInDorm:   dormId == user.InDorm,
			DrawNumber:  user.DrawNumber,
			Year:        yearNumber,
		}
	}
}
