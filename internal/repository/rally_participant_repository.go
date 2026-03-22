package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RallyParticipantRepository interface {
	CreateParticipant(ctx context.Context, participant *model.RallyParticipant) error
	GetParticipant(ctx context.Context, participantID string) (*model.RallyParticipant, error)
	GetParticipantByRallyAndUser(ctx context.Context, rallyID, userID primitive.ObjectID) (*model.RallyParticipant, error)
	UpdateParticipant(ctx context.Context, participantID string, updates *model.UpdateParticipantRequest) (*model.RallyParticipant, error)
	GetParticipantsList(ctx context.Context, rallyID primitive.ObjectID, role string, page, pageSize int) ([]model.RallyParticipantDetailResponse, int64, error)
	CountJoinedParticipants(ctx context.Context, rallyID primitive.ObjectID) (int64, error)
	GetPendingInvitations(ctx context.Context, userID primitive.ObjectID) ([]model.PendingInvitationItem, error)
}

type rallyParticipantRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewRallyParticipantRepository(db *mongo.Database) RallyParticipantRepository {
	return &rallyParticipantRepository{
		db:         db,
		collection: db.Collection("rally_participants"),
	}
}

func (r *rallyParticipantRepository) CreateParticipant(ctx context.Context, participant *model.RallyParticipant) error {
	if participant.ID.IsZero() {
		participant.ID = primitive.NewObjectID()
	}

	if participant.InvitedAt.IsZero() {
		participant.InvitedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, participant)
	return err
}

func (r *rallyParticipantRepository) GetParticipant(ctx context.Context, participantID string) (*model.RallyParticipant, error) {
	objectID, err := primitive.ObjectIDFromHex(participantID)
	if err != nil {
		return nil, err
	}

	var participant model.RallyParticipant
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&participant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &participant, nil
}

func (r *rallyParticipantRepository) GetParticipantByRallyAndUser(ctx context.Context, rallyID, userID primitive.ObjectID) (*model.RallyParticipant, error) {
	var participant model.RallyParticipant
	err := r.collection.FindOne(ctx, bson.M{
		"rally_id": rallyID,
		"user_id":  userID,
	}).Decode(&participant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &participant, nil
}

func (r *rallyParticipantRepository) UpdateParticipant(ctx context.Context, participantID string, updates *model.UpdateParticipantRequest) (*model.RallyParticipant, error) {
	objectID, err := primitive.ObjectIDFromHex(participantID)
	if err != nil {
		return nil, err
	}

	updateDoc := bson.M{}

	if updates.Role != nil {
		updateDoc["role"] = *updates.Role
	}
	if updates.Status != nil {
		updateDoc["status"] = *updates.Status
		if *updates.Status == "joined" {
			now := time.Now()
			updateDoc["joined_at"] = now
		}
	}

	if len(updateDoc) == 0 {
		return r.GetParticipant(ctx, participantID)
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return nil, err
	}

	return r.GetParticipant(ctx, participantID)
}

// GetParticipantsList retrieves a paginated list of participants for a given rally, including user and inviter information.
func (r *rallyParticipantRepository) GetParticipantsList(ctx context.Context, rallyID primitive.ObjectID, role string, page, pageSize int) ([]model.RallyParticipantDetailResponse, int64, error) {
	skip := (page - 1) * pageSize

	matchFilter := bson.M{"rally_id": rallyID}
	if role != "" {
		matchFilter["role"] = role
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "user_id",
			"foreignField": "_id",
			"as":           "user_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$user_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "invited_by",
			"foreignField": "_id",
			"as":           "inviter_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$inviter_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "joined_at", Value: -1},
			{Key: "invited_at", Value: -1},
		}}},
		{{Key: "$facet", Value: bson.M{
			"metadata": []bson.M{{"$count": "total"}},
			"data": []bson.M{
				{"$skip": skip},
				{"$limit": pageSize},
			},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	type userInfoBson struct {
		ID        primitive.ObjectID `bson:"_id"`
		Username  string             `bson:"username"`
		FirstName string             `bson:"first_name"`
		LastName  string             `bson:"last_name"`
		AvatarUrl string             `bson:"avatar_url"`
	}

	type rawParticipant struct {
		ID          primitive.ObjectID        `bson:"_id"`
		RallyID     primitive.ObjectID        `bson:"rally_id"`
		Role        model.ParticipantRole     `bson:"role"`
		Status      model.ParticipationStatus `bson:"status"`
		JoinedAt    *time.Time                `bson:"joined_at"`
		InvitedAt   time.Time                 `bson:"invited_at"`
		UserInfo    userInfoBson              `bson:"user_info"`
		InviterInfo *userInfoBson             `bson:"inviter_info"`
	}

	type FacetResult struct {
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
		Data []rawParticipant `bson:"data"`
	}

	var results []FacetResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 {
		return []model.RallyParticipantDetailResponse{}, 0, nil
	}

	total := int64(0)
	if len(results[0].Metadata) > 0 {
		total = results[0].Metadata[0].Total
	}

	data := results[0].Data
	responses := make([]model.RallyParticipantDetailResponse, len(data))
	for i, raw := range data {
		user := model.ParticipantUserInfo{
			ID:        raw.UserInfo.ID.Hex(),
			Username:  raw.UserInfo.Username,
			FirstName: raw.UserInfo.FirstName,
			LastName:  raw.UserInfo.LastName,
			AvatarUrl: raw.UserInfo.AvatarUrl,
		}

		var inviter *model.ParticipantUserInfo
		if raw.InviterInfo != nil && !raw.InviterInfo.ID.IsZero() {
			inviter = &model.ParticipantUserInfo{
				ID:        raw.InviterInfo.ID.Hex(),
				Username:  raw.InviterInfo.Username,
				FirstName: raw.InviterInfo.FirstName,
				LastName:  raw.InviterInfo.LastName,
				AvatarUrl: raw.InviterInfo.AvatarUrl,
			}
		}

		responses[i] = model.RallyParticipantDetailResponse{
			ID:        raw.ID.Hex(),
			RallyID:   raw.RallyID.Hex(),
			Role:      raw.Role,
			Status:    raw.Status,
			JoinedAt:  raw.JoinedAt,
			InvitedAt: raw.InvitedAt,
			User:      user,
			InvitedBy: inviter,
		}
	}

	return responses, total, nil
}

func (r *rallyParticipantRepository) CountJoinedParticipants(ctx context.Context, rallyID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"rally_id": rallyID,
		"status":   string(model.ParticipationStatusJoined),
	})
}

