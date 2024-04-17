package types

import "github.com/ViktorEdman/gssc-go/data"

type ServerStatusWithPlayers struct {
	Gameserver   data.Gameserver
	Serverstatus data.Serverstatus
	Players      []string
}
