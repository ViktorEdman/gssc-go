package helpers

import (
	"context"
	"fmt"

	"github.com/ViktorEdman/gssc-go/data"
	"github.com/ViktorEdman/gssc-go/types"
)

func GetLatestStatusWithPlayer(id int64, db *data.Queries) (status *types.ServerStatusWithPlayers, err error) {
	server, err := db.GetLatestServerStatus(context.Background(), id)
	if err != nil {
		return nil, err
	}
	status = &types.ServerStatusWithPlayers{
		ID:             *server.ID,
		Name:           *server.Name,
		Host:           *server.Host,
		Monitored:      *server.Monitored,
		Connectport:    server.Connectport,
		Game:           server.Game,
		Online:         server.Online,
		Currentplayers: server.Currentplayers,
		Maxplayers:     server.Maxplayers,
		Timestamp:      server.Timestamp,
		Players:        []string{},
	}
	fmt.Printf("%+v", status)
	players, err := db.GetPlayersFromStatus(context.Background(), status.ID)
	if err != nil {
		return status, nil
	}
	status.Players = append(status.Players, players...)
	return status, nil
}
