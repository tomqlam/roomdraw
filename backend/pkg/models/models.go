package models

type BumpNotification struct {
	UserID   int
	RoomID   string
	DormName string
}

type BumpNotificationQueue struct {
	Notifications []BumpNotification
}

func NewBumpNotificationQueue() *BumpNotificationQueue {
	return &BumpNotificationQueue{
		Notifications: make([]BumpNotification, 0),
	}
}

func (q *BumpNotificationQueue) Add(userID int, roomID string, dormName string) {
	q.Notifications = append(q.Notifications, BumpNotification{
		UserID:   userID,
		RoomID:   roomID,
		DormName: dormName,
	})
}
