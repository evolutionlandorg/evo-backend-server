package models

import (
	"context"
	"encoding/json"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/jinzhu/gorm"
)

type Chat struct {
	gorm.Model
	MemberId    int    `gorm:"type:int" json:"member_id"`
	Content     string `gorm:"type:text" json:"content"`
	ContentType string `json:"content_type"`
}

func broadcastChannel(ctx context.Context, content, ContentType string) {
	publish := map[string]interface{}{"content": content, "member_id": 1, "name": "Eve", "msg_type": ContentType, "wallet": "0x2e0519bd273777347407bc1908211258bfd517fa"}
	c := Chat{MemberId: 1, Content: content, ContentType: ContentType}
	util.WithContextDb(ctx).Create(&c)
	publishBytes, _ := json.Marshal(publish)
	_, _ = util.SubPoolWithContextDo(ctx)("publish", "consensus-chat", string(publishBytes))
}
