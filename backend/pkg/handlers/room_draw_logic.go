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

// returns if the first priority is higher than the second
func comparePullPriority(priority1 models.PullPriority, priority2 models.PullPriority) bool {
	log.Println("priority1", priority1)
	log.Println("priority2", priority2)

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
	} else if priority2.Inherited.Valid {
		p2EffectiveDrawNumber = priority2.Inherited.DrawNumber
		p2EffectiveYear = priority2.Inherited.Year
	}

	log.Println("p1EffectiveDrawNumber", p1EffectiveDrawNumber)
	log.Println("p1EffectiveYear", p1EffectiveYear)
	log.Println("p2EffectiveDrawNumber", p2EffectiveDrawNumber)
	log.Println("p2EffectiveYear", p2EffectiveYear)

	if priority1.Inherited.HasInDorm && p1EffectiveYear == 4 {
		p1EffectiveYear = 5
	}

	if priority2.Inherited.HasInDorm && p2EffectiveYear == 4 {
		p2EffectiveYear = 5
	}

	if priority1.PullType == 3 {
		p1EffectiveYear = 6
	}

	if priority2.PullType == 3 {
		p2EffectiveYear = 6
	}

	if p1EffectiveYear > p2EffectiveYear {
		return true
	} else if p1EffectiveYear < p2EffectiveYear {
		return false
	} else {
		return p1EffectiveDrawNumber < p2EffectiveDrawNumber
	}
}
