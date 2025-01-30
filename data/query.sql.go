// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: query.sql

package data

import (
	"context"
	"time"
)

const addServerStatus = `-- name: AddServerStatus :one
;

INSERT INTO serverstatuses 
(game, connectPort, servername, serverid, currentPlayers , maxPlayers, online, steamid, version, map) 
VALUES 
(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
returning id, serverid, game, currentplayers, maxplayers, map, servername, password, connectport, version, steamid, online, timestamp
`

type AddServerStatusParams struct {
	Game           *string `db:"game" json:"game"`
	Connectport    *int64  `db:"connectport" json:"connectport"`
	Servername     *string `db:"servername" json:"servername"`
	Serverid       int64   `db:"serverid" json:"serverid"`
	Currentplayers *int64  `db:"currentplayers" json:"currentplayers"`
	Maxplayers     *int64  `db:"maxplayers" json:"maxplayers"`
	Online         bool    `db:"online" json:"online"`
	Steamid        *int64  `db:"steamid" json:"steamid"`
	Version        *string `db:"version" json:"version"`
	Map            *string `db:"map" json:"map"`
}

func (q *Queries) AddServerStatus(ctx context.Context, arg AddServerStatusParams) (Serverstatus, error) {
	row := q.db.QueryRowContext(ctx, addServerStatus,
		arg.Game,
		arg.Connectport,
		arg.Servername,
		arg.Serverid,
		arg.Currentplayers,
		arg.Maxplayers,
		arg.Online,
		arg.Steamid,
		arg.Version,
		arg.Map,
	)
	var i Serverstatus
	err := row.Scan(
		&i.ID,
		&i.Serverid,
		&i.Game,
		&i.Currentplayers,
		&i.Maxplayers,
		&i.Map,
		&i.Servername,
		&i.Password,
		&i.Connectport,
		&i.Version,
		&i.Steamid,
		&i.Online,
		&i.Timestamp,
	)
	return i, err
}

const addStatusPlayer = `-- name: AddStatusPlayer :exec
INSERT INTO serverstatusplayers(
  playerName, statusid
) VALUES (
   ?, ?
)
`

type AddStatusPlayerParams struct {
	Playername string `db:"playername" json:"playername"`
	Statusid   int64  `db:"statusid" json:"statusid"`
}

func (q *Queries) AddStatusPlayer(ctx context.Context, arg AddStatusPlayerParams) error {
	_, err := q.db.ExecContext(ctx, addStatusPlayer, arg.Playername, arg.Statusid)
	return err
}

const createGameServer = `-- name: CreateGameServer :one
INSERT INTO gameservers (
  name, host, port, scanIntervalSeconds
) VALUES (
  ?, ?, ?, ?
)
RETURNING id, name, host, scanintervalseconds, monitored, protocol, port, lgsmenabled, lgsmuser, lgsmpassword, lgsmcommand
`

type CreateGameServerParams struct {
	Name                string `db:"name" json:"name"`
	Host                string `db:"host" json:"host"`
	Port                int64  `db:"port" json:"port"`
	Scanintervalseconds int64  `db:"scanintervalseconds" json:"scanintervalseconds"`
}

func (q *Queries) CreateGameServer(ctx context.Context, arg CreateGameServerParams) (Gameserver, error) {
	row := q.db.QueryRowContext(ctx, createGameServer,
		arg.Name,
		arg.Host,
		arg.Port,
		arg.Scanintervalseconds,
	)
	var i Gameserver
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Host,
		&i.Scanintervalseconds,
		&i.Monitored,
		&i.Protocol,
		&i.Port,
		&i.Lgsmenabled,
		&i.Lgsmuser,
		&i.Lgsmpassword,
		&i.Lgsmcommand,
	)
	return i, err
}

const createLatestServerStatus = `-- name: CreateLatestServerStatus :exec
insert into latestserverstatus(
  server_id, status_id
) VALUES (
  ?1, ?2
  )
`

