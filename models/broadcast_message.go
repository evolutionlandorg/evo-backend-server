package models

import (
	"context"
	"time"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type BroadcastMessage struct {
	gorm.Model
	Content     string `json:"content" sql:"type:text;"`
	MsgId       string `json:"msg_id"`
	ExpiredAt   int    `json:"expired_at"` // 有效时间
	ContentType string `json:"content_type"`
}

type BroadcastMessageJson struct {
	Content     string `json:"content" sql:"type:text;"`
	MsgId       string `json:"msg_id"`
	ExpiredAt   int    `json:"expired_at"` // 有效时间
	ContentType string `json:"content_type"`
}

func (bm *BroadcastMessage) New(ctx context.Context) error {
	result := util.WithContextDb(ctx).Create(&bm)
	return result.Error
}

func RecentMsgList(ctx context.Context) *[]BroadcastMessageJson {
	db := util.WithContextDb(ctx)
	var bm []BroadcastMessageJson
	now := time.Now().Unix() / 10 * 10
	db.Table("broadcast_messages").Where("expired_at > ?", now).Order("id desc").Scan(&bm)
	return &bm
}
