package sq_test

import (
	rand "crypto/rand"
	"log"
	"math/big"
	"runtime/debug"
)

func RangeUint64(min uint64, max uint64) (random uint64, err error) {
	if max < min {
		min, max = max, min
		log.Print("goclub/rand: RangeUint64(min, max) max should greater than or equal to min", "\n", string(debug.Stack()))
	}
	bigInt, err := rand.Int(rand.Reader, new(big.Int).SetUint64(max+1-min))
	if err != nil {
		return 0, err
	}
	random = bigInt.Uint64() + min
	return random, err
}
// func TestQueueMessage(t *testing.T) {
// 	func() struct{} {
// 	    // -------------
// 	    var err error ; _=err
// 	    ctx := context.Background() ; _ = ctx
// 	    err = func()(err error){
// 			db := testDB
// 			ctx := context.Background()
// 			db.SetQueueTimeLocation(time.FixedZone("CST", 8*3600))
// 			queueName := "send_email"
// 			err = db.InitQueue(ctx, queueName) // indivisible begin
// 			if err != nil { // indivisible end
// 				return err
// 			}
// 			// 发布消息
// 			rollbackNoError, err := db.BeginTransaction(ctx, sql.LevelReadCommitted, func(tx *sq.Transaction) sq.TxResult {
// 				_, err := tx.PublishMessage(ctx, queueName, sq.Publish{
// 					NextConsumeTime: time.Nanosecond,
// 					BusinessID: 1,
// 					ConsumeChance:   10,
// 				}) ; if err != nil {
// 					return tx.RollbackWithError(err)
// 				}
// 				// 插入消息详细内容(不同的业务场景详细内容不一样)
// 				// tx.InsertModel(ctx, &QueueSendEmailBody, sq.QB{})
// 				return tx.Commit()
// 			}) ; if err != nil {
// 				return
// 			}
// 			if rollbackNoError {
// 				return xerr.New("unexpected rollback no error")
// 			}
// 			// 消费消息
// 			consume := sq.Consume{
// 				QueueName:       "send_email",
// 				NextConsumeTime: nil,
// 				HandleError: func(err error) {
// 					// 消费时产生的错误应当记录,而不是退出程序
// 					// 打印错误或将错误发送到 sentry
// 					log.Printf("%+v", err)
// 				},
// 				HandleMessage: func(message sq.MessageQueue)(err error) {
// 					var random uint64
// 					log.Print("consume message:", message.ID)
// 					// random, err = RangeUint64(0, 2) // indivisible begin
// 					// if err != nil { // indivisible end
// 					//     return
// 					// }
// 					random = 2
// 					// 开启事务确保消费行为与(确认/退回/死信)的原子性
// 					rollbackNoError, err := db.BeginTransaction(ctx, sql.LevelReadCommitted, func(tx *sq.Transaction) sq.TxResult {
// 						switch random {
// 						// 确认并删除消息
// 						case 0:
// 							log.Print("ack message:", message.ID)
// 							err = message.Ack(ctx, tx) // indivisible begin
// 							if err != nil { // indivisible end
// 								return tx.RollbackWithError(err)
// 							}
// 						// 退回到队列稍后再消费
// 						case 1:
// 							log.Print("requeue message:", message.ID)
// 							err = message.Requeue(ctx, tx,) // indivisible begin
// 							if err != nil { // indivisible end
// 								return tx.RollbackWithError(err)
// 							}
// 						// 删除消息并记录到死信队列
// 						default:
// 							log.Print("deadLetter message:", message.ID)
// 							err = message.DeadLetter(ctx, tx,"进入死信的原因") // indivisible begin
// 							if err != nil { // indivisible end
// 								return tx.RollbackWithError(err)
// 							}
// 						}
// 						return tx.Commit()
// 					}) // indivisible begin
// 					if err != nil { // indivisible end
// 						return
// 					}
// 					if rollbackNoError {
// 						return xerr.New("unexpected rollback no error")
// 					}
// 					return
// 				},
// 			}
// 			err = db.ConsumeMessage(ctx, consume) ; if err != nil {
// 				return
// 			}
// 			return
// 		}()
// 		 // indivisible begin
// 		 assert.NoError(t, err) // indivisible end
// 	    // -------------
// 	    return struct{}{}
// 	}()
// }
