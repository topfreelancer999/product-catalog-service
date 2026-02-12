package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "product-catalog-service/proto/product/v1"
)

// ActivateProduct implements the ActivateProduct gRPC method.
func (h *ProductHandler) ActivateProduct(ctx context.Context, req *productv1.ActivateProductRequest) (*productv1.ActivateProductReply, error) {
	// 1. Validate proto request
	if err := validateActivateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq := mapToActivateProductRequest(req)

	// 3. Call usecase (usecase applies plan internally)
	if err := h.commands.ActivateProduct.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &productv1.ActivateProductReply{}, nil
}

func validateActivateRequest(req *productv1.ActivateProductRequest) error {
	if req.ProductId == "" {
		return status.Error(codes.InvalidArgument, "product_id is required")
	}
	return nil
}
