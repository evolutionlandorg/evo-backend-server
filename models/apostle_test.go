package models

import (
	"context"
	"testing"

	"github.com/evolutionlandorg/evo-backend/config"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_TransferOwner(t *testing.T) {
	config.InitApplication()
	assert.NoError(t, util.InitRedis())
	assert.NoError(t, util.InitMysql())
	apostle := GetApostleByTokenId(context.TODO(), "2a04000104000102000000000000000400000000000000000000000000000194")
	txn := util.DbBegin(context.TODO())
	_ = apostle.TransferOwner(txn, "0x4f1c93c5698cc0b2f506a449336ca44b0e111919", "onsell", "0x1B6b637E00f0Edf77113C3", 0, HecoChain)
	txn.DbCommit()

}

func TestGetApostleMiningPower(t *testing.T) {
	type args struct {
		strength  decimal.Decimal
		agile     decimal.Decimal
		potential decimal.Decimal
	}

	tests := []struct {
		name string
		args args
		want decimal.Decimal
	}{
		{"ok", args{strength: decimal.NewFromInt(41), agile: decimal.NewFromInt(11), potential: decimal.NewFromInt(60)}, decimal.NewFromFloat(0.9890350877192982)},
		{"ok", args{strength: decimal.NewFromInt(24), agile: decimal.NewFromInt(15), potential: decimal.NewFromInt(60)}, decimal.NewFromFloat(0.7894736842105263)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := GetApostleMiningPower(tt.args.strength, tt.args.agile, tt.args.potential)
			assert.Equalf(t, tt.want.String(), p.String(),
				"GetApostleMiningPower(%v, %v, %v)", tt.args.strength.String(), tt.args.agile.String(), tt.args.potential.String())
		})
	}
}
