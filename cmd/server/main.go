package main

import (
    "context"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"

    "cloud.google.com/go/spanner"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    pb "product-catalog-service/proto/product/v1"
    "product-catalog-service/internal/services"
    "product-catalog-service/internal/transport/grpc/product"
)

const (
    defaultGRPCPort      = "50051"
    spannerEmulatorHost  = "localhost:9010" // Make sure docker-compose is running Spanner emulator
    spannerDatabase      = "projects/test-project/instances/test-instance/databases/product_catalog"
)

func main() {
    ctx := context.Background()

    // --- Connect to Spanner ---
    os.Setenv("SPANNER_EMULATOR_HOST", spannerEmulatorHost)

    client, err := spanner.NewClient(ctx, spannerDatabase)
    if err != nil {
        log.Fatalf("failed to create Spanner client: %v", err)
    }
    defer client.Close()

    // --- Initialize all services (DI container) ---
    opts := services.NewOptions(ctx, client)

    // --- Initialize gRPC server ---
    grpcServer := grpc.NewServer()

    // --- Register ProductService handler ---
    handler := product.NewProductHandler(
        opts.CreateProduct,
        opts.UpdateProduct,
        opts.ActivateProduct,
        opts.DeactivateProduct,
        opts.ApplyDiscount,
        opts.RemoveDiscount,
        opts.GetProduct,
        opts.ListProducts,
    )
    pb.RegisterProductServiceServer(grpcServer, handler)

    // Enable reflection for debugging with grpcurl or Evans CLI
    reflection.Register(grpcServer)

    // --- Listen for incoming gRPC requests ---
    port := defaultGRPCPort
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        log.Fatalf("failed to listen on port %s: %v", port, err)
    }
    log.Printf("gRPC server listening on port %s", port)

    // --- Graceful shutdown ---
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
        <-sigCh
        log.Println("Shutting down gRPC server...")
        grpcServer.GracefulStop()
    }()

    // --- Serve ---
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve gRPC server: %v", err)
    }
}
