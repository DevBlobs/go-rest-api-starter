package main

import (
	"context"
	"github.com/DevBlobs/go-rest-api-starter/internal/app"
	"github.com/joho/godotenv"
	"log/slog"
)

func main() {

	_ = godotenv.Load(".env")

	ctx := context.Background()

	externalDeps, err := app.BuildExternalDeps()
	if err != nil {
		slog.Error("Failed to build external dependencies:", err)
		return
	}

	application, err := app.NewApp(ctx, externalDeps)
	if err != nil {
		slog.Error("Failed to initialize app:", err)
		return
	}
	defer application.Close(ctx)

	application.Echo.Logger.Fatal(application.Echo.Start(":8080"))
}
