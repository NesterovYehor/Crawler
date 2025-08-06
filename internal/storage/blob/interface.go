package blob

import "context"

type BlobStore interface {
	Save(ctx context.Context, data []byte) error
}
