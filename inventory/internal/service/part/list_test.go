package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
	"github.com/horizoonn/factory-platform/inventory/internal/service/mocks"
)

func TestService_ListParts(t *testing.T) {
	tests := []struct {
		name       string
		filter     domain.PartsFilter
		setupMocks func(ctx context.Context, repository *mocks.Repository)
		want       []domain.Part
		wantErr    error
	}{
		{
			name:   "success",
			filter: validPartsFilter(),
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				filter := validPartsFilter()
				parts := []domain.Part{validPart(), secondPart()}

				repository.EXPECT().
					ListParts(ctx, filter).
					Return(parts, nil).
					Once()
			},
			want: []domain.Part{validPart(), secondPart()},
		},
		{
			name:   "error list parts from repository",
			filter: validPartsFilter(),
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				repository.EXPECT().
					ListParts(ctx, validPartsFilter()).
					Return(nil, errRepository).
					Once()
			},
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repository := mocks.NewRepository(t)
			service := NewService(repository)

			if tt.setupMocks != nil {
				tt.setupMocks(ctx, repository)
			}

			got, err := service.ListParts(ctx, tt.filter)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_ListParts_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := mocks.NewRepository(t)
	service := NewService(repository)

	got, err := service.ListParts(ctx, validPartsFilter())

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, got)
}
