package daemons

import (
	"context"
	"testing"

	"github.com/evolutionlandorg/evo-backend/config"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/stretchr/testify/assert"
)

func TestFreshBlockStatus(t *testing.T) {
	config.InitApplication()
	assert.NoError(t, util.InitWorkers())
	assert.NoError(t, util.InitMysql())
	assert.NoError(t, util.InitRedis())
	err := FreshTxStatus(context.TODO(), "0x4a513cc63d5852bca87282852446f4f0ede056fdcc9bb5324511dfdab3963cf4", "Eth")
	assert.NoError(t, err)
	err = FreshTxStatus(context.TODO(), "54c82479560cd7bc7129c33f19de954888a9d26aefdf657a7576680a2dd4b646", "Tron")
	assert.NoError(t, err)
}
