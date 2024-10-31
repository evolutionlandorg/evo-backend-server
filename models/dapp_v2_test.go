package models

import (
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
)

func TestDappV2_Check(t *testing.T) {
	type fields struct {
		Bio       string
		Link      string
		LandId    uint
		ImageHash string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"true", fields{Bio: "", Link: ""}, true},
		{"true", fields{Bio: "3432", Link: ""}, true},
		{"true", fields{Bio: "3432", Link: "http://baidu.com"}, true},
		{"true", fields{Bio: "3432", Link: "h332ttp://baidu.com"}, true},
		{"false", fields{Bio: "3432", Link: "baidu.com"}, false},
		{"false", fields{Bio: randomdata.Letters(251), Link: ""}, false},
		{"true", fields{Bio: randomdata.Letters(249) + "中", Link: ""}, true},
		{"false", fields{Bio: randomdata.Letters(249) + "中中", Link: ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DappV2{
				Bio:       tt.fields.Bio,
				Link:      tt.fields.Link,
				LandId:    tt.fields.LandId,
				ImageHash: tt.fields.ImageHash,
			}
			assert.Equalf(t, tt.want, d.Check(), "Check()")
		})
	}
}
