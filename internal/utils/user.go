package utils

import (
	"context"

	"github.com/viktoralyoshin/utils/pkg/errs"
	"google.golang.org/grpc/metadata"
)

func GetUserID(ctx context.Context) (string, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return "", errs.ErrInvalidMetadata
	}

	values := md.Get("x-user-id")
	if len(values) == 0 {
		return "", errs.ErrMetadataNotFound
	}

	return values[0], nil
}
