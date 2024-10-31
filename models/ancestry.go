package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/evolutionlandorg/evo-backend/util"
)

func (m *Member) Parent(ctx context.Context) *Member {
	if ancestry := m.ancestorIds(); ancestry == nil {
		return nil
	} else {
		if ancestryId := util.StringToInt(ancestry[len(ancestry)-1]); ancestryId != 0 {
			return GetMember(ctx, ancestryId)
		}
		return nil
	}
}

func (m *Member) ancestorIds() []string {
	if m == nil || m.Ancestry == "" {
		return nil
	}
	ancestry := strings.Split(m.Ancestry, "/")
	if len(ancestry) == 0 {
		return nil
	}
	return ancestry
}

func (m *Member) childrenIds(ctx context.Context) []string {
	if m == nil {
		return nil
	}
	db := util.WithContextDb(ctx)
	var ids []string
	query := db.Model(&Member{}).Where("ancestry LIKE ?", "%/"+fmt.Sprintf("%d", m.ID)).Or("ancestry = ?", m.ID).Pluck("id", &ids)
	if query.Error != nil || query == nil {
		return nil
	}
	return ids
}

func (m *Member) Children(ctx context.Context) []*Member {
	if children := m.childrenIds(ctx); children == nil {
		return nil
	} else {
		db := util.WithContextDb(ctx)
		var members []*Member
		query := db.Where("id in (?)", children).Find(&members)
		if query.Error != nil || query == nil || query.RecordNotFound() {
			return nil
		}
		return members
	}
}
