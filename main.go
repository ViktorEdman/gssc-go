package main

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mcstatus-io/mcutil/v3"

	"github.com/ViktorEdman/gssc-go/auth"
	"github.com/ViktorEdman/gssc-go/broker"
	"github.com/ViktorEdman/gssc-go/data"
	"github.com/ViktorEdman/gssc-go/helpers"
	"github.com/ViktorEdman/gssc-go/templates"
	"github.com/ViktorEdman/gssc-go/types"
	"github.com/wisp-gg/gamequery"
	"github.com/wisp-gg/gamequery/api"
)

//go:embed schema.sql
var ddl string

func initDb() *data.Queries {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(ddl)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to db")
	queries := data.New(db)
	return queries
}

//go:embed static
var static embed.FS

var (
	db            = initDb()
	staticHandler = http.FileServer(http.FS(static))
	eventBroker   = broker.NewBroker(db)
)

func main() {
	authorizer := auth.NewAuthorizer()
	withAuth := func(h http.HandlerFunc) http.Handler {
		return authorizer.Middleware(http.HandlerFunc(h))
	}
	mux := http.NewServeMux()
	mux.Handle("/{$}", withAuth(indexHandler))
	mux.Handle("/api/{$}", withAuth(getApiServerHandler))
	mux.Handle("POST /servers", withAuth(addServerHandler))
	mux.Handle("DELETE /servers/{x}", withAuth(deleteServerHandler))
	mux.Handle("GET /servers/{$}", withAuth(getAllServersHandler))
	mux.Handle("GET /servers/{id}", withAuth(getServerHandler))
	mux.Handle("GET /servers/edit/{id}", withAuth(getEditServerHandler))
	mux.Handle("GET /servers/create", withAuth(getCreateServerHandler))
	mux.Handle("PUT /servers/{id}", withAuth(putServerHandler))
	mux.Handle("/auth/", authorizer.Mux)
	mux.Handle("/static/", staticHandler)
	mux.Handle("/events", withAuth(eventBroker.SSEHandler))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Couldn't find %v\n", r.URL.Path)
	})

	port := 8080
	go scanAllServers()
	for {
		fmt.Println("Serving on", fmt.Sprint(":", port))
		err := http.ListenAndServe(":"+fmt.Sprint(port), mux)
		if err != nil {
			fmt.Println(err)
			port++
		}
	}
}

func getCreateServerHandler(w http.ResponseWriter, r *http.Request) {
	template := templates.AddServerForm()
	template.Render(context.TODO(), w)
}

func getLatestStatusesWithPlayers() (statuses []types.ServerStatusWithPlayers, err error) {
	servers, err := db.GetAllServersWithLatestStatus(context.Background())
	if err != nil {
		return nil, err
	}
	for i := range servers {
		status := types.ServerStatusWithPlayers{
			Gameserver:   servers[i].Gameserver,
			Serverstatus: servers[i].Serverstatus,
		}
		players, err := db.GetPlayersFromStatus(context.Background(), status.Serverstatus.ID)
		if err != nil {
			statuses = append(statuses, status)
			continue
		}
		status.Players = append(status.Players, players...)
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func getApiServerHandler(w http.ResponseWriter, r *http.Request) {
	servers, err := getLatestStatusesWithPlayers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		fmt.Fprint(w, "Couldn't retrieve servers")
		return
	}
	bytes, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Write(bytes)
}

func getServerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Invalid id")
		return
	}
	template, err := getServerTemplate(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	template.Render(r.Context(), w)
}

func putServerHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAuthorized(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized\n")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Invalid id")
		return
	}
	name := r.FormValue("name")
	host := r.FormValue("host")
	monitored := r.FormValue("ismonitored") == "on"
	scaninterval, scanErr := strconv.ParseInt(r.FormValue("scaninterval"), 10, 64)
	port, err := strconv.ParseInt(r.FormValue("port"), 10, 64)
	if scanErr != nil || err != nil || host == "" || name == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Missing values")
		return
	}
	_, err = db.UpdateGameServer(context.Background(), data.UpdateGameServerParams{
		ID:                  id,
		Name:                name,
		Host:                host,
		Port:                port,
		Scanintervalseconds: scaninterval,
		Monitored:           monitored,
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Couldn't update server")
		return
	}
	template, err := getServerTemplate(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	buf := bytes.Buffer{}
	template.Render(r.Context(), &buf)

	eventBroker.BroadcastEvent(broker.SSEEvent{Event: fmt.Sprintf("server-%d", id), Data: buf.String()})
	template.Render(r.Context(), w)
}

func getEditServerHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAuthorized(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized\n")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Invalid id")
		return
	}
	server, err := db.GetGameServer(context.Background(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Server not found")
		return
	}
	template := templates.EditServerForm(server)
	w.WriteHeader(http.StatusOK)
	template.Render(context.TODO(), w)
}

