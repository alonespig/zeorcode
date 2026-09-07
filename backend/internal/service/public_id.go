package service

import (
	"context"

	"zoj/internal/common/errcode"
	"zoj/internal/common/publicid"
)

const publicIDGenerateAttempts = 16

type publicIDExistsFunc func(context.Context, int64) (bool, error)

func newUniquePublicID(ctx context.Context, exists publicIDExistsFunc) (int64, error) {
	for range publicIDGenerateAttempts {
		id, err := publicid.New()
		if err != nil {
			return 0, errcode.ErrInternal.WithMsg("生成公开编号失败，请重试").Wrap(err)
		}
		used, err := exists(ctx, id)
		if err != nil {
			return 0, errcode.ErrDatabase.Wrap(err)
		}
		if !used {
			return id, nil
		}
	}
	return 0, errcode.ErrInternal.WithMsg("生成公开编号失败，请重试")
}
