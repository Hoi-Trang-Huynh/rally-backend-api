package service

import (
	"context"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
)

type FeedbackService struct {
	repo repository.FeedbackRepository
}

func NewFeedbackService(repo repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{
		repo: repo,
	}
}

func (s *FeedbackService) SubmitFeedback(ctx context.Context, req model.CreateFeedbackRequest) (*model.Feedback, error) {
	feedback := &model.Feedback{
		Username:   req.Username,
		AvatarUrl:  req.AvatarUrl,
		ImageURL:   req.ImageURL,
		Comment:    req.Comment,
		Categories: req.Categories,
		Resolved:   false,
	}

	err := s.repo.CreateFeedback(ctx, feedback)
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

func (s *FeedbackService) ListFeedbacks(ctx context.Context, page, pageSize int, username string, categories []string) (*model.FeedbackListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	feedbacks, total, err := s.repo.GetFeedbacks(ctx, page, pageSize, username, categories)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &model.FeedbackListResponse{
		Feedbacks:  feedbacks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *FeedbackService) ResolveFeedback(ctx context.Context, id string, resolved bool) error {
	return s.repo.UpdateFeedbackResolved(ctx, id, resolved)
}
