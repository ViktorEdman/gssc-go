-- name: GetGameServer :one
SELECT * FROM gameservers
WHERE id = ?
LIMIT 1;

-- name: ListGameServers :many
SELECT * FROM gameservers
ORDER BY name;


-- name: CreateGameServer :one
INSERT INTO gameservers (
  name, host, port, scanIntervalSeconds
) VALUES (
  ?, ?, ?, ?
)
RETURNING *;
-- name: UpdateGameServer :one
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
RETURNING *;

-- name: SetGameServerProtocol :one
UPDATE gameservers
set
  protocol = ?
WHERE
  id = ?
RETURNING *;

-- name: DeleteGameServer :one
DELETE FROM gameservers
WHERE id = ?
RETURNING *;

-- name: GetLatestServerStatus :one
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
LIMIT 1;
-- name: GetAllServersWithLatestStatus :many
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
;

-- name: AddServerStatus :one
INSERT INTO serverstatuses 
(game, connectPort, servername, serverid, currentPlayers , maxPlayers, online, steamid, version, map) 
VALUES 
(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
returning *;

-- name: ListServerStatuses :many
SELECT * FROM serverstatuses;

-- name: GetCurrentStatusForServer :one
SELECT * from gameservers
join serverstatuses on serverstatuses.serverid=gameservers.id where gameservers.id=?
ORDER BY TIMESTAMP DESC
LIMIT 1;

-- name: GetLastUpdateAndScanInterval :one 
select scanIntervalSeconds, timestamp from gameservers 
left join serverstatuses on serverstatuses.serverid=gameservers.id 
where gameservers.id=?
order by timestamp desc 
limit 1;

-- name: DeleteAllStatusesByServer :exec
delete from serverstatuses where serverid=?;

-- name: AddStatusPlayer :exec 
INSERT INTO serverstatusplayers(
  playerName, statusid
) VALUES (
   ?, ?
);

-- name: GetPlayersFromStatus :many
SELECT playerName FROM serverstatusplayers
WHERE statusid=?;

-- name: CreateLatestServerStatus :exec
insert into latestserverstatus(
  server_id, status_id
) VALUES (
  @server_id, @status_id
  );

-- name: UpdateLatestServerStatus :exec
UPDATE latestserverstatus
  set status_id = @status_id,
      timestamp = CURRENT_TIMESTAMP
  where server_id = @server_id;
