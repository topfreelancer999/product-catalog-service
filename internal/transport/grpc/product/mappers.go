package product

import (
	"time"

	productv1 "product-catalog-service/proto/product/v1"
	createproduct "product-catalog-service/internal/app/product/usecases/create_product"
	updateproduct "product-catalog-service/internal/app/product/usecases/update_product"
	activateproduct "product-catalog-service/internal/app/product/usecases/activate_product"
	deactivateproduct "product-catalog-service/internal/app/product/usecases/deactivate_product"
	applydiscount "product-catalog-service/internal/app/product/usecases/apply_discount"
	removediscount "product-catalog-service/internal/app/product/usecases/remove_discount"
	"product-catalog-service/internal/app/product/queries/getproduct"
	"product-catalog-service/internal/app/product/queries/listproducts"
)

// Command mappers: Proto -> Application Request

func mapToCreateProductRequest(req *productv1.CreateProductRequest) createproduct.Request {
	return createproduct.Request{
		Name:                 req.Name,
		Description:          req.Description,
		Category:             req.Category,
		BasePriceNumerator:   req.BasePriceNumerator,
		BasePriceDenominator: req.BasePriceDenominator,
	}
}

func mapToUpdateProductRequest(req *productv1.UpdateProductRequest) updateproduct.Request {
	appReq := updateproduct.Request{
		ProductID: req.ProductId,
	}

	if req.Name != nil {
		appReq.Name = req.Name
	}
	if req.Description != nil {
		appReq.Description = req.Description
	}
	if req.Category != nil {
		appReq.Category = req.Category
	}

	return appReq
}

func mapToActivateProductRequest(req *productv1.ActivateProductRequest) activateproduct.Request {
	return activateproduct.Request{
		ProductID: req.ProductId,
	}
}

func mapToDeactivateProductRequest(req *productv1.DeactivateProductRequest) deactivateproduct.Request {
	return deactivateproduct.Request{
		ProductID: req.ProductId,
	}
}

func mapToApplyDiscountRequest(req *productv1.ApplyDiscountRequest) (applydiscount.Request, error) {
	return applydiscount.Request{
		ProductID:            req.ProductId,
		PercentageNumerator:   req.PercentageNumerator,
		PercentageDenominator: req.PercentageDenominator,
		StartDate:             req.StartDate.AsTime(),
		EndDate:               req.EndDate.AsTime(),
	}, nil
}

func mapToRemoveDiscountRequest(req *productv1.RemoveDiscountRequest) removediscount.Request {
	return removediscount.Request{
		ProductID: req.ProductId,
	}
}

// Query mappers: Proto -> Application Request

func mapToGetProductRequest(req *productv1.GetProductRequest) getproduct.Request {
	return getproduct.Request{
		ProductID: req.ProductId,
		Now:       time.Time{}, // Will use current time in query
	}
}

func mapToListProductsRequest(req *productv1.ListProductsRequest) listproducts.Request {
	appReq := listproducts.Request{
		PageSize:  int(req.PageSize),
		PageToken: req.PageToken,
		Now:       time.Time{}, // Will use current time in query
	}

	if req.Category != nil {
		appReq.Category = req.Category
	}

	return appReq
}

// Response mappers: Application DTO -> Proto

func mapProductDTOToProto(dto *getproduct.ProductDTO) *productv1.Product {
	return &productv1.Product{
		ProductId:      dto.ID,
		Name:           dto.Name,
		Description:    dto.Description,
		Category:       dto.Category,
		Status:         dto.Status,
		EffectivePrice: mapMoneyToProto(dto.EffectivePriceNumerator, dto.EffectivePriceDenominator),
	}
}

func mapProductListItemDTOToProto(dto listproducts.ProductListItemDTO) *productv1.ProductListItem {
	return &productv1.ProductListItem{
		ProductId:      dto.ID,
		Name:           dto.Name,
		Category:       dto.Category,
		Status:         dto.Status,
		EffectivePrice: mapMoneyToProto(dto.EffectivePriceNumerator, dto.EffectivePriceDenominator),
	}
}

func mapMoneyToProto(numerator, denominator int64) *productv1.Money {
	return &productv1.Money{
		Numerator:   numerator,
		Denominator: denominator,
	}
}
