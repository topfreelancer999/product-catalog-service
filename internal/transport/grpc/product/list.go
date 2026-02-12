package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "product-catalog-service/proto/product/v1"
)

// ListProducts implements the ListProducts gRPC method.
func (h *ProductHandler) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsReply, error) {
	// 1. Validate proto request
	if err := validateListRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 2. Map proto to application request
	appReq := mapToListProductsRequest(req)

	// 3. Call query
	result, err := h.queries.ListProducts.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Map response
	items := make([]*productv1.ProductListItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, mapProductListItemDTOToProto(item))
	}

	// 5. Return response
	return &productv1.ListProductsReply{
		Items:         items,
		NextPageToken: result.NextPageToken,
	}, nil
}

func validateListRequest(req *productv1.ListProductsRequest) error {
	if req.PageSize < 0 {
		return status.Error(codes.InvalidArgument, "page_size must be >= 0")
	}
	if req.PageSize > 1000 {
		return status.Error(codes.InvalidArgument, "page_size must be <= 1000")
	}
	return nil
}
