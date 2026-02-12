package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "product-catalog-service/proto/product/v1"
)

// UpdateProduct implements the UpdateProduct gRPC method.
func (h *ProductHandler) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductReply, error) {
	// 1. Validate proto request
	if err := validateUpdateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq := mapToUpdateProductRequest(req)

	// 3. Call usecase (usecase applies plan internally)
	if err := h.commands.UpdateProduct.Execute(ctx, appReq); err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &productv1.UpdateProductReply{}, nil
}

func validateUpdateRequest(req *productv1.UpdateProductRequest) error {
	if req.ProductId == "" {
		return status.Error(codes.InvalidArgument, "product_id is required")
	}
	// At least one field must be provided
	if req.Name == nil && req.Description == nil && req.Category == nil {
		return status.Error(codes.InvalidArgument, "at least one field (name, description, category) must be provided")
	}
	return nil
}
