package sq_test

import (
	"context"
	"database/sql"
	xerr "github.com/goclub/error"
	sq "github.com/goclub/sql"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestQueueMessage(t *testing.T) {
	log.Print("skip TestQueueMessage (return)")
	return
	func() struct{} {
		// -------------
		var err error
		_ = err
		ctx := context.Background()
		_ = ctx
		err = func() (err error) {
			db := testDB
			ctx := context.Background()
			db.QueueTimeLocation = time.FixedZone("CST", 8*3600)
			queueName := "send_email"
			err = db.InitQueue(ctx, queueName) // indivisible begin
			if err != nil {                    // indivisible end
				return err
			}
			// 发布消息
			rollbackNoError, err := db.Begin(ctx, sql.LevelReadCommitted, func(tx *sq.T) sq.TxResult {
				_, err := tx.PublishMessage(ctx, queueName, sq.Publish{
					NextConsumeTime:  time.Nanosecond,
					BusinessID:       1,
					MaxConsumeChance: 3,
				})
				if err != nil {
					return tx.RollbackWithError(err)
				}
				// 插入消息详细内容(不同的业务场景详细内容不一样)
				// tx.InsertModel(ctx, &QueueSendEmailBody, sq.QB{})
				return tx.Commit()
			})
			if err != nil {
				return
			}
			if rollbackNoError {
				return xerr.New("unexpected rollback no error")
			}

			// 消费消息
			consume := sq.Consume{
				QueueName: "send_email",
				NextConsumeTime: func(consumeChance uint16, maxConsumeChance uint16) time.Duration {
					return time.Second * 3
				},
				HandleError: func(err error) {
					// 消费时产生的错误应当记录,而不是退出程序
					// 打印错误或将错误发送到 sentry
					log.Printf("%+v", err)
				},
				HandleMessage: func(message sq.Message) sq.MessageResult {
					log.Print("consume message:", message.ID)
					random := rand.Uint64() % 3 // 0 1 2
					// random = 1
					switch random {
					// 确认并删除消息
					case 0:
						log.Print("ack message:", message.ID)
						return message.Ack()
					// 退回到队列稍后再消费
					case 1:
						log.Print("requeue message:", message.ID)
						return message.Requeue(nil) // indivisible begin
					// 删除消息并记录到死信队列
					default:
						log.Print("deadLetter message:", message.ID)
						return message.DeadLetter("进入死信的原因", nil)
					}
				},
			}
			err = db.ConsumeMessage(ctx, consume)
			if err != nil {
				return
			}
			return
		}()
		// indivisible begin
		assert.NoError(t, err) // indivisible end
		// -------------
		return struct{}{}
	}()
}
