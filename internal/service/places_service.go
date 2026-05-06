package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
)

const (
	placesBaseURL = "https://maps.googleapis.com/maps/api/place"
	photoBaseURL  = "https://maps.googleapis.com/maps/api/place/photo"
)

type PlacesService struct {
	apiKey     string
	httpClient *http.Client
}

func NewPlacesService(apiKey string) *PlacesService {
	return &PlacesService{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NearbySearch proxies Google Places Nearby Search and returns up to maxCount results.
func (s *PlacesService) NearbySearch(ctx context.Context, lat, lng float64, placeType string, maxCount int) ([]model.PlaceResult, error) {
	params := url.Values{}
	params.Set("location", fmt.Sprintf("%f,%f", lat, lng))
	params.Set("radius", "5000")
	params.Set("type", placeType)
	params.Set("key", s.apiKey)

	reqURL := fmt.Sprintf("%s/nearbysearch/json?%s", placesBaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		Status  string              `json:"status"`
		Results []nearbyPlaceResult `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if apiResp.Status == "REQUEST_DENIED" || apiResp.Status == "OVER_QUERY_LIMIT" {
		return []model.PlaceResult{}, nil
	}
	if apiResp.Status != "OK" && apiResp.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("places API error: %s", apiResp.Status)
	}

	results := make([]model.PlaceResult, 0, min(len(apiResp.Results), maxCount))
	for i, r := range apiResp.Results {
		if i >= maxCount {
			break
		}
		place := model.PlaceResult{
			ID:      r.PlaceID,
			Name:    r.Name,
			Lat:     r.Geometry.Location.Lat,
			Lng:     r.Geometry.Location.Lng,
			Address: r.Vicinity,
			Type:    placeType,
		}
		if r.Rating > 0 {
			v := r.Rating
			place.Rating = &v
		}
		if r.UserRatingsTotal > 0 {
			v := r.UserRatingsTotal
			place.ReviewCount = &v
		}
		if r.PriceLevel > 0 {
			place.PriceLevel = priceLevelStr(r.PriceLevel)
		}
		if len(r.Photos) > 0 {
			place.ImageUrl = buildPhotoURL(r.Photos[0].PhotoReference, s.apiKey)
		}
		if r.OpeningHours != nil {
			v := r.OpeningHours.OpenNow
			place.OpenNow = &v
			if v {
				place.Hours = "Open now"
			} else {
				place.Hours = "Closed"
			}
		}
		place.Distance = formatDistance(haversineKm(lat, lng, r.Geometry.Location.Lat, r.Geometry.Location.Lng))
		results = append(results, place)
	}

	return results, nil
}

// GetPlaceDetails proxies Google Places Details and returns a single PlaceResult.
func (s *PlacesService) GetPlaceDetails(ctx context.Context, placeID string) (*model.PlaceResult, error) {
	params := url.Values{}
	params.Set("place_id", placeID)
	params.Set("fields", "place_id,name,photos,rating,user_ratings_total,price_level,formatted_address,opening_hours,geometry,editorial_summary,types")
	params.Set("key", s.apiKey)

	reqURL := fmt.Sprintf("%s/details/json?%s", placesBaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		Status string            `json:"status"`
		Result placeDetailResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if apiResp.Status == "NOT_FOUND" || apiResp.Status == "INVALID_REQUEST" {
		return nil, errors.New("place not found")
	}
	if apiResp.Status != "OK" {
		return nil, fmt.Errorf("places API error: %s", apiResp.Status)
	}

	r := apiResp.Result
	place := &model.PlaceResult{
		ID:      r.PlaceID,
		Name:    r.Name,
		Lat:     r.Geometry.Location.Lat,
		Lng:     r.Geometry.Location.Lng,
		Address: r.FormattedAddress,
	}
	if r.Rating > 0 {
		v := r.Rating
		place.Rating = &v
	}
	if r.UserRatingsTotal > 0 {
		v := r.UserRatingsTotal
		place.ReviewCount = &v
	}
	if r.PriceLevel > 0 {
		place.PriceLevel = priceLevelStr(r.PriceLevel)
	}
	if len(r.Photos) > 0 {
		place.ImageUrl = buildPhotoURL(r.Photos[0].PhotoReference, s.apiKey)
	}
	if r.EditorialSummary != nil {
		place.Description = r.EditorialSummary.Overview
	}
	if r.OpeningHours != nil {
		v := r.OpeningHours.OpenNow
		place.OpenNow = &v
		if len(r.OpeningHours.WeekdayText) > 0 {
			place.Hours = todayHours(r.OpeningHours.WeekdayText)
		}
	}
	if len(r.Types) > 0 {
		place.Type = r.Types[0]
	}

	return place, nil
}

// ── Internal Google API structs ───────────────────────────────────────────────

type nearbyPlaceResult struct {
	PlaceID  string `json:"place_id"`
	Name     string `json:"name"`
	Geometry struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
	Photos []struct {
		PhotoReference string `json:"photo_reference"`
	} `json:"photos"`
	Rating           float64 `json:"rating"`
	UserRatingsTotal int     `json:"user_ratings_total"`
	PriceLevel       int     `json:"price_level"`
	Vicinity         string  `json:"vicinity"`
	OpeningHours     *struct {
		OpenNow bool `json:"open_now"`
	} `json:"opening_hours"`
}

type placeDetailResult struct {
	PlaceID  string `json:"place_id"`
	Name     string `json:"name"`
	Geometry struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
	Photos []struct {
		PhotoReference string `json:"photo_reference"`
	} `json:"photos"`
	Rating           float64  `json:"rating"`
	UserRatingsTotal int      `json:"user_ratings_total"`
	PriceLevel       int      `json:"price_level"`
	FormattedAddress string   `json:"formatted_address"`
	OpeningHours     *struct {
		OpenNow     bool     `json:"open_now"`
		WeekdayText []string `json:"weekday_text"`
	} `json:"opening_hours"`
	EditorialSummary *struct {
		Overview string `json:"overview"`
	} `json:"editorial_summary"`
	Types []string `json:"types"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildPhotoURL(ref, apiKey string) string {
	return fmt.Sprintf("%s?maxwidth=800&photo_reference=%s&key=%s", photoBaseURL, url.QueryEscape(ref), apiKey)
}

func priceLevelStr(level int) string {
	switch level {
	case 1:
		return "$"
	case 2:
		return "$$"
	case 3:
		return "$$$"
	case 4:
		return "$$$$"
	default:
		return ""
	}
}

// todayHours extracts the current weekday's hours string from weekday_text.
// weekday_text is Monday-indexed: [Mon, Tue, Wed, Thu, Fri, Sat, Sun].
func todayHours(weekdayText []string) string {
	// time.Weekday: 0=Sunday … 6=Saturday; convert to Monday-indexed (0=Mon … 6=Sun)
	idx := (int(time.Now().Weekday()) + 6) % 7
	if idx >= len(weekdayText) {
		return ""
	}
	text := weekdayText[idx]
	// strip "Monday: " prefix → "9:00 AM – 5:00 PM"
	if i := strings.Index(text, ": "); i != -1 {
		return text[i+2:]
	}
	return text
}

// haversineKm returns the great-circle distance in kilometres between two coordinates.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return r * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// formatDistance converts kilometres to a human-readable string.
func formatDistance(km float64) string {
	if km < 1 {
		return fmt.Sprintf("%.0f m", km*1000)
	}
	return fmt.Sprintf("%.1f km", km)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
