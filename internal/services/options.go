package services

import (
    "cloud.google.com/go/spanner"
    "context"

    // Domain contracts
    "product-catalog-service/internal/app/product/contracts"

    // Repositories
    "product-catalog-service/internal/app/product/repo"

    // Usecases (Commands)
    "product-catalog-service/internal/app/product/usecases/create_product"
    "product-catalog-service/internal/app/product/usecases/update_product"
    "product-catalog-service/internal/app/product/usecases/activate_product"
    "product-catalog-service/internal/app/product/usecases/deactivate_product"
    "product-catalog-service/internal/app/product/usecases/apply_discount"
    "product-catalog-service/internal/app/product/usecases/remove_discount"

    // Queries
    "product-catalog-service/internal/app/product/queries/get_product"
    "product-catalog-service/internal/app/product/queries/list_products"

    // Infrastructure
    "product-catalog-service/internal/pkg/committer"
    "product-catalog-service/internal/pkg/clock"
)

// Options holds all service dependencies
type Options struct {
    // Shared
    Clock     clock.Clock
    Committer *committer.Committer

    // Repositories
    ProductRepo contracts.ProductRepo
    OutboxRepo  contracts.OutboxRepo

    // Usecases (Commands)
    CreateProduct     *create_product.Interactor
    UpdateProduct     *update_product.Interactor
    ActivateProduct   *activate_product.Interactor
    DeactivateProduct *deactivate_product.Interactor
    ApplyDiscount     *apply_discount.Interactor
    RemoveDiscount    *remove_discount.Interactor

    // Queries
    GetProduct   *get_product.Query
    ListProducts *list_products.Query
}

// NewOptions constructs all dependencies
func NewOptions(ctx context.Context, spannerClient *spanner.Client) *Options {
    // Shared infrastructure
    clk := clock.NewRealClock()
    comm := committer.New(spannerClient)

    // Repositories
    prodRepo := repo.NewProductRepo(spannerClient)
    outboxRepo := repo.NewOutboxRepo(spannerClient)

    // Usecases
    createProductUC := create_product.NewInteractor(prodRepo, outboxRepo, comm, clk)
    updateProductUC := update_product.NewInteractor(prodRepo, outboxRepo, comm, clk)
    activateProductUC := activate_product.NewInteractor(prodRepo, outboxRepo, comm, clk)
    deactivateProductUC := deactivate_product.NewInteractor(prodRepo, outboxRepo, comm, clk)
    applyDiscountUC := apply_discount.NewInteractor(prodRepo, outboxRepo, comm, clk)
    removeDiscountUC := remove_discount.NewInteractor(prodRepo, outboxRepo, comm, clk)

    // Queries
    getProductQuery := get_product.NewQuery(prodRepo)
    listProductsQuery := list_products.NewQuery(prodRepo)

    return &Options{
        Clock:            clk,
        Committer:        comm,
        ProductRepo:      prodRepo,
        OutboxRepo:       outboxRepo,
        CreateProduct:    createProductUC,
        UpdateProduct:    updateProductUC,
        ActivateProduct:  activateProductUC,
        DeactivateProduct: deactivateProductUC,
        ApplyDiscount:    applyDiscountUC,
        RemoveDiscount:   removeDiscountUC,
        GetProduct:       getProductQuery,
        ListProducts:     listProductsQuery,
    }
}
