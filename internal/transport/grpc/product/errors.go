package product

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"product-catalog-service/internal/app/product/domain"
)

// mapDomainErrorToGRPC maps domain errors to gRPC status errors.
func mapDomainErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	// Check for domain sentinel errors
	if errors.Is(err, domain.ErrProductNotActive) {
		return status.Error(codes.FailedPrecondition, "product is not active")
	}

	if errors.Is(err, domain.ErrInvalidDiscountPeriod) {
		return status.Error(codes.InvalidArgument, "invalid discount period")
	}

	// Check for common error patterns
	if errors.Is(err, errors.New("product not found")) {
		return status.Error(codes.NotFound, "product not found")
	}

	// Default to internal error for unknown errors
	return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
}
