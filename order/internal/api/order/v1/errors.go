package orderv1

import (
	"context"
	"errors"
	"log/slog"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func (h *Handler) NewError(_ context.Context, err error) *orderopenapi.GenericErrorStatusCode {
	slog.Error("request failed", "error", err)

	statusCode := 500
	message := "internal server error"

	switch {
	case errors.Is(err, domain.ErrNotFound):
		statusCode = 404
		message = "resource not found"
	case errors.Is(err, domain.ErrUserIDRequired),
		errors.Is(err, domain.ErrEmptyParts),
		errors.Is(err, domain.ErrInvalidPaymentMethod),
		errors.Is(err, domain.ErrOrderIDRequired),
		errors.Is(err, domain.ErrPartsNotFound):
		statusCode = 400
		message = err.Error()
	case errors.Is(err, domain.ErrOrderAlreadyPaid),
		errors.Is(err, domain.ErrOrderCancelled),
		errors.Is(err, domain.ErrInvalidOrderStatus):
		statusCode = 409
		message = err.Error()
	case errors.Is(err, domain.ErrNotImplemented):
		statusCode = 501
		message = "not implemented"
	}

	return &orderopenapi.GenericErrorStatusCode{
		StatusCode: statusCode,
		Response: orderopenapi.GenericError{
			Code:    orderopenapi.NewOptInt(statusCode),
			Message: orderopenapi.NewOptString(message),
		},
	}
}

func badRequest(message string) *orderopenapi.BadRequestError {
	return &orderopenapi.BadRequestError{
		Code:    400,
		Message: message,
	}
}

func notFound(message string) *orderopenapi.NotFoundError {
	return &orderopenapi.NotFoundError{
		Code:    404,
		Message: message,
	}
}

func conflict(message string) *orderopenapi.ConflictError {
	return &orderopenapi.ConflictError{
		Code:    409,
		Message: message,
	}
}

func internalError() *orderopenapi.InternalServerError {
	return &orderopenapi.InternalServerError{
		Code:    500,
		Message: "internal server error",
	}
}
