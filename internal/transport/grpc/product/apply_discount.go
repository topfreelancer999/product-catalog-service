package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "product-catalog-service/proto/product/v1"
)

// ApplyDiscount implements the ApplyDiscount gRPC method.
func (h *ProductHandler) ApplyDiscount(ctx context.Context, req *productv1.ApplyDiscountRequest) (*productv1.ApplyDiscountReply, error) {
	// 1. Validate proto request
	if err := validateApplyDiscountRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq, err := mapToApplyDiscountRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3. Call usecase (usecase applies plan internally)
	if err := h.commands.ApplyDiscount.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &productv1.ApplyDiscountReply{}, nil
}

func validateApplyDiscountRequest(req *productv1.ApplyDiscountRequest) error {
	if req.ProductId == "" {
		return status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.PercentageDenominator <= 0 {
		return status.Error(codes.InvalidArgument, "percentage_denominator must be > 0")
	}
	if req.StartDate == nil {
		return status.Error(codes.InvalidArgument, "start_date is required")
	}
	if req.EndDate == nil {
		return status.Error(codes.InvalidArgument, "end_date is required")
	}
	return nil
}
