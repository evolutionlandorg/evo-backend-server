package models

import (
	"context"
	"testing"
)

func TestGetMemberBy(t *testing.T) {
	v := MemberQueryField{ItCode: "fafa"}
	v.GetMemberBy(context.TODO(), "ItCode")
}
