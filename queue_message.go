package sq

import (
	"context"
	"database/sql"
	"time"
)

type MessageQueue struct {
	QueueName string
	ID        uint64 `db:"id" sq:"ignoreInsert|ignoreUpdate"`
	BusinessID uint64 `db:"business_id"`
	NextConsumeTime time.Time `db:"next_consume_time"`
	ConsumeChance uint16 `db:"consume_chance"`
	UpdateID sql.NullString `db:"update_id"`
	Priority uint8 `db:"priority"`
	CreateTime time.Time `db:"create_time"`
	consume Consume
	DefaultLifeCycle
	WithoutSoftDelete
}
func (message *MessageQueue) TableName () string {
	return "queue_" + message.QueueName
}
func (v *MessageQueue) AfterInsert(result Result) error {
	id, err := result.LastInsertUint64Id(); if err != nil {
		return err
	}
	v.ID = uint64(id)
	return nil
}
type DeadLetterQueueMessage struct {
	QueueName string
	ID uint64 `db:"id" sq:"ignoreInsert|ignoreUpdate"`
	BusinessID uint64 `db:"business_id"`
	Reason string `db:"reason"`
	CreateTime time.Time `db:"create_time"`
	DefaultLifeCycle
	WithoutSoftDelete
}

func (q *DeadLetterQueueMessage) TableName () string {
	return "queue_" + q.QueueName + "_dead_letter"
}

func (v *DeadLetterQueueMessage) AfterInsert(result Result) error {
	id, err := result.LastInsertUint64Id(); if err != nil {
		return err
	}
	v.ID = uint64(id)
	return nil
}
func  (message MessageQueue) Ack(ctx context.Context, tx *Transaction) (err error) {
	_, err = tx.HardDelete(ctx, QB{
		From:  &message,
		Where: And("id", Equal(message.ID)),
		Limit: 1,
	}) // indivisible begin
	if err != nil { // indivisible end
		return
	}
	return
}
func  (message MessageQueue) Requeue(ctx context.Context, tx *Transaction) (err error) {
	return message.RequeueWithError(ctx, tx, nil)
}
func  (message MessageQueue) RequeueWithError(ctx context.Context, tx *Transaction, consumeError error) (err error) {
	if message.ConsumeChance == 0 {
		return message.DeadLetterWithError(ctx, tx, "consume chance is zero, can not requeue", consumeError)
	}
	if consumeError != nil {
		message.consume.HandleError(consumeError)
	}
	nextConsumeDuration := message.consume.NextConsumeTime(message.ConsumeChance)
	_, err = tx.Update(ctx, QB{
		From: &message,
		Where: And("id", Equal(message.ID)),
		Set: Set("next_consume_time", time.Now().In(message.consume.queueTimeLocation).Add(nextConsumeDuration)),
		Limit: 1,
	}) // indivisible begin
	if err != nil { // indivisible end
		return
	}
	return
}
func  (message MessageQueue) DeadLetter(ctx context.Context, tx *Transaction, reason string) (err error) {
	return message.DeadLetterWithError(ctx, tx, reason, nil)
}
func  (message MessageQueue) DeadLetterWithError(ctx context.Context, tx *Transaction, reason string, consumeError error) (err error) {
	if consumeError != nil {
		message.consume.HandleError(consumeError)
	}
	_, err = tx.HardDelete(ctx, QB{
		From:  &message,
		Where: And("id", Equal(message.ID)),
		Limit: 1,
	})              // indivisible begin
	if err != nil { // indivisible end
		return tx.RollbackWithError(err)
	}
	_, err = tx.InsertModel(ctx, &DeadLetterQueueMessage{
		QueueName:  message.QueueName,
		BusinessID: message.BusinessID,
		Reason:     reason,
	}, QB{})        // indivisible begin
	if err != nil { // indivisible end
		return tx.RollbackWithError(err)
	}
	return
}