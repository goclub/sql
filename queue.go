package sq

import (
	"context"
	xerr "github.com/goclub/error"
	"strconv"
	"time"
)

type Publish struct {
	BusinessID       uint64
	NextConsumeTime  time.Duration
	MaxConsumeChance uint16
	Priority         uint8 `default:"100"`
}

func corePublishMessage(ctx context.Context, queueTimeLocation *time.Location, s interface {
	InsertModel(ctx context.Context, ptr Model, qb QB) (err error)
}, queueName string, publish Publish) (message Message, err error) {
	if queueName == "" {
		err = xerr.New("goclub/sql: PublishMessage(ctx, queueName, publish) queue can not be empty string")
		return
	}
	if publish.Priority == 0 {
		publish.Priority = 100
	}
	message = Message{
		QueueName:        queueName,
		BusinessID:       publish.BusinessID,
		Priority:         publish.Priority,
		NextConsumeTime:  time.Now().In(queueTimeLocation).Add(publish.NextConsumeTime),
		ConsumeChance:    0,
		MaxConsumeChance: publish.MaxConsumeChance,
		UpdateID:         "",
	}
	err = s.InsertModel(ctx, &message, QB{})
	if err != nil {
		return
	}
	return
}
func (db *Database) PublishMessage(ctx context.Context, queueName string, publish Publish) (message Message, err error) {
	return corePublishMessage(ctx, db.QueueTimeLocation, db, queueName, publish)
}
func (tx *T) PublishMessage(ctx context.Context, queueName string, publish Publish) (message Message, err error) {
	return corePublishMessage(ctx, tx.db.QueueTimeLocation, tx, queueName, publish)
}

type Consume struct {
	QueueName         string
	HandleError       func(err error)
	HandleMessage     func(message Message) MessageResult
	NextConsumeTime   func(consumeChance uint16, maxConsumeChance uint16) time.Duration
	queueTimeLocation *time.Location
}

func (data *Consume) initAndCheck(db *Database) (err error) {
	data.queueTimeLocation = db.QueueTimeLocation
	if data.NextConsumeTime == nil {
		data.NextConsumeTime = func(consumeChance uint16, maxConsumeChance uint16) time.Duration {
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
func (db *Database) InitQueue(ctx context.Context, queueName string) (err error) {
	createQueueTableSQL := "CREATE TABLE IF NOT EXISTS `queue_" + queueName + "` (" + `
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		business_id bigint(20) unsigned NOT NULL,
		priority tinyint(3) unsigned NOT NULL,
		update_id char(24) NOT NULL,
		consume_chance smallint(6) unsigned NOT NULL,
		max_consume_chance smallint(6) unsigned NOT NULL,
		next_consume_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY business_id (business_id),
		KEY update_id (update_id),
		KEY next_consume_time__consume_chance__priority (next_consume_time,consume_chance,max_consume_chance,priority)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	_, err = db.Exec(ctx, createQueueTableSQL, nil) // indivisible begin
	if err != nil {                                 // indivisible end
		return err
	}
	createDeadLetterTableSQL := "CREATE TABLE IF NOT EXISTS `queue_" + queueName + "_dead_letter` (" + `
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		business_id bigint(20) unsigned NOT NULL,
		reason varchar(255) NOT NULL DEFAULT '',
		handled tinyint(3) unsigned NOT NULL,
  		handled_result varchar(255) NOT NULL,
		create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY business_id (business_id),
		KEY create_time (create_time)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	_, err = db.Exec(ctx, createDeadLetterTableSQL, nil) // indivisible begin
	if err != nil {                                      // indivisible end
		return err
	}
	return
}
func (db *Database) ConsumeMessage(ctx context.Context, consume Consume) error {
	err := consume.initAndCheck(db) // indivisible begin
	if err != nil {                 // indivisible end
		return err
	}
	readInterval := time.Second

	for {
		time.Sleep(readInterval)
		consumed, err := db.tryReadQueueMessage(ctx, consume) // indivisible begin
		if err != nil {                                       // indivisible end
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
	message := Message{QueueName: consume.QueueName}
	message.consume = consume
	var queueIDs []uint64
	// 查询10个id
	err = db.QuerySliceScaner(ctx, QB{
		From:   &message,
		Select: []Column{"id"},
		Where: AndRaw(`next_consume_time <= ?`, time.Now().In(db.QueueTimeLocation)).
			AndRaw(`consume_chance < max_consume_chance`),
		OrderBy: []OrderBy{
			{"priority", DESC},
		},
		Limit: 10,
	}, ScanUint64s(&queueIDs))
	if err != nil {
		return
	}
	// 无结果则退出更新
	if len(queueIDs) == 0 {
		return
	}
	updateID := NanoID24()
	// 通过更新并发安全的标记数据 (使用where id = 进行更新,避免并发事务死锁)
	change, err := db.UpdateAffected(ctx, &message, QB{
		Index: "update_id",
		Set: Set("update_id", updateID).
			SetRaw(`consume_chance = consume_chance + ?`, 1).
			// 先将下次消费时间固定更新到10分钟后避免后续进程中断或sql执行失败导致被重复消费
			Set("next_consume_time", time.Now().In(db.QueueTimeLocation).Add(time.Minute*10)),
		Where: And("id", In(queueIDs)).AndRaw("consume_chance < max_consume_chance"),
		OrderBy: []OrderBy{
			{"priority", DESC},
		},
		Limit: 1,
	}) // indivisible begin
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
		err = xerr.New("goclub/sql: unexpected: Database{}.ConsumeMessage(): update_id(" + updateID + ") should has")
		return
	}
	consumed = true
	mqResult := consume.HandleMessage(message)
	if mqResult.err != nil {
		consume.HandleError(mqResult.err)
	}
	var execErr error
	if mqResult.ack {
		if execErr = message.execAck(db); execErr != nil {
			consume.HandleError(execErr)
		}
	} else if mqResult.requeue {
		if execErr = message.execRequeue(db, mqResult.requeueDelay); execErr != nil {
			consume.HandleError(execErr)
		}
	} else if mqResult.deadLetter {
		if execErr = message.execDeadLetter(db, mqResult.deadLetterReason); execErr != nil {
			consume.HandleError(execErr)
		}
	} else {
		consume.HandleError(xerr.New("consume.HandleMessage not allow return empty MessageRequest,messageID:" + strconv.FormatUint(message.ID, 10)))
	}
	return
}
