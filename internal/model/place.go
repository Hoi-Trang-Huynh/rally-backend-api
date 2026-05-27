package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlaceReview is a single Google Places review.
type PlaceReview struct {
	AuthorName  string `json:"authorName"            bson:"author_name"`
	AuthorPhoto string `json:"authorPhoto,omitempty" bson:"author_photo,omitempty"`
	Rating      int    `json:"rating"`
	TimeAgo     string `json:"timeAgo"               bson:"time_ago"`
	Text        string `json:"text"                  bson:"text"`
}

// PlaceResult is a single place returned from the Google Places proxy.
type PlaceResult struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Lat           float64       `json:"lat"`
	Lng           float64       `json:"lng"`
	ImageUrl      string        `json:"imageUrl,omitempty"`
	Photos        []string      `json:"photos,omitempty"        bson:"photos,omitempty"`
	Rating        *float64      `json:"rating,omitempty"`
	ReviewCount   *int          `json:"reviewCount,omitempty"`
	PriceLevel    string        `json:"priceLevel,omitempty"`
	Description   string        `json:"description,omitempty"`
	Address       string        `json:"address,omitempty"`
	Hours         string        `json:"hours,omitempty"`
	OpenNow       *bool         `json:"openNow,omitempty"`
	WeekdayHours  []string      `json:"weekdayHours,omitempty"  bson:"weekday_hours,omitempty"`
	Phone         string        `json:"phone,omitempty"         bson:"phone,omitempty"`
	Website       string        `json:"website,omitempty"       bson:"website,omitempty"`
	Type          string        `json:"type,omitempty"`
	Distance      string        `json:"distance,omitempty"`
	PricePerNight string        `json:"pricePerNight,omitempty"`
	Reviews       []PlaceReview `json:"reviews,omitempty"       bson:"reviews,omitempty"`
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
