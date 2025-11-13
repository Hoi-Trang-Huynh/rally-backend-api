package model

import "time"

type User struct {
	UserID      string    `json:"userId" db:"user_id"`
	FirebaseUID string    `json:"firebaseUid" db:"firebase_uid"`
	Email       string    `json:"email" db:"email"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}