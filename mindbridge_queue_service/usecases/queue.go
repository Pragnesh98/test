package repository

import (
	"context"
	"mindbridge_queue_service/interfaces"
	"mindbridge_queue_service/models"
)

type QueueUcase struct {
	QueueRepo interfaces.QueueRepoInterface
}

func NewQueueUcase(repo interfaces.QueueRepoInterface) interfaces.QueueRepoInterface {
	return &QueueUcase{
		QueueRepo: repo,
	}
}

func (r *QueueUcase) HandleQueue(ctx context.Context, ConferenceUUID, conferenceName string) (*models.QueueApiResponse, error) {
	return r.QueueRepo.HandleQueue(ctx, ConferenceUUID, conferenceName)
}

func (r *QueueUcase) HandleEntryQueue(ctx context.Context, ConferenceUUID, conferenceName string) (*models.QueueApiResponse, error) {
	return r.QueueRepo.HandleEntryQueue(ctx, ConferenceUUID, conferenceName)
}

func (r *QueueUcase) DeleteQueue(ctx context.Context) (*models.QueueApiResponse, error) {
	return r.QueueRepo.DeleteQueue(ctx)
}
