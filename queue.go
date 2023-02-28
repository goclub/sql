package sq

import (
	"context"
	"database/sql"
	xerr "github.com/goclub/error"
	"time"
)



type Publish struct {
	BusinessID uint64
	NextConsumeTime time.Duration
	ConsumeChance uint16
	Priority uint8
}
func (tx *Transaction) PublishMessage(ctx context.Context, queueName string, publish Publish) (message MessageQueue, err error) {
	if queueName == "" {
		err = xerr.New("goclub/sql: Transaction{}.PublishMessage(ctx, queueName, publish) queue can not be empty string")
		return
	}
	message = MessageQueue{
		QueueName:       queueName,
		BusinessID:      publish.BusinessID,
		Priority: publish.Priority,
		NextConsumeTime: time.Now().In(tx.db.QueueTimeLocation).Add(publish.NextConsumeTime),
		ConsumeChance:   publish.ConsumeChance,
		UpdateID:        sql.NullString{},
	}
	_, err = tx.InsertModel(ctx, &message, QB{}) ; if err != nil {
	    return
	}
	return
}
type Consume struct {
	QueueName    string
	HandleError func(err error)
	HandleMessage func(message MessageQueue, tx *Transaction)(err error)
	NextConsumeTime func(consumeChance uint16) time.Duration
	queueTimeLocation *time.Location
}
func (data *Consume) initAndCheck (db *Database) (err error) {
	data.queueTimeLocation = db.QueueTimeLocation
	if data.NextConsumeTime == nil {
		data.NextConsumeTime = func(consumeChance uint16) time.Duration {
			return time.Minute
		}
	}
	if data.HandleMessage == nil {
		return xerr.New("goclub/sql: Database{}.ConsumeMessage(ctx, consume) consume.HandleMessage can not be nil")
	}
	if data.HandleError == nil {
		return xerr.New("goclub/sql: Database{}.ConsumeMessage(ctx, consume) consume.HandleError can not be nil")
	}
	return
}
func (db *Database) InitQueue (ctx context.Context, queueName string) (err error) {
	createQueueTableSQL := "CREATE TABLE IF NOT EXISTS `queue_" + queueName + "` ("+ `
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		business_id bigint(20) unsigned NOT NULL,
		priority tinyint(3) unsigned NOT NULL,
		update_id char(21) DEFAULT '',
		consume_chance smallint(6) unsigned NOT NULL,
		next_consume_time datetime NOT NULL,
		create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY business_id (business_id),
		KEY update_id (update_id),
		KEY next_consume_time__consume_chance__priority (next_consume_time,consume_chance,priority)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	_, err = db.Exec(ctx, createQueueTableSQL, nil) // indivisible begin
	if err != nil { // indivisible end
		return err
	}
	createDeadLetterTableSQL := "CREATE TABLE IF NOT EXISTS `queue_" + queueName + "_dead_letter` ("+ `
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		business_id bigint(20) unsigned NOT NULL,
		reason varchar(255) NOT NULL DEFAULT '',
		create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY business_id (business_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	_, err = db.Exec(ctx, createDeadLetterTableSQL, nil) // indivisible begin
	if err != nil { // indivisible end
		return err
	}
	return
}
func (db *Database) ConsumeMessage(ctx context.Context, consume Consume) error {
	err := consume.initAndCheck(db) // indivisible begin
	if err != nil { // indivisible end
	    return err
	}
	readInterval := time.Second

	for {
		time.Sleep(readInterval)
		consumed, err := db.tryReadQueueMessage(ctx, consume) // indivisible begin
		if err != nil { // indivisible end
			consumed = false
		    consume.HandleError(err)
		}
		if consumed {
			readInterval = time.Nanosecond
		} else {
			readInterval = time.Second
		}
	}
}
func (db *Database) tryReadQueueMessage(ctx context.Context, consume Consume) (consumed bool, err error) {
	message := MessageQueue{QueueName: consume.QueueName}
	message.consume = consume
	var queueIDs []uint64
	// 查询10个id
	err = db.QuerySliceScaner(ctx, QB{
		From: &message,
		Select: []Column{"id"},
		Where: AndRaw(`next_consume_time <= ?`, time.Now().In(db.QueueTimeLocation)).
			AndRaw(`consume_chance > 0`),
			OrderBy: []OrderBy{
				{"priority", DESC},
			},
		Limit: 10,
	},ScanUint64s(&queueIDs)) ; if err != nil {
		return
	}
	// 无结果则退出更新
	if len(queueIDs) == 0 {
		return
	}
		updateID := NanoID21()
	// 通过更新并发安全的标记数据 (使用where id = 进行更新,避免并发事务死锁)
	change, err := RowsAffected(db.Update(ctx, QB{
		From: &message,
		Index: "update_id",
		Set: Set("update_id", updateID).
			SetRaw(`consume_chance = consume_chance - ?`, 1).
			// 先将下次消费时间固定更新到10分钟后避免后续进程中断或sql执行失败导致被重复消费
			Set("next_consume_time", time.Now().In(db.QueueTimeLocation).Add(time.Minute*10)),
		Where: And("id", In(queueIDs)).AndRaw("consume_chance > 0"),
		OrderBy: []OrderBy{
			{"priority", DESC},
		},
		Limit: 1,
	})) // indivisible begin
	if err != nil { // indivisible end
		return
	}
	// 无结果则退出更新
	if change == 0 {
		return
	}
	// 查询完整queue数据
	hasUpdateMessage, err := db.Query(ctx, &message, QB{
		Where: And("update_id", Equal(updateID)),
		Limit: 1,
	}) // indivisible begin
	if err != nil { // indivisible end
	    return
	}
	if hasUpdateMessage == false {
		err = xerr.New("goclub/sql: unexpected: Database{}.ConsumeMessage(): update_id("+ updateID +") should has")
		return
	}
	consumed = true
	var txErr error
	if _, txErr = db.BeginTransaction(ctx, sql.LevelReadCommitted, func(tx *Transaction) TxResult {
 		handleError := consume.HandleMessage(message, tx) // indivisible begin
 		if handleError != nil { // indivisible end
 			return tx.RollbackWithError(handleError)
 		}
		 return tx.Commit()
 	}); txErr != nil {
		consume.HandleError(txErr)
	    return
	}

	return
}
