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

type RallyRepository interface {
	CreateRally(ctx context.Context, rally *model.Rally) error
	GetRallyByID(ctx context.Context, rallyID string) (*model.Rally, error)
	UpdateRally(ctx context.Context, rallyID string, updates *model.UpdateRallyRequest) (*model.Rally, error)
	GetRalliesList(ctx context.Context, userID primitive.ObjectID, nameFilter string, statusFilter string, sortOrder string, page int, pageSize int) ([]model.Rally, int, error)
	AddPlace(ctx context.Context, rallyID string, placeID string) error
	RemovePlace(ctx context.Context, rallyID string, placeID string) error
}

type rallyRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewRallyRepository(db *mongo.Database) RallyRepository {
	return &rallyRepository{
		db:         db,
		collection: db.Collection("rallies"),
	}
}

func (r *rallyRepository) CreateRally(ctx context.Context, rally *model.Rally) error {
	if rally.ID.IsZero() {
		rally.ID = primitive.NewObjectID()
	}

	now := time.Now()
	if rally.CreatedAt.IsZero() {
		rally.CreatedAt = now
	}
	if rally.UpdatedAt.IsZero() {
		rally.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, rally)
	return err
}

func (r *rallyRepository) GetRallyByID(ctx context.Context, rallyID string) (*model.Rally, error) {
	objectID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, err
	}

	var rally model.Rally
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&rally)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &rally, nil
}

func (r *rallyRepository) UpdateRally(ctx context.Context, rallyID string, updates *model.UpdateRallyRequest) (*model.Rally, error) {
	objectID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return nil, err
	}

	updateDoc := bson.M{
		"updated_at": time.Now(),
	}

	if updates.Name != nil {
		updateDoc["name"] = *updates.Name
	}
	if updates.Description != nil {
		updateDoc["description"] = *updates.Description
	}
	if updates.CoverImageUrl != nil {
		updateDoc["cover_image_url"] = *updates.CoverImageUrl
	}
	if updates.Status != nil {
		updateDoc["status"] = *updates.Status
	}
	if updates.StartDate != nil {
		updateDoc["start_date"] = *updates.StartDate
	}
	if updates.EndDate != nil {
		updateDoc["end_date"] = *updates.EndDate
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		return nil, err
	}

	return r.GetRallyByID(ctx, rallyID)
}

func (r *rallyRepository) GetRalliesList(ctx context.Context, userID primitive.ObjectID, nameFilter string, statusFilter string, sortOrder string, page int, pageSize int) ([]model.Rally, int, error) {
	// Build base pipeline for filtering
	basePipeline := []bson.M{}

	// Stage 1: Lookup rally_participants to filter by user participation
	basePipeline = append(basePipeline, bson.M{
		"$lookup": bson.M{
			"from": "rally_participants",
			"let":  bson.M{"rallyId": "$_id"},
			"pipeline": []bson.M{
				{
					"$match": bson.M{
						"$expr": bson.M{
							"$and": []bson.M{
								{"$eq": []interface{}{"$rally_id", "$$rallyId"}},
								{"$eq": []interface{}{"$user_id", userID}},
								{"$eq": []interface{}{"$status", "joined"}},
							},
						},
					},
				},
			},
			"as": "participations",
		},
	})

	// Stage 2: Match only rallies where user has joined
	basePipeline = append(basePipeline, bson.M{
		"$match": bson.M{
			"participations": bson.M{"$ne": []interface{}{}},
		},
	})

	// Stage 3: Build match filters for name and status
	matchStage := bson.M{}
	if nameFilter != "" {
		matchStage["name"] = bson.M{"$regex": nameFilter, "$options": "i"}
	}
	if statusFilter != "" {
		matchStage["status"] = statusFilter
	}
	if len(matchStage) > 0 {
		basePipeline = append(basePipeline, bson.M{"$match": matchStage})
	}

	// Create a facet pipeline to get both count and paginated results
	sortDirection := 1 // 1 for ascending, -1 for descending
	if sortOrder == "desc" {
		sortDirection = -1
	}

	skip := (page - 1) * pageSize

	facetPipeline := append(basePipeline, bson.M{
		"$facet": bson.M{
			"metadata": []bson.M{
				{"$count": "total"},
			},
			"data": []bson.M{
				{"$sort": bson.M{"start_date": sortDirection}},
				{"$skip": skip},
				{"$limit": pageSize},
				{"$project": bson.M{"participations": 0}},
			},
		},
	})

	cursor, err := r.collection.Aggregate(ctx, facetPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Metadata []struct {
			Total int `bson:"total"`
		} `bson:"metadata"`
		Data []model.Rally `bson:"data"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 || len(results[0].Data) == 0 {
		return []model.Rally{}, 0, nil
	}

	total := 0
	if len(results[0].Metadata) > 0 {
		total = results[0].Metadata[0].Total
	}

	return results[0].Data, total, nil
}

func (r *rallyRepository) AddPlace(ctx context.Context, rallyID string, placeID string) error {
	objectID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$addToSet": bson.M{"place_ids": placeID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *rallyRepository) RemovePlace(ctx context.Context, rallyID string, placeID string) error {
	objectID, err := primitive.ObjectIDFromHex(rallyID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$pull": bson.M{"place_ids": placeID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}
