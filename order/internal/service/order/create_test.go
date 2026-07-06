package order

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
	"github.com/horizoonn/factory-platform/order/internal/service/mocks"
)

func TestService_CreateOrder(t *testing.T) {
	tests := []struct {
		name       string
		req        servicedto.CreateOrderRequest
		setupMocks func(ctx context.Context, repository *mocks.Repository, inventoryClient *mocks.InventoryClient)
		want       domain.Order
		wantErr    error
	}{
		{
			name: "success",
			req:  validCreateOrderRequest(),
			setupMocks: func(ctx context.Context, repository *mocks.Repository, inventoryClient *mocks.InventoryClient) {
				part := validPart()
				order := validOrder()

				inventoryClient.EXPECT().
					ListParts(ctx, []uuid.UUID{partID}).
					Return([]domain.Part{part}, nil).
					Once()

				repository.EXPECT().
					CreateOrder(ctx, order).
					Return(order, nil).
					Once()
			},
			want: validOrder(),
		},
		{
			name: "success duplicate part ids",
			req: servicedto.CreateOrderRequest{
				UserID:  userID,
				PartIDs: []uuid.UUID{partID, partID},
			},
			setupMocks: func(ctx context.Context, repository *mocks.Repository, inventoryClient *mocks.InventoryClient) {
				part := validPart()
				order := validOrder()
				order.PartIDs = []uuid.UUID{partID, partID}
				order.TotalPrice = part.Price * 2

				inventoryClient.EXPECT().
					ListParts(ctx, []uuid.UUID{partID, partID}).
					Return([]domain.Part{part}, nil).
					Once()

				repository.EXPECT().
					CreateOrder(ctx, order).
					Return(order, nil).
					Once()
			},
			want: domain.Order{
				ID:         orderID,
				UserID:     userID,
				PartIDs:    []uuid.UUID{partID, partID},
				TotalPrice: validPart().Price * 2,
				Status:     domain.OrderStatusPendingPayment,
			},
		},
		{
			name: "error empty user id",
			req: servicedto.CreateOrderRequest{
				PartIDs: []uuid.UUID{partID},
			},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name: "error empty parts",
			req: servicedto.CreateOrderRequest{
				UserID: userID,
			},
			wantErr: domain.ErrEmptyParts,
		},
		{
			name: "error list parts from inventory client",
			req:  validCreateOrderRequest(),
			setupMocks: func(ctx context.Context, _ *mocks.Repository, inventoryClient *mocks.InventoryClient) {
				inventoryClient.EXPECT().
					ListParts(ctx, []uuid.UUID{partID}).
					Return(nil, errClient).
					Once()
			},
			wantErr: errClient,
		},
		{
			name: "error parts not found",
			req:  validCreateOrderRequest(),
			setupMocks: func(ctx context.Context, _ *mocks.Repository, inventoryClient *mocks.InventoryClient) {
				inventoryClient.EXPECT().
					ListParts(ctx, []uuid.UUID{partID}).
					Return(nil, nil).
					Once()
			},
			wantErr: domain.ErrPartsNotFound,
		},
		{
			name: "error create order in repository",
			req:  validCreateOrderRequest(),
			setupMocks: func(ctx context.Context, repository *mocks.Repository, inventoryClient *mocks.InventoryClient) {
				part := validPart()
				order := validOrder()

				inventoryClient.EXPECT().
					ListParts(ctx, []uuid.UUID{partID}).
					Return([]domain.Part{part}, nil).
					Once()

				repository.EXPECT().
					CreateOrder(ctx, order).
					Return(domain.Order{}, errRepository).
					Once()
			},
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repository := mocks.NewRepository(t)
			inventoryClient := mocks.NewInventoryClient(t)
			service := newServiceWithOrderID(repository, inventoryClient, nil, orderID)

			if tt.setupMocks != nil {
				tt.setupMocks(ctx, repository, inventoryClient)
			}

			got, err := service.CreateOrder(ctx, tt.req)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, domain.Order{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_CreateOrder_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := mocks.NewRepository(t)
	inventoryClient := mocks.NewInventoryClient(t)
	service := newServiceWithOrderID(repository, inventoryClient, nil, orderID)

	got, err := service.CreateOrder(ctx, validCreateOrderRequest())

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, domain.Order{}, got)
}
