package product

import (
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

// ProductHandler wires gRPC methods to application usecases.
type ProductHandler struct {
	productv1.UnimplementedProductServiceServer

	// Commands
	commands struct {
		CreateProduct   *createproduct.Interactor
		UpdateProduct   *updateproduct.Interactor
		ActivateProduct *activateproduct.Interactor
		DeactivateProduct *deactivateproduct.Interactor
		ApplyDiscount   *applydiscount.Interactor
		RemoveDiscount  *removediscount.Interactor
	}

	// Queries
	queries struct {
		GetProduct  *getproduct.Query
		ListProducts *listproducts.Query
	}
}

// NewProductHandler creates a new ProductHandler with all usecases and queries wired.
func NewProductHandler(
	createProduct *createproduct.Interactor,
	updateProduct *updateproduct.Interactor,
	activateProduct *activateproduct.Interactor,
	deactivateProduct *deactivateproduct.Interactor,
	applyDiscount *applydiscount.Interactor,
	removeDiscount *removediscount.Interactor,
	getProduct *getproduct.Query,
	listProducts *listproducts.Query,
) *ProductHandler {
	return &ProductHandler{
		commands: struct {
			CreateProduct   *createproduct.Interactor
			UpdateProduct   *updateproduct.Interactor
			ActivateProduct *activateproduct.Interactor
			DeactivateProduct *deactivateproduct.Interactor
			ApplyDiscount   *applydiscount.Interactor
			RemoveDiscount  *removediscount.Interactor
		}{
			CreateProduct:   createProduct,
			UpdateProduct:   updateProduct,
			ActivateProduct: activateProduct,
			DeactivateProduct: deactivateProduct,
			ApplyDiscount:   applyDiscount,
			RemoveDiscount:  removeDiscount,
		},
		queries: struct {
			GetProduct  *getproduct.Query
			ListProducts *listproducts.Query
		}{
			GetProduct:  getProduct,
			ListProducts: listProducts,
		},
	}
}
