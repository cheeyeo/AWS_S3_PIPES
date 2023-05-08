package pipes

import "context"

type Pipe interface {
	Stream(ctx context.Context, pipe string, bucket string, key string) error
}

type PipeWithCancellation struct {
	pipe Pipe
}

func (c *PipeWithCancellation) Stream(ctx context.Context, pipe string, bucket string, key string) error {

	errChan := make(chan error)

	go func() {
		err := c.pipe.Stream(ctx, pipe, bucket, key)
		errChan <- err
	}()
	defer close(errChan)

	for {
		select {
		case err2 := <-errChan:
			// fmt.Printf("ERR OCCURED: %v\n", err2)
			return err2
		case <-ctx.Done():
			// fmt.Println("FUN CANCELLED")
			return ctx.Err()
		}
	}
}

func NewPipeWithCancellation(p Pipe) *PipeWithCancellation {
	return &PipeWithCancellation{pipe: p}
}
