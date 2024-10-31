package pve

import (
	"github.com/shopspring/decimal"
)

// CalcATK (strength*1(物理职业系数)+intellect*1(魔法职业系数))*(1+(finesse/100*0.3))/10
func CalcATK(strength, intellect, finesse int) decimal.Decimal {
	return decimal.NewFromInt32(int32(strength + intellect)).
		Mul(decimal.NewFromInt32(1).
			Add(decimal.NewFromInt32(int32(finesse)).Div(decimal.NewFromInt32(1000)).Mul(decimal.NewFromInt32(3)))).
		Div(decimal.NewFromInt32(10))
}

// CalcCRIT (lucky*1.5+mood*0.2)*0.05/100
func CalcCRIT(lucky, mood int) decimal.Decimal {
	return (decimal.NewFromFloat(1.5).Mul(decimal.NewFromInt32(int32(lucky))).
		Add((decimal.NewFromFloat(0.2)).Mul(decimal.NewFromInt32(int32(mood))))).
		Mul(decimal.NewFromFloat(0.0005))
}

// CalcHPLimit hp*(1+(charm/100*0.3)) hp + hp*charm/1000*3
func CalcHPLimit(hp, charm int) decimal.Decimal {
	return decimal.NewFromInt32(int32(hp)).
		Add(decimal.NewFromInt32(int32(hp)).Mul(decimal.NewFromInt32(int32(charm))).Div(decimal.NewFromInt32(1000)).Mul(decimal.NewFromInt32(3)))
}

// CalcDEF Life*(1+agile*0.3)/100
func CalcDEF(Life int, agile int) decimal.Decimal {
	hundred := decimal.NewFromInt32(100)
	return decimal.NewFromInt32(int32(Life)).
		Mul(decimal.NewFromInt32(1).Add(decimal.NewFromInt32(int32(agile)).Div(hundred).Mul(decimal.NewFromFloat(0.3)))).
		Div(hundred)
}
