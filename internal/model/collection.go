package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Collection is a MongoDB document representing a user-curated place list.
type Collection struct {
	ID        primitive.ObjectID `json:"id,omitempty"  bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"-"             bson:"user_id"`
	Name      string             `json:"name"          bson:"name"`
	PlaceIDs  []string           `json:"placeIds"      bson:"place_ids"`
	CreatedAt time.Time          `json:"createdAt"     bson:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt"     bson:"updated_at"`
}

// CreateCollectionRequest is the body for POST /collections.
type CreateCollectionRequest struct {
	Name string `json:"name"`
}

// AddPlaceToCollectionRequest is the body for POST /collections/:id/places.
type AddPlaceToCollectionRequest struct {
	PlaceID string `json:"placeId"`
}

// CollectionResponse is the public representation of a Collection.
type CollectionResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	PlaceIDs      []string  `json:"placeIds"`
	CoverImageURL string    `json:"coverImageUrl,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// CollectionsResponse wraps a list of collections.
type CollectionsResponse struct {
	Collections []CollectionResponse `json:"collections"`
}
