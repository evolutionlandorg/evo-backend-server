package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUrl(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"baidu.com", args{str: "baidu.com"}, false},
		{"http://baidu.com", args{str: "http://baidu.com"}, true},
		{"http://baidu.com:9090", args{str: "http://baidu.com:9090"}, true},
		{"https://baidu.com:9090", args{str: "https://baidu.com:9090"}, true},
		{"baidu.com:9090", args{str: "baidu.com:9090"}, false},
		{"/testing-path", args{str: "/testing-path"}, false},
		{"alskjff#?asf//dfas", args{str: "alskjff#?asf//dfas"}, false},
		{"http://127.0.0.1", args{str: "http://127.0.0.1"}, true},
		{"http://127.0.0.1:8080", args{str: "http://127.0.0.1:8080"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsUrl(tt.args.str), "IsUrl(%v)", tt.args.str)
		})
	}
}