type CreateLatestServerStatusParams struct {
	ServerID *int64 `db:"server_id" json:"server_id"`
	StatusID *int64 `db:"status_id" json:"status_id"`
}

func (q *Queries) CreateLatestServerStatus(ctx context.Context, arg CreateLatestServerStatusParams) error {
	_, err := q.db.ExecContext(ctx, createLatestServerStatus, arg.ServerID, arg.StatusID)
	return err
}

const deleteAllStatusesByServer = `-- name: DeleteAllStatusesByServer :exec
delete from serverstatuses where serverid=?
`

func (q *Queries) DeleteAllStatusesByServer(ctx context.Context, serverid int64) error {
	_, err := q.db.ExecContext(ctx, deleteAllStatusesByServer, serverid)
	return err
}

const deleteGameServer = `-- name: DeleteGameServer :one
DELETE FROM gameservers
WHERE id = ?
RETURNING id, name, host, scanintervalseconds, monitored, protocol, port, lgsmenabled, lgsmuser, lgsmpassword, lgsmcommand
`

func (q *Queries) DeleteGameServer(ctx context.Context, id int64) (Gameserver, error) {
	row := q.db.QueryRowContext(ctx, deleteGameServer, id)
	var i Gameserver
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Host,
		&i.Scanintervalseconds,
		&i.Monitored,
		&i.Protocol,
		&i.Port,
		&i.Lgsmenabled,
		&i.Lgsmuser,
		&i.Lgsmpassword,
		&i.Lgsmcommand,
	)
	return i, err
}

const getAllServersWithLatestStatus = `-- name: GetAllServersWithLatestStatus :many
select 
  gameservers.id,
  gameservers.name,
  gameservers.host,
  gameservers.monitored,
  serverstatuses.game,
  serverstatuses.connectport,
  serverstatuses.online,
  serverstatuses.currentplayers,
  serverstatuses.maxplayers,
  serverstatuses.timestamp,
  MAX(timestamp)
from gameservers 
join serverstatuses on serverstatuses.serverid=gameservers.id
join (
  select serverid, MAX(timestamp) AS max_timestamp
  FROM serverstatuses
  GROUP BY serverid
) mt on serverstatuses.serverid and serverstatuses.timestamp = mt.max_timestamp
order by gameservers.id asc
`

type GetAllServersWithLatestStatusRow struct {
	ID             int64       `db:"id" json:"id"`
	Name           string      `db:"name" json:"name"`
	Host           string      `db:"host" json:"host"`
	Monitored      bool        `db:"monitored" json:"monitored"`
	Game           *string     `db:"game" json:"game"`
	Connectport    *int64      `db:"connectport" json:"connectport"`
	Online         bool        `db:"online" json:"online"`
	Currentplayers *int64      `db:"currentplayers" json:"currentplayers"`
	Maxplayers     *int64      `db:"maxplayers" json:"maxplayers"`
	Timestamp      *time.Time  `db:"timestamp" json:"timestamp"`
	Max            interface{} `db:"max" json:"max"`
}

