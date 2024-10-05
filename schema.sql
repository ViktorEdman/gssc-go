CREATE TABLE IF NOT EXISTS gameservers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name text UNIQUE NOT NULL,
  host text NOT NULL,
  scanIntervalSeconds INTEGER NOT NULL DEFAULT 30,
  monitored boolean not null default true,
  protocol text,
  port INTEGER NOT NULL,
  lgsmenabled boolean not null default false,
  lgsmuser text,
  lgsmpassword text,
  lgsmcommand text
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_hostport 
ON gameservers (host, port);

CREATE TABLE IF NOT EXISTS serverstatuses (
  id INTEGER PRIMARY KEY,
  serverid INTEGER NOT NULL,
  game text,
  currentPlayers INTEGER,
  maxPlayers INTEGER,
  map text,
  servername text,
  password boolean,
  connectPort INTEGER,
  version text,
  steamid INTEGER,
  online BOOLEAN DEFAULT FALSE NOT NULL,
  timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(serverid) REFERENCES gameservers(id)
);

CREATE TABLE IF NOT EXISTS serverstatusplayers( 
  id INTEGER PRIMARY KEY,
  playerName TEXT NOT NULL,
  statusid INTEGER NOT NULL,
  FOREIGN KEY(statusid) REFERENCES serverstatuses(id)
);

CREATE INDEX idx_serverstatuses_serverid ON serverstatuses (serverid);
CREATE INDEX idx_serverstatusplayers_statusid ON serverstatusplayers (statusid);
CREATE INDEX idx_serverstatuses_timestamp ON serverstatuses ("timestamp");
