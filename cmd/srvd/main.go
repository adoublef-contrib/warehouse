package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/adoublef-contrib/warehouse-management/internal/category"
	"github.com/adoublef-contrib/warehouse-management/internal/net/http"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

func ConnectToDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	rwc, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to return potgres pool: %w", err)
	}
	return rwc, nil
}

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, unix.SIGINT, unix.SIGKILL, unix.SIGTERM)
	defer cancel()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL not set")
	}

	conn, err := ConnectToDB(ctx, connStr)
	if err != nil {
		log.Fatal("APP_PORT not assigned")
	}
	defer conn.Close()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	s := &http.Server{
		Addr:    ":" + port,
		Handler: http.Handler(&category.DB{RWC: conn}),
	}
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.ListenAndServe()
	})
	g.Go(func() error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return s.Shutdown(ctx)
	})

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
