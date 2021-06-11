package id

import (
	"context"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type DB struct {
	rdb *redis.Client
}

var ctx = context.Background()

const idKey = "voter_id"

func NewDB(redisHost string, redisPw string, dbNum int) DB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPw, // no password set
		DB:       dbNum,   // use default DB
	})
	return DB{
		rdb: rdb,
	}
}

func (db DB) idExists(key string) bool {
	res, err := db.rdb.Exists(ctx, key).Result()
	if err != nil {
		log.Warn(err)
	}
	return res > 0
}

func (db DB) ClaimId() []byte {
	if !db.idExists(idKey) {
		return make([]byte, 0)
	}
	res, err := db.rdb.RPop(ctx, idKey).Result()
	if err != nil {
		log.Warn(err)
	}
	return []byte(res)
}

func (db DB) DisownId(id []byte) {
	db.rdb.LPush(ctx, idKey, string(id))
}
