package config

import (
	"github.com/evolutionlandorg/evo-backend/util"
)

func InitApplication() {
	// initLog()
	util.LoadConf()
	util.InitTron()
}
