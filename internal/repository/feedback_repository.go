package repository

import (
	"context"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FeedbackRepository interface {
	CreateFeedback(ctx context.Context, feedback *model.Feedback) error
	GetFeedbacks(ctx context.Context, page, pageSize int, username string, categories []string) ([]model.Feedback, int64, error)
	UpdateFeedbackResolved(ctx context.Context, id string, resolved bool) error
}

type feedbackRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewFeedbackRepository(db *mongo.Database) FeedbackRepository {
	return &feedbackRepository{
		db:         db,
		collection: db.Collection("feedbacks"),
	}
}

func (r *feedbackRepository) CreateFeedback(ctx context.Context, feedback *model.Feedback) error {
	if feedback.ID.IsZero() {
		feedback.ID = primitive.NewObjectID()
	}
	now := time.Now()
	if feedback.CreatedAt.IsZero() {
		feedback.CreatedAt = now
	}
	if feedback.UpdatedAt.IsZero() {
		feedback.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, feedback)
	return err
}

func (r *feedbackRepository) GetFeedbacks(ctx context.Context, page, pageSize int, username string, categories []string) ([]model.Feedback, int64, error) {
	filter := bson.M{}

	if username != "" {
		filter["username"] = primitive.Regex{Pattern: username, Options: "i"}
	}

	if len(categories) > 0 {
		filter["categories"] = bson.M{"$in": categories}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var feedbacks []model.Feedback
	if err = cursor.All(ctx, &feedbacks); err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

func (r *feedbackRepository) UpdateFeedbackResolved(ctx context.Context, id string, resolved bool) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"resolved":   resolved,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}