func getServerTemplate(id int64) (templ.Component, error) {
	server, err := helpers.GetLatestStatusWithPlayer(id, db)
	if err != nil {
		return nil, err
	}
	return templates.ServerTemplate(*server), nil
}

func getAllServersHandler(w http.ResponseWriter, r *http.Request) {
	servers, err := getLatestStatusesWithPlayers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		fmt.Fprint(w, "Couldn't retrieve servers")
		return
	}
	templates.ServerList(servers).Render(context.TODO(), w)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	servers, err := getLatestStatusesWithPlayers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		fmt.Fprint(w, "Couldn't retrieve servers")
		return
	}
	ctx := r.Context()
	fmt.Println(ctx.Value(auth.ContextKey("Authorized")))
	templates.Index(servers).Render(r.Context(), w)
}

func deleteServerHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAuthorized(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized\n")
		return
	}
	url := strings.Split(r.URL.Path, "/")
	id, err := strconv.ParseInt(url[len(url)-1], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Invalid id\n")
		return
	}

	server, err := db.DeleteGameServer(context.Background(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Found no server with id %v\n", id)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Println("Deleted server with ID", server.ID)
	eventBroker.BroadcastEvent(broker.SSEEvent{Event: fmt.Sprintf("server-%d", server.ID), Data: "deleted"})
	fmt.Fprint(w, "")
}

func addServerHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAuthorized(r) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized\n")
		return
	}
	r.ParseForm()
	host := r.FormValue("host")
	name := r.FormValue("name")
	scanInterval, err := strconv.ParseInt(r.FormValue("scaninterval"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Invalid scan interval\n")
		return
	}
	port, err := strconv.ParseInt(r.FormValue("port"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Invalid port\n")
		return
	}
	if host == "" || name == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Missing host or name\n")
		return
	}
	r.ParseForm()
	newServer := data.CreateGameServerParams{
		Port:                port,
		Name:                name,
		Host:                host,
		Scanintervalseconds: scanInterval,
	}
	server, err := db.CreateGameServer(context.Background(), newServer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to create server.\n", err, "\n")
		return
	}
	w.WriteHeader(http.StatusCreated)
	go func() {
		scanServer(server)
		eventBroker.BroadcastEvent(broker.SSEEvent{Event: "newserver", Data: strconv.Itoa(int(server.ID))})
	}()
}

type ServerStatus struct {
	response api.Response
	protocol string
}

func scanHost(server data.Gameserver, timeout time.Duration) (response *api.Response, protocol *string, error error) {
	host := server.Host
	port := uint16(server.Port)
	if server.Protocol != nil {
		switch *server.Protocol {
		case "source":
			res, err := sourceQuery(host, port)
			if err != nil {
				return nil, nil, err
			}
			protocol := "source"
			return res, &protocol, nil

		case "minecraft":
			res, err := minecraftQuery(context.Background(), host, port)
			if err != nil {
				return nil, nil, err
			}
			protocol := "minecraft"
			return res, &protocol, nil
		case "bedrock":
			res, err := bedrockQuery(context.Background(), host, port)
			if err != nil {
				return nil, nil, err
			}
			protocol := "bedrock"
			return res, &protocol, nil

		}
	}
	resultChan := make(chan ServerStatus)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go func() {
		res, err := sourceQuery(host, port)
		if err != nil {
			return
		}
		resultChan <- ServerStatus{
			response: *res,
			protocol: "source",
		}
		close(resultChan)
	}()
	go func() {
		res, err := minecraftQuery(ctx, host, port)
		if err != nil {
			return
		}
		resultChan <- ServerStatus{response: *res, protocol: "minecraft"}
		close(resultChan)
	}()
	go func() {
		res, err := bedrockQuery(ctx, host, port)
		if err != nil {
			return
		}
		resultChan <- ServerStatus{
			response: *res,
			protocol: "bedrock",
		}
		close(resultChan)
	}()

	select {
	case result, ok := <-resultChan:
		if !ok {
			return nil, nil, errors.New("no response")
		}
		return &result.response, &result.protocol, nil
	case <-ctx.Done():
		return nil, nil, errors.New("no response")
	}
}

