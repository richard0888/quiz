package lib

import (
	"time"

	"golang.org/x/net/websocket"
)

func setInterval(function func(), d time.Duration, async bool, clear chan bool) {
	ticker := time.NewTicker(d)
	go func() {
		for {
			select {
			case <-ticker.C:
				if async {
					go function()
				} else {
					function()
				}
			case <-clear:
				ticker.Stop()
				return
			}
		}
	}()
}
func heartbeat(ws *websocket.Conn, d time.Duration) {
	var connected = false
	var connected2 = true
	var data2 = map[string]interface{}{
		"Type": "connect",
	}
	websocket.JSON.Send(ws, data2)
	clear := make(chan bool)
	setInterval(func() {
		if !connected {
			close(clear)
			connected2 = false
			return
		}
		var data = map[string]interface{}{
			"Type": "connect",
		}
		websocket.JSON.Send(ws, data)
		connected = false
	}, d, true, clear)
	for {
		if !connected2 {
			return
		}
		var data Connect
		websocket.JSON.Receive(ws, &data)
		switch data.Type {
		case "connect":
			connected = true
		}
	}
}
