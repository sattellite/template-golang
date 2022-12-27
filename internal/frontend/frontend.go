package frontend

import (
	"context"
	"sync"
)

type Frontend interface {
	Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup)
}