func (q *Queries) GetAllServersWithLatestStatus(ctx context.Context) ([]GetAllServersWithLatestStatusRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllServersWithLatestStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllServersWithLatestStatusRow
	for rows.Next() {
		var i GetAllServersWithLatestStatusRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Host,
			&i.Monitored,
			&i.Game,
			&i.Connectport,
			&i.Online,
			&i.Currentplayers,
			&i.Maxplayers,
			&i.Timestamp,
			&i.Max,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCurrentStatusForServer = `-- name: GetCurrentStatusForServer :one
SELECT gameservers.id, name, host, scanintervalseconds, monitored, protocol, port, lgsmenabled, lgsmuser, lgsmpassword, lgsmcommand, serverstatuses.id, serverid, game, currentplayers, maxplayers, map, servername, password, connectport, version, steamid, online, timestamp from gameservers
join serverstatuses on serverstatuses.serverid=gameservers.id where gameservers.id=?
ORDER BY TIMESTAMP DESC
LIMIT 1
`

type GetCurrentStatusForServerRow struct {
	ID                  int64      `db:"id" json:"id"`
	Name                string     `db:"name" json:"name"`
	Host                string     `db:"host" json:"host"`
	Scanintervalseconds int64      `db:"scanintervalseconds" json:"scanintervalseconds"`
	Monitored           bool       `db:"monitored" json:"monitored"`
	Protocol            *string    `db:"protocol" json:"protocol"`
	Port                int64      `db:"port" json:"port"`
	Lgsmenabled         bool       `db:"lgsmenabled" json:"lgsmenabled"`
	Lgsmuser            *string    `db:"lgsmuser" json:"lgsmuser"`
	Lgsmpassword        *string    `db:"lgsmpassword" json:"lgsmpassword"`
	Lgsmcommand         *string    `db:"lgsmcommand" json:"lgsmcommand"`
	ID_2                int64      `db:"id_2" json:"id_2"`
	Serverid            int64      `db:"serverid" json:"serverid"`
	Game                *string    `db:"game" json:"game"`
	Currentplayers      *int64     `db:"currentplayers" json:"currentplayers"`
	Maxplayers          *int64     `db:"maxplayers" json:"maxplayers"`
	Map                 *string    `db:"map" json:"map"`
	Servername          *string    `db:"servername" json:"servername"`
	Password            *bool      `db:"password" json:"password"`
	Connectport         *int64     `db:"connectport" json:"connectport"`
	Version             *string    `db:"version" json:"version"`
	Steamid             *int64     `db:"steamid" json:"steamid"`
	Online              bool       `db:"online" json:"online"`
	Timestamp           *time.Time `db:"timestamp" json:"timestamp"`
}

func (q *Queries) GetCurrentStatusForServer(ctx context.Context, id int64) (GetCurrentStatusForServerRow, error) {
	row := q.db.QueryRowContext(ctx, getCurrentStatusForServer, id)
	var i GetCurrentStatusForServerRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Host,
		&i.Scanintervalseconds,
		&i.Monitored,
		&i.Protocol,
		&i.Port,
		&i.Lgsmenabled,
		&i.Lgsmuser,
		&i.Lgsmpassword,
		&i.Lgsmcommand,
		&i.ID_2,
		&i.Serverid,
		&i.Game,
		&i.Currentplayers,
		&i.Maxplayers,
		&i.Map,
		&i.Servername,
		&i.Password,
		&i.Connectport,
		&i.Version,
		&i.Steamid,
		&i.Online,
		&i.Timestamp,
	)
	return i, err
}

const getGameServer = `-- name: GetGameServer :one
SELECT id, name, host, scanintervalseconds, monitored, protocol, port, lgsmenabled, lgsmuser, lgsmpassword, lgsmcommand FROM gameservers
WHERE id = ?
LIMIT 1
`

func (q *Queries) GetGameServer(ctx context.Context, id int64) (Gameserver, error) {
	row := q.db.QueryRowContext(ctx, getGameServer, id)
	var i Gameserver
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Host,
		&i.Scanintervalseconds,
		&i.Monitored,
		&i.Protocol,
		&i.Port,
		&i.Lgsmenabled,
		&i.Lgsmuser,
		&i.Lgsmpassword,
		&i.Lgsmcommand,
	)
	return i, err
}

const getLastUpdateAndScanInterval = `-- name: GetLastUpdateAndScanInterval :one
select scanIntervalSeconds, timestamp from gameservers 
left join serverstatuses on serverstatuses.serverid=gameservers.id 
where gameservers.id=?
order by timestamp desc 
limit 1
`

type GetLastUpdateAndScanIntervalRow struct {
	Scanintervalseconds int64      `db:"scanintervalseconds" json:"scanintervalseconds"`
	Timestamp           *time.Time `db:"timestamp" json:"timestamp"`
}

