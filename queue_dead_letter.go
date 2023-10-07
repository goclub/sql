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

type DeadLetterQueueMessage struct {
	QueueName     string
	ID            uint64    `db:"id" sq:"ignoreInsert"`
	BusinessID    uint64    `db:"business_id"`
	Reason        string    `db:"reason"`
	Handled       bool      `db:"handled"`
	HandledResult string    `db:"handled_result"`
	CreateTime    time.Time `db:"create_time"`
	DefaultLifeCycle
	WithoutSoftDelete
}

func (q *DeadLetterQueueMessage) TableName() string {
	return "queue_" + q.QueueName + "_dead_letter"
}

func (v *DeadLetterQueueMessage) AfterInsert(result Result) error {
	id, err := result.LastInsertUint64Id()
	if err != nil {
		return err
	}
	v.ID = id
	return nil
}
