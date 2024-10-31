package daemons

import (
	"context"
	"github.com/evolutionlandorg/evo-backend/models"
	"github.com/evolutionlandorg/evo-backend/util"
	"time"
)

func StartSnapshot(ctx context.Context) {
	chains := []string{
		models.CrabChain,
		models.EthChain,
		models.PolygonChain,
		models.TronChain,
	}

	for _, chain := range chains {
		go util.ScheduledTask(ctx, models.SaveSnapshot(ctx, chain), time.Minute*2)
	}

}
