package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlaceResult is a single place returned from the Google Places proxy.
type PlaceResult struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Lat           float64  `json:"lat"`
	Lng           float64  `json:"lng"`
	ImageUrl      string   `json:"imageUrl,omitempty"`
	Rating        *float64 `json:"rating,omitempty"`
	ReviewCount   *int     `json:"reviewCount,omitempty"`
	PriceLevel    string   `json:"priceLevel,omitempty"`
	Description   string   `json:"description,omitempty"`
	Address       string   `json:"address,omitempty"`
	Hours         string   `json:"hours,omitempty"`
	OpenNow       *bool    `json:"openNow,omitempty"`
	Type          string   `json:"type,omitempty"`
	Distance      string   `json:"distance,omitempty"`
	PricePerNight string   `json:"pricePerNight,omitempty"`
}

// NearbyPlacesResponse wraps a list of nearby places.
type NearbyPlacesResponse struct {
	Places []PlaceResult `json:"places"`
}

// SavedPlace is a MongoDB document storing a user's bookmarked place.
type SavedPlace struct {
	ID        primitive.ObjectID `json:"id,omitempty"   bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"-"              bson:"user_id"`
	PlaceData PlaceResult        `json:"placeData"      bson:"place_data"`
	SavedAt   time.Time          `json:"savedAt"        bson:"saved_at"`
}

// SavePlaceRequest is the request body for saving a place.
type SavePlaceRequest struct {
	PlaceID string `json:"placeId"`
}

// SavedPlacesResponse wraps a list of saved places.
type SavedPlacesResponse struct {
	Places []PlaceResult `json:"places"`
}
