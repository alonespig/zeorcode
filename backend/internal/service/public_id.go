package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	"github.com/alonespig/zeorcode/backend/pkg/publicid"
)

const publicIDGenerateAttempts = 16

type publicIDExistsFunc func(context.Context, int64) (bool, error)

func newUniquePublicID(ctx context.Context, exists publicIDExistsFunc) (int64, error) {
	for range publicIDGenerateAttempts {
		id, err := publicid.New()
		if err != nil {
			return 0, errcode.ErrInternal.Wrap(fmt.Errorf("generate public ID: %w", err))
		}
		used, err := exists(ctx, id)
		if err != nil {
			return 0, errcode.ErrDatabase.Wrap(err)
		}
		if !used {
			return id, nil
		}
	}
	return 0, errcode.ErrInternal.Wrap(errors.New("could not allocate a unique public ID"))
}
