package database

import (
	"context"
	"sync"
)

type Database interface {
	Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup)
	// TODO: add methods
}
