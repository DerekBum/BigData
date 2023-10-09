package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	snapshotManager "example.com/manager"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var port = flag.String("port", ":8081", `Addr of the localhost server. Example: ":8081"`)
var peersFile = flag.String("pf", "", "File, containing all the peers")
var source = flag.String("username", "Tulchin", "Your name, will be sent to other clients")

var manager = snapshotManager.NewManager()

//go:embed index.html
var index []byte

func main() {
	flag.Parse()

	defer manager.Cancel()
	go manager.Manage()

	peersInfo, _ := os.ReadFile(*peersFile)
	var allPeers []string
	json.Unmarshal(peersInfo, &allPeers)

	for _, peer := range allPeers {
		peer := peer
		go func() {
			for {
				ctx := context.Background()
				c, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/ws", peer), nil)
				if err != nil {
					continue
				}
				for {
					var t snapshotManager.Transaction
					err := wsjson.Read(ctx, c, &t)
					if err != nil {
						break
					}
					manager.Put(t)
				}
			}
		}()
	}

	// Start the server and listen on port
	http.HandleFunc("/replace", handleReplace())
	http.HandleFunc("/get", handleGet())
	http.HandleFunc("/test", handleTest())
	http.HandleFunc("/vclock", handleClock())
	http.HandleFunc("/ws", handleWs())
	http.ListenAndServe(*port, nil)
}

func handleReplace() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		manager.Put(manager.NewTransaction(string(body), *source))
	}
}

func handleGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(manager.Snapshot()))
	}
}

func handleTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	}
}

func handleClock() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currClock := manager.Clock()
		w.Write([]byte(fmt.Sprintf("%v\n", currClock)))
	}
}

func handleWs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
			OriginPatterns:     []string{"*"},
		})
		if err != nil {
			return
		}
		id, transactions, query := manager.NewClient()
		for _, trx := range transactions {
			wsjson.Write(r.Context(), c, trx)
		}
		for tx := range query {
			wsjson.Write(r.Context(), c, tx)
		}

		manager.DeleteClient(id)

		select {
		case <-query:
		default:
		}
	}
}
