package commands

import "github.com/dacolabs/daco/internal/opendpi"

func findConnectionName(conn *opendpi.Connection, connections map[string]opendpi.Connection) string {
	if conn == nil {
		return "unknown"
	}
	for name, c := range connections {
		if c.Type == conn.Type && c.Host == conn.Host {
			return name
		}
	}
	return "unknown"
}
