package usecase

import "context"

type ResourceGroupManager interface {
	CreateResourceGroup(ctx context.Context, rgName, location string) error
	DeleteResourceGroup(ctx context.Context, rgName string) error
}
