package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructToSql(t *testing.T) {
	type A struct {
		A string `json:"a" table_name:"a" symbol:"IN"`
		B int    `json:"b" table_name:"a"`
		C int    `json:"c" table_name:"c"`
		D bool   `json:"d" table_name:"d"`
	}

	type args struct {
		src interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantWheres []string
		wantValues []interface{}
	}{
		{"empty", args{src: A{}}, nil, nil},
		{"", args{src: A{A: "1", B: 1, C: 2}}, []string{"a.a IN (?)", "a.b = ?", "c.c = ?"}, []interface{}{[]string{"1"}, 1, 2}},
		{"", args{src: A{A: "1", B: 1, C: 2, D: true}}, []string{"a.a IN (?)", "a.b = ?", "c.c = ?", "d.d = ?"}, []interface{}{[]string{"1"}, 1, 2, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWheres, gotValues := StructToSql(tt.args.src)
			assert.Equalf(t, tt.wantWheres, gotWheres, "StructToSql(%v)", tt.args.src)
			assert.Equalf(t, tt.wantValues, gotValues, "StructToSql(%v)", tt.args.src)
		})
	}
}
