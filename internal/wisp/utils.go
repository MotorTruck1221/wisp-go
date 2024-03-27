package wisp

import (
	"net/http"
	"github.com/gobwas/ws"
	//"github.com/gobwas/ws/wsutil"
    "fmt"
    "net"
)

func HandleUpgrade(w http.ResponseWriter, r *http.Request) net.Conn {
    conn, _, _, err := ws.UpgradeHTTP(r, w)
    if err != nil { fmt.Println(err) }
    return conn
}
