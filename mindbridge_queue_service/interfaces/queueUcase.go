package interfaces

import (
	"context"
	"mindbridge_queue_service/models"
)

type QueueUcaseInterface interface {
	HandleQueue(ctx context.Context, conferenceUUID, conferenceName string) (*models.QueueApiResponse, error)
	HandleEntryQueue(ctx context.Context, conferenceUUID, conferenceName string) (*models.QueueApiResponse, error)
	DeleteQueue(ctx context.Context) (*models.QueueApiResponse, error)
}
