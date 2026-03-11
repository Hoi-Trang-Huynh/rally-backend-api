package service

import (
	"context"
	"fmt"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/repository"
	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/utils"
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
		Username:       req.Username,
		AvatarUrl:      req.AvatarUrl,
		AttachmentURLs: req.AttachmentURLs,
		Comment:        req.Comment,
		Categories:     req.Categories,
		Resolved:       false,
	}

	err := s.repo.CreateFeedback(ctx, feedback)
	if err != nil {
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	return feedback, nil
}

func (s *FeedbackService) ListFeedbacks(ctx context.Context, page, pageSize int, username string, categories []string) (*model.FeedbackListResponse, error) {
	page, pageSize = utils.ClampPagination(page, pageSize, 50)

	feedbacks, total, err := s.repo.GetFeedbacks(ctx, page, pageSize, username, categories)
	if err != nil {
		return nil, fmt.Errorf("failed to list feedbacks: %w", err)
	}

	totalPages := utils.CalcTotalPages(total, pageSize)

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
