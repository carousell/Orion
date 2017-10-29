package data

import "context"

type StorageService interface {
	Close()
	AddComment(ctx context.Context, comment string) (UUID string, err error)
	GetComment(ctx context.Context, uuid string) (comment string, err error)
	SearchComments(ctx context.Context, term string) (comments []Message, err error)
}
