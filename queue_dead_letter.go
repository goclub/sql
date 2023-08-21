package sq

import (
	"context"
	"time"
)

type DeadLetterHandler interface {
	// // DeleteDeadLetter 从死信队列中删除消息。
	// DeleteDeadLetter(ctx context.Context, id uint64) (err error)

	// HandleDeadLetter 将死信消息标记为已处理。
	HandleDeadLetter(ctx context.Context, id uint64, remark string) (err error)
	// RequeueDeadLetter 将死信消息重新入队以重新处理。
	RequeueDeadLetter(ctx context.Context, id uint64, publish Publish) (err error)
	// ArchiveHandledDeadLetter 将已处理过的死信消息归档以备将来分析
	ArchiveHandledDeadLetter(ctx context.Context, ago time.Duration) (cleanCount bool, err error)
}
