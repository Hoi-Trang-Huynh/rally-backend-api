package model

import "time"

type User struct {
	UserID      string    `json:"userId" bson:"user_id"`
	FirebaseUID string    `json:"firebaseUid" bson:"firebase_uid"`
	Email       string    `json:"email" bson:"email"`
	CreatedAt   time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updated_at"`
}