package templates

import "time"

type Server struct {
	Protocol       *string
	Currentplayers *int64
	Maxplayers     *int64
	Timestamp      *time.Time
	Steamid        *int64
	Host           string
	Name           string
	Online         bool
	Port           int64
	ID             int64
}