func sourceQuery(host string, port uint16) (*api.Response, error) {
	res, err := gamequery.Query(api.Request{
		Game: "source",
		IP:   host,
		Port: port,
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func minecraftQuery(ctx context.Context, host string, port uint16) (*api.Response, error) {
	res, err := mcutil.Status(ctx, host, port)
	if err != nil {
		return nil, err
	}
	result := api.Response{
		Players: api.PlayersResponse{
			Current: int(*res.Players.Online),
			Max:     int(*res.Players.Max),
		},
		Name: res.MOTD.Clean,
	}
	for _, player := range res.Players.Sample {
		result.Players.Names = append(result.Players.Names, player.NameClean)
	}
	return &result, nil
}

func bedrockQuery(ctx context.Context, host string, port uint16) (*api.Response, error) {
	res, err := mcutil.StatusBedrock(ctx, host, port)
	if err != nil {
		return nil, err
	}
	result := api.Response{
		Players: api.PlayersResponse{
			Current: int(*res.OnlinePlayers),
			Max:     int(*res.MaxPlayers),
		},
		Name: res.MOTD.Clean,
	}
	return &result, nil
}

func scanAllServers() {
	for {
		servers, _ := db.ListGameServers(context.Background())
		for _, server := range servers {
			go func(server data.Gameserver) {
				serverData, err := db.GetLastUpdateAndScanInterval(context.Background(), server.ID)
				if err != nil {
					log.Println(err)
					return
				}
				var timeToScan bool
				if serverData.Timestamp == nil {
					timeToScan = true
				} else {
					timeToScan = time.Since(*serverData.Timestamp) > time.Duration(serverData.Scanintervalseconds*int64(time.Second))
				}
				if timeToScan && server.Monitored {
					scanServer(server)
				}
			}(server)
		}

		time.Sleep(time.Second * 10)
	}
}

func notifyUpdate(id int64) {
	serverEvent := fmt.Sprintf("server-%d", id)
	template, err := getServerTemplate(id)
	if err != nil {
		eventBroker.Events <- broker.SSEEvent{Event: serverEvent}
		return
	}
	buf := bytes.Buffer{}
	template.Render(context.Background(), &buf)
	eventBroker.BroadcastEvent(broker.SSEEvent{Event: serverEvent, Data: buf.String()})
}

func scanServer(server data.Gameserver) {
	defer notifyUpdate(server.ID)
	log.Printf("Updating Server ID %v %v:%v\n", server.ID, server.Host, server.Port)
	res, protocol, err := scanHost(server, time.Second*3)
	if err != nil {
		fmt.Println(err)
		db.AddServerStatus(context.Background(), data.AddServerStatusParams{
			Serverid: server.ID,
			Online:   false,
		})
		return
	}
	db.SetGameServerProtocol(context.Background(), data.SetGameServerProtocolParams{
		ID:       server.ID,
		Protocol: protocol,
	})
	log.Printf("%v: %v/%v players \n", server.Name, res.Players.Current, res.Players.Max)
	maxPlayers := int64(res.Players.Max)
	currentPlayers := int64(res.Players.Current)
	serverParams := data.AddServerStatusParams{
		Serverid:       server.ID,
		Servername:     &res.Name,
		Currentplayers: &currentPlayers,
		Maxplayers:     &maxPlayers,
		Online:         true,
	}
	if server.Protocol != nil {
		if *server.Protocol == "source" {
			raw, ok := res.Raw.(api.SourceQuery_A2SInfo)
			if ok {
				steamId := int64(raw.ExtraData.SteamID)
				connectPort := int64(raw.ExtraData.Port)
				serverParams.Steamid = &steamId
				serverParams.Game = &raw.Game
				serverParams.Map = &raw.Map
				serverParams.Version = &raw.Version
				serverParams.Connectport = &connectPort

			}
		}
		if *server.Protocol == "minecraft" || *server.Protocol == "bedrock" {
			serverParams.Connectport = &server.Port
		}
		if *server.Protocol == "minecraft" {
			name := "Minecraft Java"
			serverParams.Game = &name
		}
		if *server.Protocol == "bedrock" {
			name := "Minecraft Bedrock"
			serverParams.Game = &name
		}

	}

	status, err := db.AddServerStatus(context.Background(), serverParams)
	if err != nil {
		log.Println("Error when adding status to server", server.ID, server.Name, err)
		db.AddServerStatus(context.Background(), data.AddServerStatusParams{
			Serverid: server.ID,
			Online:   false,
		})
		return
	}
	for _, name := range res.Players.Names {
		err := db.AddStatusPlayer(
			context.Background(),
			data.AddStatusPlayerParams{
				Playername: name,
				Statusid:   status.ID,
			},
		)
		if err != nil {
			log.Println(err)
		}
	}
}
