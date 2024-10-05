package types

import "time"

type ServerStatusWithPlayers struct {
	ID             int64      `db:"id" json:"id"`
	Name           string     `db:"name" json:"name"`
	Host           string     `db:"host" json:"host"`
	Monitored      bool       `db:"monitored" json:"monitored"`
	Connectport    *int64     `db:"connectport" json:"connectport"`
	Online         bool       `db:"online" json:"online"`
	Game           *string    `db:"game" json:"game"`
	Currentplayers *int64     `db:"currentplayers" json:"currentplayers"`
	Maxplayers     *int64     `db:"maxplayers" json:"maxplayers"`
	Timestamp      *time.Time `db:"timestamp" json:"timestamp"`
	Players        []string
}
