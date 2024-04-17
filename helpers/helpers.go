package helpers

import (
	"context"

	"github.com/ViktorEdman/gssc-go/data"
	"github.com/ViktorEdman/gssc-go/types"
)

func GetLatestStatusWithPlayer(id int64, db *data.Queries) (status *types.ServerStatusWithPlayers, err error) {
	server, err := db.GetLatestServerStatus(context.Background(), id)
	if err != nil {
		return nil, err
	}
	status = &types.ServerStatusWithPlayers{
		Gameserver:   server.Gameserver,
		Serverstatus: server.Serverstatus,
	}
	players, err := db.GetPlayersFromStatus(context.Background(), status.Serverstatus.ID)
	if err != nil {
		return status, nil
	}
	status.Players = append(status.Players, players...)
	return status, nil
}
