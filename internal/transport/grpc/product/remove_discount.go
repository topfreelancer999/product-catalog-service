package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "product-catalog-service/proto/product/v1"
)

// RemoveDiscount implements the RemoveDiscount gRPC method.
func (h *ProductHandler) RemoveDiscount(ctx context.Context, req *productv1.RemoveDiscountRequest) (*productv1.RemoveDiscountReply, error) {
	// 1. Validate proto request
	if err := validateRemoveDiscountRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq := mapToRemoveDiscountRequest(req)

	// 3. Call usecase (usecase applies plan internally)
	if err := h.commands.RemoveDiscount.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &productv1.RemoveDiscountReply{}, nil
}

func validateRemoveDiscountRequest(req *productv1.RemoveDiscountRequest) error {
	if req.ProductId == "" {
		return status.Error(codes.InvalidArgument, "product_id is required")
	}
	return nil
}
