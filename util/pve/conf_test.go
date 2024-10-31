package pve

import (
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
)

func TestConf_GetAtkDefBuffByOccupational(t *testing.T) {
	InitCof(&Option{ConfDir: "../../config"})

	type args struct {
		atk          decimal.Decimal
		defBuff      decimal.Decimal
		Occupational string
	}
	tests := []struct {
		name  string
		args  args
		want  decimal.Decimal
		want1 decimal.Decimal
	}{
		{"Guard", args{atk: decimal.NewFromInt(100), defBuff: decimal.NewFromInt(100), Occupational: "Guard"}, decimal.NewFromInt(100), decimal.NewFromInt(103)},
		{"Saber", args{atk: decimal.NewFromInt(100), defBuff: decimal.NewFromInt(100), Occupational: "Saber"}, decimal.NewFromInt(103), decimal.NewFromInt(100)},

		{"Guard", args{atk: decimal.NewFromInt(0), defBuff: decimal.NewFromInt(0), Occupational: "Guard"}, decimal.NewFromInt(0), decimal.NewFromInt(0)},
		{"Saber", args{atk: decimal.NewFromInt(0), defBuff: decimal.NewFromInt(0), Occupational: "Saber"}, decimal.NewFromInt(0), decimal.NewFromInt(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetStageConf().GetAtkDefBuffByOccupational(tt.args.atk, tt.args.defBuff, tt.args.Occupational)
			if !reflect.DeepEqual(got.String(), tt.want.String()) {
				t.Errorf("GetAtkDefBuffByOccupational() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1.String(), tt.want1.String()) {
				t.Errorf("GetAtkDefBuffByOccupational() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConf_GetAtkDefBuffByEquipment(t *testing.T) {
	InitCof(&Option{ConfDir: "../../config"})

	type args struct {
		atk       decimal.Decimal
		defBuff   decimal.Decimal
		equipment string
		rarity    int
		level     int
	}
	tests := []struct {
		name  string
		args  args
		want  decimal.Decimal
		want1 decimal.Decimal
	}{
		{"Shield-1-0", args{
			atk:       decimal.NewFromInt(0),
			defBuff:   decimal.NewFromInt(0),
			equipment: "Shield",
			rarity:    1,
			level:     0,
		}, decimal.NewFromInt(0), decimal.NewFromInt(1)},

		{"Shield-1-1", args{
			atk:       decimal.NewFromInt(0),
			defBuff:   decimal.NewFromInt(0),
			equipment: "Shield",
			rarity:    1,
			level:     1,
		}, decimal.NewFromInt(0), decimal.NewFromInt(3)},

		{"Sword-1-0", args{
			atk:       decimal.NewFromInt(0),
			defBuff:   decimal.NewFromInt(0),
			equipment: "Sword",
			rarity:    1,
			level:     0,
		}, decimal.NewFromFloat(0 + 0.1), decimal.NewFromFloat32(0)},

		{"Sword-1-1", args{
			atk:       decimal.NewFromInt(0),
			defBuff:   decimal.NewFromInt(0),
			equipment: "Sword",
			rarity:    1,
			level:     1,
		}, decimal.NewFromFloat(0 + 0.3), decimal.NewFromFloat32(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetStageConf().GetAtkDefBuffByEquipment(tt.args.atk, tt.args.defBuff, tt.args.equipment, tt.args.rarity, tt.args.level)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAtkDefBuffByEquipment() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetAtkDefBuffByEquipment() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
