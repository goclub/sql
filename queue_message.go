package sq

import (
	"context"
	"database/sql"
	xerr "github.com/goclub/error"
	"time"
)

type Message struct {
	QueueName        string
	ID               uint64    `db:"id" sq:"ignoreInsert"`
	BusinessID       uint64    `db:"business_id"`
	NextConsumeTime  time.Time `db:"next_consume_time"`
	ConsumeChance    uint16    `db:"consume_chance"`
	MaxConsumeChance uint16    `db:"max_consume_chance"`
	UpdateID         string    `db:"update_id"`
	Priority         uint8     `db:"priority"`
	CreateTime       time.Time `db:"create_time"`
	consume          Consume
	DefaultLifeCycle
	WithoutSoftDelete
}

func (message *Message) TableName() string {
	return "queue_" + message.QueueName
}
func (v *Message) AfterInsert(result Result) error {
	id, err := result.LastInsertUint64Id()
	if err != nil {
		return err
	}
	v.ID = uint64(id)
	return nil
}

type MessageResult struct {
	ack              bool
	requeue          bool
	requeueDelay     time.Duration
	deadLetter       bool
	deadLetterReason string
	err              error
}

func (v MessageResult) WithError(err error) MessageResult {
	if err != nil {
		if v.err == nil {
			v.err = err
		} else {
			v.err = xerr.WrapPrefix(err.Error(), err)
		}
	}
	return v
}
func (Message) Ack() MessageResult {
	return MessageResult{
		ack: true,
	}
}
func (Message) RequeueDelay(duration time.Duration, err error) MessageResult {
	return MessageResult{
		requeue:      true,
		requeueDelay: duration,
		err:          err,
	}
}
func (Message) Requeue(err error) MessageResult {
	return MessageResult{
		requeue: true,
		err:     err,
	}
}
func (Message) DeadLetter(reason string, err error) MessageResult {
	return MessageResult{
		deadLetter:       true,
		deadLetterReason: reason,
		err:              err,
	}
}
func (message Message) execAck(db *Database) (err error) {
	ctx := context.Background()
	if err = db.HardDelete(ctx, &message, QB{
		Where: And("id", Equal(message.ID)),
		Limit: 1,
	}); err != nil {
		return
	}
	return
}

func (message Message) execRequeue(db *Database, delay time.Duration) (err error) {
	ctx := context.Background()
	if message.ConsumeChance == message.MaxConsumeChance {
		return message.execDeadLetter(db, "MAX_CONSUME_CHANCE")
	}
	nextConsumeDuration := delay
	if nextConsumeDuration == 0 {
		nextConsumeDuration = message.consume.NextConsumeTime(message.ConsumeChance, message.MaxConsumeChance)
	}
	if err = db.Update(ctx, &message, QB{
		Where: And("id", Equal(message.ID)),
		Set:   Set("next_consume_time", time.Now().In(message.consume.queueTimeLocation).Add(nextConsumeDuration)),
		Limit: 1,
	}); err != nil {
		return
	}
	return
}
func (message Message) execDeadLetter(db *Database, reason string) (err error) {
	ctx := context.Background()
	var rollbackNoError bool
	if rollbackNoError, err = db.Begin(ctx, sql.LevelReadCommitted, func(tx *T) TxResult {
		if err = tx.HardDelete(ctx, &message, QB{
			Where: And("id", Equal(message.ID)),
			Limit: 1,
		}); err != nil { // indivisible end
			return tx.RollbackWithError(err)
		}
		if err = tx.InsertModel(ctx, &DeadLetterQueueMessage{
			QueueName:  message.QueueName,
			BusinessID: message.BusinessID,
			Reason:     reason,
		}, QB{}); err != nil { // indivisible end
			return tx.RollbackWithError(err)
		}
		return tx.Commit()
	}); err != nil {
		return
	}
	if rollbackNoError {
		return xerr.New("unexpected rollbackNoError")
	}
	return
}
