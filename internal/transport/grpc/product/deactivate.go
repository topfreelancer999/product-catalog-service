package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "product-catalog-service/proto/product/v1"
)

// DeactivateProduct implements the DeactivateProduct gRPC method.
func (h *ProductHandler) DeactivateProduct(ctx context.Context, req *productv1.DeactivateProductRequest) (*productv1.DeactivateProductReply, error) {
	// 1. Validate proto request
	if err := validateDeactivateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq := mapToDeactivateProductRequest(req)

	// 3. Call usecase (usecase applies plan internally)
	if err := h.commands.DeactivateProduct.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &productv1.DeactivateProductReply{}, nil
}

func validateDeactivateRequest(req *productv1.DeactivateProductRequest) error {
	if req.ProductId == "" {
		return status.Error(codes.InvalidArgument, "product_id is required")
	}
	return nil
}