func (q *Queries) GetLastUpdateAndScanInterval(ctx context.Context, id int64) (GetLastUpdateAndScanIntervalRow, error) {
	row := q.db.QueryRowContext(ctx, getLastUpdateAndScanInterval, id)
	var i GetLastUpdateAndScanIntervalRow
	err := row.Scan(&i.Scanintervalseconds, &i.Timestamp)
	return i, err
}

const getLatestServerStatus = `-- name: GetLatestServerStatus :one
SELECT 
  gameservers.id,
  gameservers.name,
  gameservers.host,
  gameservers.monitored,
  serverstatuses.game,
  serverstatuses.connectport,
  serverstatuses.online,
  serverstatuses.currentplayers,
  serverstatuses.maxplayers,
  serverstatuses.timestamp
 FROM serverstatuses
left join gameservers on gameservers.id=serverstatuses.serverid
WHERE serverid = ?
ORDER BY TIMESTAMP DESC 
LIMIT 1
`

type GetLatestServerStatusRow struct {
	ID             *int64     `db:"id" json:"id"`
	Name           *string    `db:"name" json:"name"`
	Host           *string    `db:"host" json:"host"`
	Monitored      *bool      `db:"monitored" json:"monitored"`
	Game           *string    `db:"game" json:"game"`
	Connectport    *int64     `db:"connectport" json:"connectport"`
	Online         bool       `db:"online" json:"online"`
	Currentplayers *int64     `db:"currentplayers" json:"currentplayers"`
	Maxplayers     *int64     `db:"maxplayers" json:"maxplayers"`
	Timestamp      *time.Time `db:"timestamp" json:"timestamp"`
}

func (q *Queries) GetLatestServerStatus(ctx context.Context, serverid int64) (GetLatestServerStatusRow, error) {
	row := q.db.QueryRowContext(ctx, getLatestServerStatus, serverid)
	var i GetLatestServerStatusRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Host,
		&i.Monitored,
		&i.Game,
		&i.Connectport,
		&i.Online,
		&i.Currentplayers,
		&i.Maxplayers,
		&i.Timestamp,
	)
	return i, err
}

const getPlayersFromStatus = `-- name: GetPlayersFromStatus :many
SELECT playerName FROM serverstatusplayers
WHERE statusid=?
`

func (q *Queries) GetPlayersFromStatus(ctx context.Context, statusid int64) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, getPlayersFromStatus, statusid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var playername string
		if err := rows.Scan(&playername); err != nil {
			return nil, err
		}
		items = append(items, playername)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGameServers = `-- name: ListGameServers :many
SELECT id, name, host, scanintervalseconds, monitored, protocol, port, lgsmenabled, lgsmuser, lgsmpassword, lgsmcommand FROM gameservers
ORDER BY name
`

