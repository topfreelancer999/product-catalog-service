package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "product-catalog-service/proto/product/v1"
)

// CreateProduct implements the CreateProduct gRPC method.
func (h *ProductHandler) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductReply, error) {
	// 1. Validate proto request
	if err := validateCreateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq := mapToCreateProductRequest(req)

	// 3. Call usecase (usecase applies plan internally)
	productID, err := h.commands.CreateProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &productv1.CreateProductReply{
		ProductId: productID,
	}, nil
}

func validateCreateRequest(req *productv1.CreateProductRequest) error {
	if req.Name == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Category == "" {
		return status.Error(codes.InvalidArgument, "category is required")
	}
	if req.BasePriceDenominator <= 0 {
		return status.Error(codes.InvalidArgument, "base_price_denominator must be > 0")
	}
	return nil
}
