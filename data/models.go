// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package data

import (
	"time"
)

type Gameserver struct {
	ID                  int64   `db:"id" json:"id"`
	Name                string  `db:"name" json:"name"`
	Host                string  `db:"host" json:"host"`
	Scanintervalseconds int64   `db:"scanintervalseconds" json:"scanintervalseconds"`
	Monitored           bool    `db:"monitored" json:"monitored"`
	Protocol            *string `db:"protocol" json:"protocol"`
	Port                int64   `db:"port" json:"port"`
	Lgsmenabled         bool    `db:"lgsmenabled" json:"lgsmenabled"`
	Lgsmuser            *string `db:"lgsmuser" json:"lgsmuser"`
	Lgsmpassword        *string `db:"lgsmpassword" json:"lgsmpassword"`
	Lgsmcommand         *string `db:"lgsmcommand" json:"lgsmcommand"`
}

type Serverstatus struct {
	ID             int64      `db:"id" json:"id"`
	Serverid       int64      `db:"serverid" json:"serverid"`
	Game           *string    `db:"game" json:"game"`
	Currentplayers *int64     `db:"currentplayers" json:"currentplayers"`
	Maxplayers     *int64     `db:"maxplayers" json:"maxplayers"`
	Map            *string    `db:"map" json:"map"`
	Servername     *string    `db:"servername" json:"servername"`
	Password       *bool      `db:"password" json:"password"`
	Connectport    *int64     `db:"connectport" json:"connectport"`
	Version        *string    `db:"version" json:"version"`
	Steamid        *int64     `db:"steamid" json:"steamid"`
	Online         bool       `db:"online" json:"online"`
	Timestamp      *time.Time `db:"timestamp" json:"timestamp"`
}

type Serverstatusplayer struct {
	ID         int64  `db:"id" json:"id"`
	Playername string `db:"playername" json:"playername"`
	Statusid   int64  `db:"statusid" json:"statusid"`
}
