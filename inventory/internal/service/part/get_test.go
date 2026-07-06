package part

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
	"github.com/horizoonn/factory-platform/inventory/internal/service/mocks"
)

func TestService_GetPart(t *testing.T) {
	tests := []struct {
		name       string
		partID     uuid.UUID
		setupMocks func(ctx context.Context, repository *mocks.Repository)
		want       domain.Part
		wantErr    error
	}{
		{
			name:   "success",
			partID: partID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				part := validPart()

				repository.EXPECT().
					GetPart(ctx, partID).
					Return(part, nil).
					Once()
			},
			want: validPart(),
		},
		{
			name:    "error empty part id",
			partID:  uuid.Nil,
			wantErr: domain.ErrPartIDRequired,
		},
		{
			name:   "error get part from repository",
			partID: partID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				repository.EXPECT().
					GetPart(ctx, partID).
					Return(domain.Part{}, errRepository).
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

			got, err := service.GetPart(ctx, tt.partID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, domain.Part{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetPart_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := mocks.NewRepository(t)
	service := NewService(repository)

	got, err := service.GetPart(ctx, partID)

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, domain.Part{}, got)
}
