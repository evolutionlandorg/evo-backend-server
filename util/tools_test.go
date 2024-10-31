package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandStr(t *testing.T) {
	str := RandStr(5)
	if str == "" {
		t.Errorf("Expect can get the rand str, but not")
	}
}

func TestSnakeString(t *testing.T) {
	str := SnakeString("getCurrencyList")
	if str != "get_currency_list" {
		t.Errorf("Expect can get the right snake str, but get %s", str)
	}
}

func TestPadding(t *testing.T) {
	hexInput := "0x4a7fb5492e8f42f230922e440460e59eeafee5b9"
	paddingOutput := Padding(hexInput)
	if paddingOutput != "0000000000000000000000004a7fb5492e8f42f230922e440460e59eeafee5b9" {
		t.Errorf("Expect can get the right padding hex, but get %s", paddingOutput)
	}

}

func TestStringInSlice(t *testing.T) {
	inputSlice := []string{"hh123", "hhh456"}
	if StringInSlice("hh123", inputSlice) == false {
		t.Errorf("Expect can get string in slice true, but get false")
	}
	if StringInSlice("fawfawfa", inputSlice) {
		t.Errorf("Expect can get string in slice false, but get true")
	}
}

func TestBigToDecimal(t *testing.T) {
	type test struct {
		ArgsV        string
		ArgsDecimals int32
		Want         string
	}
	tests := []test{
		{ArgsV: "123456789123", ArgsDecimals: 18, Want: "0.0000200159983414"},
		{ArgsV: "0x3073c8bc7f5da9a59fc7f0ce828e0f000344c85c8f7b83bc80dbefb4517418ce", ArgsDecimals: 18, Want: "21915589575598824011932973275287907752303490077616123197439.9650447577299089"},
		{ArgsV: "0xa1511e5c683a007056caa1d9a8d6a37464826280", ArgsDecimals: 18, Want: "920956519276815727089579745411.3759865233454127"},

		{ArgsV: "123456789123", ArgsDecimals: 9, Want: "20015.998341411"},
		{ArgsV: "0x3073c8bc7f5da9a59fc7f0ce828e0f000344c85c8f7b83bc80dbefb4517418ce", ArgsDecimals: 9, Want: "21915589575598824011932973275287907752303490077616123197439965044757.729908942"},
		{ArgsV: "0xa1511e5c683a007056caa1d9a8d6a37464826280", ArgsDecimals: 9, Want: "920956519276815727089579745411375986523.345412736"},
	}
	for _, v := range tests {
		want := BigToDecimal(U256(v.ArgsV), v.ArgsDecimals)
		if want.String() != v.Want {
			t.Fatalf("%+v BigToDecimal error: want %v, got %v", v, v.Want, want.String())
		}
	}

}

func TestAddHex(t *testing.T) {
	input := "3073c8bc7f5da9a59fc7f0ce828e0f000344c85c8f7b83bc80dbefb4517418ce"
	if AddHex(input) != "0x3073c8bc7f5da9a59fc7f0ce828e0f000344c85c8f7b83bc80dbefb4517418ce" {
		t.Errorf("Expect can get right hex, but get %s", AddHex(input))
	}
	input = "0x6e2b9c6552e0695f163b108e6047756c142ec52a"
	if AddHex(input) != "0x6e2b9c6552e0695f163b108e6047756c142ec52a" {
		t.Errorf("Expect can get right hex, but get %s", AddHex(input))
	}
}

func TestStringsJoinQuot(t *testing.T) {
	assert.Equal(t, "'a','b','c'", StringsJoinQuot([]string{"a", "b", "c"}))
	assert.Equal(t, "", StringsJoinQuot([]string{}))
}

func TestIntsMaxIndex(t *testing.T) {
	assert.Equal(t, []int{4}, IntsMaxIndex([]int{1, 2, 3, 4, 5}))
	assert.Equal(t, []int{0, 4}, IntsMaxIndex([]int{5, 2, 3, 4, 5}))
}

func TestFormatAddressFromLog(t *testing.T) {
	assert.Equal(t, "0xa1511e5c683a007056caa1d9a8d6a37464826280", FormatAddressFromLog("0x000000000000000000000000a1511e5c683a007056caa1d9a8d6a37464826280", "Eth"))
	assert.Equal(t, "", FormatAddressFromLog("", "Eth"))
}

func TestStringsExclude(t *testing.T) {
	assert.Equal(t, []string{"a"}, StringsExclude([]string{"a", "b", "c"}, []string{"b", "c"}))
	assert.Nil(t, StringsExclude([]string{"a", "b", "c"}, []string{"a", "b", "c"}))
	assert.Equal(t, []string{"a", "b", "c"}, StringsExclude([]string{"a", "b", "c"}, []string{}))
}
