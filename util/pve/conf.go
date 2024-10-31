package pve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

type Monster struct {
	HP   decimal.Decimal `json:"HP"`
	ATK  decimal.Decimal `json:"ATK"`
	DEF  decimal.Decimal `json:"DEF"`
	CRIT decimal.Decimal `json:"CRIT"`
}

type Stage struct {
	Level   string `json:"level"`
	Index   int    `json:"index"`
	Monster string `json:"monster"`
	Reward  struct {
		Material map[string]decimal.Decimal `json:"material,omitempty"`
		Element  decimal.Decimal            `json:"element"`
	} `json:"reward"`
	Card [4]int `json:"card"`
}

type Conf struct {
	Monster      map[string]Monster              `json:"monster"`
	Stage        map[string]Stage                `json:"stage"`
	Level        []string                        `json:"level"`
	Material     map[string]Material             `json:"material"`
	Cards        map[string][]string             `json:"cards"`
	Occupational map[string]Occupational         `json:"occupational"`
	Equipments   map[string]map[string]Equipment `json:"equipments"`
}

type Material struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Rarity string `json:"rarity"`
}

type Equipment struct {
	Id          int                        `json:"id"`
	Name        string                     `json:"name"`
	Rarity      string                     `json:"rarity"`
	Class       string                     `json:"class"`
	Materials   map[string]decimal.Decimal `json:"materials"`
	Limit       string                     `json:"limit"`
	Element     decimal.Decimal            `json:"element"`
	SuccessRate decimal.Decimal            `json:"success_rate"`
	Buff        struct {
		ATK []decimal.Decimal `json:"atk"`
		Def []decimal.Decimal `json:"def"`
	} `json:"buff"`
	Enhance map[string]decimal.Decimal `json:"enhance"`
}

type Occupational struct {
	Id             int             `json:"id"`
	Name           string          `json:"name"`
	EquipmentLimit string          `json:"equipment_limit"`
	Fee            decimal.Decimal `json:"fee"`
	Buff           struct {
		ATK decimal.Decimal `json:"atk"`
		DEF decimal.Decimal `json:"def"`
	} `json:"buff"`
}

var (
	stageConf        = make(map[string]Conf)
	defaultStageConf Conf
)

type Option struct {
	ConfDir string
}

func InitCof(opt *Option) {

	dirPath := util.GetEnv("CONF_DIR", "config")
	if opt != nil {
		dirPath = opt.ConfDir
	}
	var readConf = func(path string, destPtr interface{}) {
		c, err := os.ReadFile(path)
		_ = json.Unmarshal(c, destPtr)
		util.Panic(err)
	}
	pveDir, _ := filepath.Abs(filepath.Join(dirPath, "pve"))
	dirs, err := os.ReadDir(pveDir)
	if err != nil {
		panic(err)
	}
	for _, v := range dirs {
		if !v.IsDir() {
			continue
		}
		if v.Name() == "default" {
			continue
		}
		var c Conf
		readConf(filepath.Join(pveDir, v.Name(), "stage.json"), &c)
		readConf(filepath.Join(pveDir, v.Name(), "monster.json"), &c.Monster)
		readConf(filepath.Join(pveDir, v.Name(), "material.json"), &c.Material)
		readConf(filepath.Join(pveDir, v.Name(), "card.json"), &c.Cards)
		readConf(filepath.Join(pveDir, v.Name(), "occupational.json"), &c.Occupational)
		readConf(filepath.Join(pveDir, v.Name(), "equipment.json"), &c.Equipments)
		stageConf[util.Title(strings.ToLower(v.Name()))] = c
	}

	// default config
	readConf(filepath.Join(pveDir, "default", "stage.json"), &defaultStageConf)
	readConf(filepath.Join(pveDir, "default", "monster.json"), &defaultStageConf.Monster)
	readConf(filepath.Join(pveDir, "default", "material.json"), &defaultStageConf.Material)
	readConf(filepath.Join(pveDir, "default", "card.json"), &defaultStageConf.Cards)
	readConf(filepath.Join(pveDir, "default", "occupational.json"), &defaultStageConf.Occupational)
	readConf(filepath.Join(pveDir, "default", "equipment.json"), &defaultStageConf.Equipments)
}

// GetStageConf 如果不传chain就返回默认数据, 否则返回chain指定数据
func GetStageConf(chain ...string) Conf {
	if len(chain) != 0 && chain[0] != "" {
		chain[0] = util.Title(strings.ToLower(chain[0]))
		if _, ok := stageConf[chain[0]]; ok {
			return stageConf[chain[0]]
		}
	}
	return defaultStageConf
}

// GetAtkDefBuffByOccupational 根据职业信息获取增加/减少的 ATK 以及 defBuff
func (c Conf) GetAtkDefBuffByOccupational(atk, defBuff decimal.Decimal, Occupational string) (decimal.Decimal, decimal.Decimal) {
	o, ok := c.Occupational[Occupational]
	if !ok {
		return atk, defBuff
	}
	if !o.Buff.DEF.IsZero() {
		defBuff = defBuff.Add(defBuff.Mul(o.Buff.DEF.Div(decimal.NewFromInt(100))))
	}
	if !o.Buff.ATK.IsZero() {
		atk = atk.Add(atk.Mul(o.Buff.ATK.Div(decimal.NewFromInt(100))))
	}
	return atk, defBuff
}

// GetAtkDefBuffByEquipment 根据装备信息获取增加/减少的 ATK 以及 defBuff
func (c Conf) GetAtkDefBuffByEquipment(atk, defBuff decimal.Decimal, equipment string, rarity, level int) (decimal.Decimal, decimal.Decimal) {
	e, ok := c.Equipments[equipment]
	if !ok {
		return atk, defBuff
	}

	if b, ok := e[cast.ToString(rarity)]; ok && len(b.Buff.ATK) > 0 {
		atk = atk.Add(b.Buff.ATK[level])
	}

	if b, ok := e[cast.ToString(rarity)]; ok && len(b.Buff.Def) > 0 {
		defBuff = defBuff.Add(b.Buff.Def[level])
	}
	return atk, defBuff
}

func MaterialIdToSymbol(id int) string {
	for k, v := range GetStageConf().Material {
		if v.Id == id {
			return k
		}
	}
	return ""
}

func Id2Occupational(id int) string {
	for _, v := range GetStageConf().Occupational {
		if v.Id == id {
			return v.Name
		}
	}
	return ""
}
