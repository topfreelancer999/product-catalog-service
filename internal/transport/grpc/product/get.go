package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "product-catalog-service/proto/product/v1"
)

// GetProduct implements the GetProduct gRPC method.
func (h *ProductHandler) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductReply, error) {
	// 1. Validate proto request
	if err := validateGetRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq := mapToGetProductRequest(req)

	// 3. Call query
	product, err := h.queries.GetProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &productv1.GetProductReply{
		Product: mapProductDTOToProto(product),
	}, nil
}

func validateGetRequest(req *productv1.GetProductRequest) error {
	if req.ProductId == "" {
		return status.Error(codes.InvalidArgument, "product_id is required")
	}
	return nil
}
