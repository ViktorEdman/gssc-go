package main

import (
	"fmt"

	"github.com/ViktorEdman/gssc-go/types"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	servers []types.ServerStatusWithPlayers
	cursor  int
}

func initialModel() model {
	servers, err := getLatestStatusesWithPlayers()
	if err != nil {
		servers = []types.ServerStatusWithPlayers{}
	}
	return model{
		servers: servers,
		cursor:  0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor >= 0 {
				m.cursor--
			}
		case "down":
			if m.cursor <= len(m.servers) {
				m.cursor++
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Currently monitored servers\n\n"
	for i, server := range m.servers {
		if m.cursor == i {
			s += ">  "
		}
		s += server.Name + "\n"
	}
	s += "\n\n"
	s += fmt.Sprintf("Server %s is currently ", m.servers[m.cursor].Name)
	if m.servers[m.cursor].Online {
		s += "online with "
		s += fmt.Sprintf("%d/%d players", *m.servers[m.cursor].Currentplayers, *m.servers[m.cursor].Maxplayers)
	} else {
		s += "offline"
	}
	s += "\n Press q to quit. \n"
	return s
}
