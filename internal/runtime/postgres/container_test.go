package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	. "github.com/adoublef-contrib/warehouse-management/internal/runtime/postgres"
	"github.com/testcontainers/testcontainers-go"
	"go.adoublef.dev/testing/is"
)

func TestContainer_ConnectionPool(t *testing.T) {
	p, err := container.ConnectionPool(t.Context())
	is.OK(t, err) // Pool
	p.Close()     //
}

func TestMain(m *testing.M) {
	err := setup(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	err = cleanup(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

var container *Container

// setup initialises containers within the pacakge.
func setup(ctx context.Context) (err error) {
	container, err = Run(ctx, "")
	if err != nil {
		return
	}
	return
}

// cleanup stops all running containers for the pacakge.
func cleanup(ctx context.Context) (err error) {
	var cc = []testcontainers.Container{container}
	for _, c := range cc {
		if c != nil {
			err = errors.Join(err, c.Terminate(ctx))
		}
	}
	return err
}
