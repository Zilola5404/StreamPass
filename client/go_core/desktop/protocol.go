//go:build windows

package main

import "encoding/json"

// Request is one newline-delimited command from the Flutter UI.
type Request struct {
	ID               int    `json:"id"`
	Cmd              string `json:"cmd"`
	Token            string `json:"token"`
	RelayHost        string `json:"relayHost"`
	RelayPort        int    `json:"relayPort"`
	ConnectionConfig string `json:"connectionConfig"`
	RulesJSON        string `json:"rulesJson"`
	ExclusionsJSON   string `json:"exclusionsJson"`
	NetworkMode      string `json:"networkMode"`
	MTU              int    `json:"mtu"`
	BlockUDP443      bool   `json:"blockUdp443"`
}

// Reply is a command result (id matches Request).
type Reply struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// Event is an unsolicited log or status line.
type Event struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
	Event   string `json:"event,omitempty"`
	Relay   string `json:"relay,omitempty"`
	PingMs  int    `json:"pingMs,omitempty"`
	Error   string `json:"error,omitempty"`
}

func marshalLine(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"type":"log","message":"marshal error"}`)
	}
	return append(b, '\n')
}