func (q *Queries) ListGameServers(ctx context.Context) ([]Gameserver, error) {
	rows, err := q.db.QueryContext(ctx, listGameServers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Gameserver
	for rows.Next() {
		var i Gameserver
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Host,
			&i.Scanintervalseconds,
			&i.Monitored,
			&i.Protocol,
			&i.Port,
			&i.Lgsmenabled,
			&i.Lgsmuser,
			&i.Lgsmpassword,
			&i.Lgsmcommand,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listServerStatuses = `-- name: ListServerStatuses :many
SELECT id, serverid, game, currentplayers, maxplayers, map, servername, password, connectport, version, steamid, online, timestamp FROM serverstatuses
`

func (q *Queries) ListServerStatuses(ctx context.Context) ([]Serverstatus, error) {
	rows, err := q.db.QueryContext(ctx, listServerStatuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Serverstatus
	for rows.Next() {
		var i Serverstatus
		if err := rows.Scan(
			&i.ID,
			&i.Serverid,
			&i.Game,
			&i.Currentplayers,
			&i.Maxplayers,
			&i.Map,
			&i.Servername,
			&i.Password,
			&i.Connectport,
			&i.Version,
			&i.Steamid,
			&i.Online,
			&i.Timestamp,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setGameServerProtocol = `-- name: SetGameServerProtocol :one
UPDATE gameservers
set
  protocol = ?
WHERE
  id = ?
RETURNING id, name, host, scanintervalseconds, monitored, protocol, port, lgsmenabled, lgsmuser, lgsmpassword, lgsmcommand
`

type SetGameServerProtocolParams struct {
	Protocol *string `db:"protocol" json:"protocol"`
	ID       int64   `db:"id" json:"id"`
}

func (q *Queries) SetGameServerProtocol(ctx context.Context, arg SetGameServerProtocolParams) (Gameserver, error) {
	row := q.db.QueryRowContext(ctx, setGameServerProtocol, arg.Protocol, arg.ID)
	var i Gameserver
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Host,
		&i.Scanintervalseconds,
		&i.Monitored,
		&i.Protocol,
		&i.Port,
		&i.Lgsmenabled,
		&i.Lgsmuser,
		&i.Lgsmpassword,
		&i.Lgsmcommand,
	)
	return i, err
}

const updateGameServer = `-- name: UpdateGameServer :one
UPDATE gameservers
set
  name = ?,
  host = ?,
  port = ?,
  scanIntervalSeconds = ?,
  monitored = ?,
  lgsmenabled = ?,
  lgsmuser = ?,
  lgsmpassword = ?,
  lgsmpassword = ?,
  protocol = null
WHERE 
  id = ?
RETURNING id, name, host, scanintervalseconds, monitored, protocol, port, lgsmenabled, lgsmuser, lgsmpassword, lgsmcommand
`

type UpdateGameServerParams struct {
	Name                string  `db:"name" json:"name"`
	Host                string  `db:"host" json:"host"`
	Port                int64   `db:"port" json:"port"`
	Scanintervalseconds int64   `db:"scanintervalseconds" json:"scanintervalseconds"`
	Monitored           bool    `db:"monitored" json:"monitored"`
	Lgsmenabled         bool    `db:"lgsmenabled" json:"lgsmenabled"`
	Lgsmuser            *string `db:"lgsmuser" json:"lgsmuser"`
	Lgsmpassword        *string `db:"lgsmpassword" json:"lgsmpassword"`
	Lgsmpassword_2      *string `db:"lgsmpassword_2" json:"lgsmpassword_2"`
	ID                  int64   `db:"id" json:"id"`
}

func (q *Queries) UpdateGameServer(ctx context.Context, arg UpdateGameServerParams) (Gameserver, error) {
	row := q.db.QueryRowContext(ctx, updateGameServer,
		arg.Name,
		arg.Host,
		arg.Port,
		arg.Scanintervalseconds,
		arg.Monitored,
		arg.Lgsmenabled,
		arg.Lgsmuser,
		arg.Lgsmpassword,
		arg.Lgsmpassword_2,
		arg.ID,
	)
	var i Gameserver
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Host,
		&i.Scanintervalseconds,
		&i.Monitored,
		&i.Protocol,
		&i.Port,
		&i.Lgsmenabled,
		&i.Lgsmuser,
		&i.Lgsmpassword,
		&i.Lgsmcommand,
	)
	return i, err
}

const updateLatestServerStatus = `-- name: UpdateLatestServerStatus :exec
UPDATE latestserverstatus
  set status_id = ?1,
      timestamp = CURRENT_TIMESTAMP
  where server_id = ?2
`

type UpdateLatestServerStatusParams struct {
	StatusID *int64 `db:"status_id" json:"status_id"`
	ServerID *int64 `db:"server_id" json:"server_id"`
}

func (q *Queries) UpdateLatestServerStatus(ctx context.Context, arg UpdateLatestServerStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateLatestServerStatus, arg.StatusID, arg.ServerID)
	return err
}