// GetPendingInvitations retrieves all "invited" participant records for a user, enriched with rally and inviter info.
func (r *rallyParticipantRepository) GetPendingInvitations(ctx context.Context, userID primitive.ObjectID) ([]model.PendingInvitationItem, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"status":  string(model.ParticipationStatusInvited),
		}}},
		// Join rally info
		{{Key: "$lookup", Value: bson.M{
			"from":         "rallies",
			"localField":   "rally_id",
			"foreignField": "_id",
			"as":           "rally_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$rally_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		// Join inviter user info
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "invited_by",
			"foreignField": "_id",
			"as":           "inviter_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$inviter_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		// Count joined participants per rally
		{{Key: "$lookup", Value: bson.M{
			"from": "rally_participants",
			"let":  bson.M{"rid": "$rally_id"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{
					"$expr": bson.M{"$eq": bson.A{"$rally_id", "$$rid"}},
					"status": string(model.ParticipationStatusJoined),
				}}},
				{{Key: "$count", Value: "count"}},
			},
			"as": "member_count_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$member_count_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "invited_at", Value: -1}}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type rallyInfoBson struct {
		ID            primitive.ObjectID `bson:"_id"`
		Name          string             `bson:"name"`
		Description   string             `bson:"description"`
		CoverImageUrl string             `bson:"cover_image_url"`
		StartDate     *time.Time         `bson:"start_date"`
		EndDate       *time.Time         `bson:"end_date"`
	}

	type inviterInfoBson struct {
		ID        primitive.ObjectID `bson:"_id"`
		Username  string             `bson:"username"`
		FirstName string             `bson:"first_name"`
		LastName  string             `bson:"last_name"`
		AvatarUrl string             `bson:"avatar_url"`
	}

	type memberCountBson struct {
		Count int `bson:"count"`
	}

	type rawResult struct {
		ID              primitive.ObjectID    `bson:"_id"`
		RallyID         primitive.ObjectID    `bson:"rally_id"`
		Role            model.ParticipantRole `bson:"role"`
		InvitedAt       time.Time             `bson:"invited_at"`
		RallyInfo       *rallyInfoBson        `bson:"rally_info"`
		InviterInfo     *inviterInfoBson      `bson:"inviter_info"`
		MemberCountInfo *memberCountBson      `bson:"member_count_info"`
	}

	var results []rawResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	items := make([]model.PendingInvitationItem, 0, len(results))
	for _, raw := range results {
		item := model.PendingInvitationItem{
			ParticipantID: raw.ID.Hex(),
			RallyID:       raw.RallyID.Hex(),
			Role:          raw.Role,
			InvitedAt:     raw.InvitedAt,
		}

		if raw.RallyInfo != nil {
			item.RallyName = raw.RallyInfo.Name
			item.Description = raw.RallyInfo.Description
			item.CoverImageUrl = raw.RallyInfo.CoverImageUrl
			item.StartDate = raw.RallyInfo.StartDate
			item.EndDate = raw.RallyInfo.EndDate
		}

		if raw.MemberCountInfo != nil {
			item.MemberCount = raw.MemberCountInfo.Count
		}

		if raw.InviterInfo != nil && !raw.InviterInfo.ID.IsZero() {
			item.InvitedBy = &model.ParticipantUserInfo{
				ID:        raw.InviterInfo.ID.Hex(),
				Username:  raw.InviterInfo.Username,
				FirstName: raw.InviterInfo.FirstName,
				LastName:  raw.InviterInfo.LastName,
				AvatarUrl: raw.InviterInfo.AvatarUrl,
			}
		}

		items = append(items, item)
	}

	return items, nil
}
