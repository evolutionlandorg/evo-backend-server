package pve

import (
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalcHPLimit(t *testing.T) {
	type args struct {
		hp    int
		charm int
	}
	tests := []struct {
		name string
		args args
		want decimal.Decimal
	}{
		{"+", args{hp: 99, charm: 80}, decimal.NewFromFloat32(122.76)}, // 99*(1+(80/100*0.3))
		{"-", args{hp: -1, charm: 80}, decimal.NewFromFloat32(-1.24)},  // 99*(1+(80/100*0.3)), 这里可能要加一个case?
		{"+", args{hp: 13, charm: 35}, decimal.NewFromFloat32(14.365)}, // 99*(1+(80/100*0.3))
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalcHPLimit(tt.args.hp, tt.args.charm); !got.Equal(tt.want) {
				t.Errorf("CalcHPLimit() = %v, want %v; got type = %s; want type = %s;", got, tt.want, reflect.TypeOf(got), reflect.TypeOf(tt.want))
			}
		})
	}
}

func TestCalcCRIT(t *testing.T) {
	type args struct {
		lucky int
		mood  int
	}
	tests := []struct {
		name string
		args args
		want decimal.Decimal
	}{
		// (25*1.5+10*0.2)*0.05/100
		{"+", args{lucky: 25, mood: 10}, decimal.NewFromFloat32(0.01975)},
		{"-", args{lucky: -25, mood: 10}, decimal.NewFromFloat32(-0.01775)},
		{"+", args{lucky: 99, mood: 50}, decimal.NewFromFloat32(0.07925)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalcCRIT(tt.args.lucky, tt.args.mood); !got.Equal(tt.want) {
				t.Errorf("CalcCRIT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalcATK(t *testing.T) {
	type args struct {
		strength  int
		intellect int
		finesse   int
	}
	tests := []struct {
		name string
		args args
		want decimal.Decimal
	}{
		// (strength*1()+intellect*1())*(1+(finesse/100*0.3))/10
		{"+", args{strength: 25, intellect: 35, finesse: 45}, decimal.NewFromFloat32(6.81)},
		{"+", args{strength: 26, intellect: 35, finesse: 45}, decimal.NewFromFloat32(6.9235)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalcATK(tt.args.strength, tt.args.intellect, tt.args.finesse); !got.Equal(tt.want) {
				t.Errorf("CalcATK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalcDEF(t *testing.T) {
	type args struct {
		Life  int
		agile int
	}
	tests := []struct {
		name string
		args args
		want decimal.Decimal
	}{
		// Life*(1+agile*0.3)/100
		{"+", args{Life: 1, agile: 1}, decimal.NewFromFloat32(0.01003)},
		{"+", args{Life: 100, agile: 1}, decimal.NewFromFloat32(1.003)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalcDEF(tt.args.Life, tt.args.agile); !got.Equal(tt.want) {
				t.Errorf("CalcDEF() = %v, want %v", got, tt.want)
			}
		})
	}
}
