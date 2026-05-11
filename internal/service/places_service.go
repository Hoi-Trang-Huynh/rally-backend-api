package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
)

const (
	placesBaseURL = "https://places.googleapis.com/v1"

	// usePlaceholder toggles static placeholder data instead of calling the Google Places API.
	usePlaceholder = false

	nearbyFieldMask  = "places.id,places.displayName,places.location,places.rating,places.userRatingCount,places.priceLevel,places.shortFormattedAddress,places.photos,places.currentOpeningHours,places.types"
	detailsFieldMask = "id,displayName,location,rating,userRatingCount,priceLevel,formattedAddress,photos,currentOpeningHours,types,editorialSummary"
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

// NearbySearch proxies Google Places Nearby Search (New) and returns up to maxCount results.
func (s *PlacesService) NearbySearch(ctx context.Context, lat, lng float64, placeType string, maxCount int) ([]model.PlaceResult, error) {
	if usePlaceholder {
		var results []model.PlaceResult
		for _, p := range placeholderPlaces {
			if p.Type == placeType {
				results = append(results, p)
				if len(results) >= maxCount {
					break
				}
			}
		}
		return results, nil
	}

	body, _ := json.Marshal(map[string]any{
		"includedTypes":   []string{placeType},
		"maxResultCount":  maxCount,
		"locationRestriction": map[string]any{
			"circle": map[string]any{
				"center": map[string]any{
					"latitude":  lat,
					"longitude": lng,
				},
				"radius": 5000.0,
			},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, placesBaseURL+"/places:searchNearby", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", s.apiKey)
	req.Header.Set("X-Goog-FieldMask", nearbyFieldMask)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("[places] NearbySearch: HTTP %d — %s", resp.StatusCode, errResp.Error.Message)
		return []model.PlaceResult{}, nil
	}

	var apiResp struct {
		Places []newPlaceResult `json:"places"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	results := make([]model.PlaceResult, 0, len(apiResp.Places))
	for _, r := range apiResp.Places {
		results = append(results, s.convertPlace(r, placeType, lat, lng))
	}
	return results, nil
}

// GetPlaceDetails proxies Google Places Details (New) and returns a single PlaceResult.
func (s *PlacesService) GetPlaceDetails(ctx context.Context, placeID string) (*model.PlaceResult, error) {
	if usePlaceholder {
		for _, p := range placeholderPlaces {
			if p.ID == placeID {
				result := p
				return &result, nil
			}
		}
		return nil, errors.New("place not found")
	}

	reqURL := fmt.Sprintf("%s/places/%s", placesBaseURL, url.PathEscape(placeID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Goog-Api-Key", s.apiKey)
	req.Header.Set("X-Goog-FieldMask", detailsFieldMask)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("place not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("places API error: HTTP %d", resp.StatusCode)
	}

	var r newPlaceResult
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	place := s.convertPlace(r, "", 0, 0)
	if r.EditorialSummary != nil {
		place.Description = r.EditorialSummary.Text
	}
	if r.FormattedAddress != "" {
		place.Address = r.FormattedAddress
	}
	if len(r.CurrentOpeningHours.WeekdayDescriptions) > 0 {
		place.Hours = todayHours(r.CurrentOpeningHours.WeekdayDescriptions)
	}
	return &place, nil
}

// TextSearch proxies Google Places Text Search (New) and returns up to maxCount results.
func (s *PlacesService) TextSearch(ctx context.Context, query string, lat, lng float64, maxCount int) ([]model.PlaceResult, error) {
	if usePlaceholder {
		q := strings.ToLower(query)
		results := make([]model.PlaceResult, 0)
		for _, p := range placeholderPlaces {
			if strings.Contains(strings.ToLower(p.Name), q) ||
				strings.Contains(strings.ToLower(p.Description), q) ||
				strings.Contains(strings.ToLower(p.Address), q) {
				results = append(results, p)
				if len(results) >= maxCount {
					break
				}
			}
		}
		return results, nil
	}

	body, _ := json.Marshal(map[string]any{
		"textQuery":      query,
		"maxResultCount": maxCount,
		"locationBias": map[string]any{
			"circle": map[string]any{
				"center": map[string]any{
					"latitude":  lat,
					"longitude": lng,
				},
				"radius": 5000.0,
			},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, placesBaseURL+"/places:searchText", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", s.apiKey)
	req.Header.Set("X-Goog-FieldMask", nearbyFieldMask)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		log.Printf("[places] TextSearch: HTTP %d — %s", resp.StatusCode, errResp.Error.Message)
		return []model.PlaceResult{}, nil
	}

	var apiResp struct {
		Places []newPlaceResult `json:"places"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	results := make([]model.PlaceResult, 0, len(apiResp.Places))
	for _, r := range apiResp.Places {
		results = append(results, s.convertPlace(r, "", lat, lng))
	}
	return results, nil
}

// ── Internal Google Places API (New) structs ──────────────────────────────────

type newPlaceResult struct {
	ID          string `json:"id"`
	DisplayName struct {
		Text string `json:"text"`
	} `json:"displayName"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Rating                float64  `json:"rating"`
	UserRatingCount       int      `json:"userRatingCount"`
	PriceLevel            string   `json:"priceLevel"`
	ShortFormattedAddress string   `json:"shortFormattedAddress"`
	FormattedAddress      string   `json:"formattedAddress"`
	Types                 []string `json:"types"`
	Photos                []struct {
		Name string `json:"name"`
	} `json:"photos"`
	CurrentOpeningHours struct {
		OpenNow             bool     `json:"openNow"`
		WeekdayDescriptions []string `json:"weekdayDescriptions"`
	} `json:"currentOpeningHours"`
	EditorialSummary *struct {
		Text string `json:"text"`
	} `json:"editorialSummary"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *PlacesService) convertPlace(r newPlaceResult, placeType string, originLat, originLng float64) model.PlaceResult {
	addr := r.ShortFormattedAddress
	if addr == "" {
		addr = r.FormattedAddress
	}
	pType := placeType
	if pType == "" && len(r.Types) > 0 {
		pType = r.Types[0]
	}

	place := model.PlaceResult{
		ID:      r.ID,
		Name:    r.DisplayName.Text,
		Lat:     r.Location.Latitude,
		Lng:     r.Location.Longitude,
		Address: addr,
		Type:    pType,
	}
	if r.Rating > 0 {
		v := r.Rating
		place.Rating = &v
	}
	if r.UserRatingCount > 0 {
		v := r.UserRatingCount
		place.ReviewCount = &v
	}
	place.PriceLevel = newPriceLevelStr(r.PriceLevel)
	if len(r.Photos) > 0 {
		place.ImageUrl = buildPhotoURL(r.Photos[0].Name, s.apiKey)
	}
	if r.CurrentOpeningHours.OpenNow || len(r.CurrentOpeningHours.WeekdayDescriptions) > 0 {
		v := r.CurrentOpeningHours.OpenNow
		place.OpenNow = &v
		if v {
			place.Hours = "Open now"
		} else {
			place.Hours = "Closed"
		}
	}
	if originLat != 0 || originLng != 0 {
		place.Distance = formatDistance(haversineKm(originLat, originLng, r.Location.Latitude, r.Location.Longitude))
	}
	return place
}

func buildPhotoURL(photoName, apiKey string) string {
	return fmt.Sprintf("%s/%s/media?maxWidthPx=800&key=%s", placesBaseURL, photoName, apiKey)
}

func newPriceLevelStr(level string) string {
	switch level {
	case "PRICE_LEVEL_FREE":
		return ""
	case "PRICE_LEVEL_INEXPENSIVE":
		return "$"
	case "PRICE_LEVEL_MODERATE":
		return "$$"
	case "PRICE_LEVEL_EXPENSIVE":
		return "$$$"
	case "PRICE_LEVEL_VERY_EXPENSIVE":
		return "$$$$"
	default:
		return ""
	}
}

// todayHours extracts the current weekday's hours string from weekdayDescriptions.
// weekdayDescriptions is Monday-indexed: [Mon, Tue, Wed, Thu, Fri, Sat, Sun].
func todayHours(weekdayText []string) string {
	idx := (int(time.Now().Weekday()) + 6) % 7
	if idx >= len(weekdayText) {
		return ""
	}
	text := weekdayText[idx]
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

func floatPtr(v float64) *float64 { return &v }
func intPtr(v int) *int           { return &v }
func boolPtr(v bool) *bool        { return &v }

var placeholderPlaces = []model.PlaceResult{
	// tourist_attraction
	{ID: "placeholder_ta_1", Name: "Notre-Dame Cathedral Basilica", Lat: 10.7797, Lng: 106.6990, Type: "tourist_attraction", ImageUrl: "https://picsum.photos/seed/placeholder_ta_1/800/600", Rating: floatPtr(4.6), ReviewCount: intPtr(12480), Description: "Iconic 19th-century French colonial cathedral in the heart of the city.", Address: "Công xã Paris, Bến Nghé, Quận 1, TP.HCM", Hours: "8AM – 5PM", OpenNow: boolPtr(true), Distance: "0.4 km"},
	{ID: "placeholder_ta_2", Name: "Reunification Palace", Lat: 10.7771, Lng: 106.6956, Type: "tourist_attraction", ImageUrl: "https://picsum.photos/seed/placeholder_ta_2/800/600", Rating: floatPtr(4.5), ReviewCount: intPtr(9320), Description: "Historic presidential palace where the Vietnam War officially ended in 1975.", Address: "135 Nam Kỳ Khởi Nghĩa, Bến Thành, Quận 1, TP.HCM", Hours: "7:30AM – 5PM", OpenNow: boolPtr(true), Distance: "0.8 km"},
	{ID: "placeholder_ta_3", Name: "War Remnants Museum", Lat: 10.7794, Lng: 106.6919, Type: "tourist_attraction", ImageUrl: "https://picsum.photos/seed/placeholder_ta_3/800/600", Rating: floatPtr(4.7), ReviewCount: intPtr(21050), Description: "Powerful museum documenting the Vietnam War through photographs and artefacts.", Address: "28 Võ Văn Tần, Phường 6, Quận 3, TP.HCM", Hours: "7:30AM – 6PM", OpenNow: boolPtr(true), Distance: "1.2 km"},
	{ID: "placeholder_ta_4", Name: "Ben Thanh Market", Lat: 10.7721, Lng: 106.6981, Type: "tourist_attraction", ImageUrl: "https://picsum.photos/seed/placeholder_ta_4/800/600", Rating: floatPtr(4.2), ReviewCount: intPtr(34100), Description: "Bustling landmark market selling street food, souvenirs, and local goods.", Address: "Lê Lợi, Bến Thành, Quận 1, TP.HCM", Hours: "6AM – 6PM", OpenNow: boolPtr(true), Distance: "0.6 km"},
	{ID: "placeholder_ta_5", Name: "Saigon Central Post Office", Lat: 10.7798, Lng: 106.6998, Type: "tourist_attraction", ImageUrl: "https://picsum.photos/seed/placeholder_ta_5/800/600", Rating: floatPtr(4.5), ReviewCount: intPtr(8760), Description: "Stunning French colonial post office designed by Gustave Eiffel.", Address: "2 Công xã Paris, Bến Nghé, Quận 1, TP.HCM", Hours: "7AM – 7PM", OpenNow: boolPtr(true), Distance: "0.5 km"},

	// restaurant
	{ID: "placeholder_rs_1", Name: "Pho 24", Lat: 10.7749, Lng: 106.7019, Type: "restaurant", ImageUrl: "https://picsum.photos/seed/placeholder_rs_1/800/600", Rating: floatPtr(4.3), ReviewCount: intPtr(5640), PriceLevel: "$", Description: "Popular chain serving classic Vietnamese pho in a clean, air-conditioned setting.", Address: "5 Nguyễn Thiệp, Bến Nghé, Quận 1, TP.HCM", Hours: "6AM – 10PM", OpenNow: boolPtr(true), Distance: "0.3 km"},
	{ID: "placeholder_rs_2", Name: "The Deck Saigon", Lat: 10.8031, Lng: 106.7305, Type: "restaurant", ImageUrl: "https://picsum.photos/seed/placeholder_rs_2/800/600", Rating: floatPtr(4.5), ReviewCount: intPtr(2890), PriceLevel: "$$$", Description: "Riverside restaurant with stunning views and an eclectic international menu.", Address: "38 Nguyễn U Dĩ, Thảo Điền, Quận 2, TP.HCM", Hours: "11AM – 11PM", OpenNow: boolPtr(true), Distance: "4.1 km"},
	{ID: "placeholder_rs_3", Name: "Cục Gạch Quán", Lat: 10.7839, Lng: 106.6945, Type: "restaurant", ImageUrl: "https://picsum.photos/seed/placeholder_rs_3/800/600", Rating: floatPtr(4.6), ReviewCount: intPtr(4120), PriceLevel: "$$", Description: "Charming heritage house restaurant serving authentic Vietnamese home-cooking.", Address: "10 Đặng Tất, Tân Định, Quận 1, TP.HCM", Hours: "11AM – 10PM", OpenNow: boolPtr(true), Distance: "1.5 km"},
	{ID: "placeholder_rs_4", Name: "Nhà Hàng Ngon", Lat: 10.7779, Lng: 106.6939, Type: "restaurant", ImageUrl: "https://picsum.photos/seed/placeholder_rs_4/800/600", Rating: floatPtr(4.4), ReviewCount: intPtr(7830), PriceLevel: "$$", Description: "Garden restaurant with live street-food stalls covering every Vietnamese region.", Address: "160 Pasteur, Bến Nghé, Quận 1, TP.HCM", Hours: "7AM – 10PM", OpenNow: boolPtr(true), Distance: "0.9 km"},
	{ID: "placeholder_rs_5", Name: "Bun Bo Hue An Nam", Lat: 10.7701, Lng: 106.6935, Type: "restaurant", ImageUrl: "https://picsum.photos/seed/placeholder_rs_5/800/600", Rating: floatPtr(4.4), ReviewCount: intPtr(3210), PriceLevel: "$", Description: "Beloved local spot for spicy Hue-style beef noodle soup.", Address: "14 Tôn Thất Tùng, Phạm Ngũ Lão, Quận 1, TP.HCM", Hours: "6AM – 2PM", OpenNow: boolPtr(false), Distance: "1.1 km"},

	// amusement_park
	{ID: "placeholder_ap_1", Name: "Dam Sen Water Park", Lat: 10.7554, Lng: 106.6472, Type: "amusement_park", ImageUrl: "https://picsum.photos/seed/placeholder_ap_1/800/600", Rating: floatPtr(4.1), ReviewCount: intPtr(18200), PriceLevel: "$$", Description: "Large water park with slides, wave pools, and family attractions.", Address: "3 Hòa Bình, Phường 3, Quận 11, TP.HCM", Hours: "8AM – 6PM", OpenNow: boolPtr(true), Distance: "5.8 km"},
	{ID: "placeholder_ap_2", Name: "Suoi Tien Theme Park", Lat: 10.8623, Lng: 106.8337, Type: "amusement_park", ImageUrl: "https://picsum.photos/seed/placeholder_ap_2/800/600", Rating: floatPtr(4.0), ReviewCount: intPtr(22400), PriceLevel: "$$", Description: "Vast cultural theme park blending Vietnamese folklore with rides and water attractions.", Address: "Xa lộ Hà Nội, Hiệp Phú, Quận 9, TP.HCM", Hours: "8AM – 6PM", OpenNow: boolPtr(true), Distance: "14.2 km"},
	{ID: "placeholder_ap_3", Name: "VinKE Times City", Lat: 10.7735, Lng: 106.6593, Type: "amusement_park", ImageUrl: "https://picsum.photos/seed/placeholder_ap_3/800/600", Rating: floatPtr(4.3), ReviewCount: intPtr(6700), PriceLevel: "$$$", Description: "Indoor entertainment complex with interactive science exhibits for children.", Address: "458 Minh Khai, Quận 11, TP.HCM", Hours: "9AM – 9PM", OpenNow: boolPtr(true), Distance: "3.5 km"},

	// bar
	{ID: "placeholder_br_1", Name: "Chill Skybar", Lat: 10.7748, Lng: 106.7028, Type: "bar", ImageUrl: "https://picsum.photos/seed/placeholder_br_1/800/600", Rating: floatPtr(4.3), ReviewCount: intPtr(6540), PriceLevel: "$$$", Description: "Rooftop bar on the 26th floor with panoramic city views and craft cocktails.", Address: "AB Tower, 76A Lê Lai, Bến Thành, Quận 1, TP.HCM", Hours: "5PM – 2AM", OpenNow: boolPtr(true), Distance: "0.4 km"},
	{ID: "placeholder_br_2", Name: "The Observatory", Lat: 10.7679, Lng: 106.6971, Type: "bar", ImageUrl: "https://picsum.photos/seed/placeholder_br_2/800/600", Rating: floatPtr(4.5), ReviewCount: intPtr(3120), PriceLevel: "$$", Description: "Intimate craft beer and cocktail bar with a curated vinyl music experience.", Address: "5 Nguyễn Siêu, Bến Nghé, Quận 1, TP.HCM", Hours: "6PM – 2AM", OpenNow: boolPtr(false), Distance: "1.4 km"},
	{ID: "placeholder_br_3", Name: "EON Heli Bar", Lat: 10.7729, Lng: 106.7034, Type: "bar", ImageUrl: "https://picsum.photos/seed/placeholder_br_3/800/600", Rating: floatPtr(4.4), ReviewCount: intPtr(4880), PriceLevel: "$$$", Description: "Upscale helipad bar atop the Bitexco Tower with 360° views of the city.", Address: "2 Hải Triều, Bến Nghé, Quận 1, TP.HCM", Hours: "11AM – 1AM", OpenNow: boolPtr(true), Distance: "0.7 km"},
	{ID: "placeholder_br_4", Name: "Pasteur Street Brewing Co.", Lat: 10.7771, Lng: 106.6981, Type: "bar", ImageUrl: "https://picsum.photos/seed/placeholder_br_4/800/600", Rating: floatPtr(4.5), ReviewCount: intPtr(5980), PriceLevel: "$$", Description: "Pioneer craft brewery offering bold Vietnamese-ingredient beers on tap.", Address: "144 Pasteur, Bến Nghé, Quận 1, TP.HCM", Hours: "11AM – 11:30PM", OpenNow: boolPtr(true), Distance: "0.9 km"},

	// lodging
	{ID: "placeholder_lg_1", Name: "Park Hyatt Saigon", Lat: 10.7772, Lng: 106.7025, Type: "lodging", ImageUrl: "https://picsum.photos/seed/placeholder_lg_1/800/600", Rating: floatPtr(4.8), ReviewCount: intPtr(4230), PriceLevel: "$$$$", PricePerNight: "$320/night", Description: "Elegant colonial-style luxury hotel on Lam Son Square with world-class dining.", Address: "2 Lam Sơn Square, Bến Nghé, Quận 1, TP.HCM", Hours: "Open 24 hours", OpenNow: boolPtr(true), Distance: "0.5 km"},
	{ID: "placeholder_lg_2", Name: "The Myst Dong Khoi", Lat: 10.7783, Lng: 106.7011, Type: "lodging", ImageUrl: "https://picsum.photos/seed/placeholder_lg_2/800/600", Rating: floatPtr(4.6), ReviewCount: intPtr(2870), PriceLevel: "$$$", PricePerNight: "$150/night", Description: "Boutique heritage hotel housed in a restored 1930s French colonial building.", Address: "6-8 Hồ Huấn Nghiệp, Bến Nghé, Quận 1, TP.HCM", Hours: "Open 24 hours", OpenNow: boolPtr(true), Distance: "0.6 km"},
	{ID: "placeholder_lg_3", Name: "Caravelle Saigon", Lat: 10.7762, Lng: 106.7020, Type: "lodging", ImageUrl: "https://picsum.photos/seed/placeholder_lg_3/800/600", Rating: floatPtr(4.7), ReviewCount: intPtr(5610), PriceLevel: "$$$$", PricePerNight: "$210/night", Description: "Iconic 5-star hotel with a rich history and the legendary Saigon Saigon Rooftop Bar.", Address: "19-23 Lam Sơn Square, Bến Nghé, Quận 1, TP.HCM", Hours: "Open 24 hours", OpenNow: boolPtr(true), Distance: "0.4 km"},
	{ID: "placeholder_lg_4", Name: "Bui Vien Boutique Hostel", Lat: 10.7669, Lng: 106.6937, Type: "lodging", ImageUrl: "https://picsum.photos/seed/placeholder_lg_4/800/600", Rating: floatPtr(4.3), ReviewCount: intPtr(1890), PriceLevel: "$", PricePerNight: "$18/night", Description: "Lively backpacker hostel in the heart of the Bui Vien walking street area.", Address: "245 Đề Thám, Phạm Ngũ Lão, Quận 1, TP.HCM", Hours: "Open 24 hours", OpenNow: boolPtr(true), Distance: "1.8 km"},
}
