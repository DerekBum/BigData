package main

import (
	snapshotManager "example.com/manager"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/renameio"
)

var port = flag.String("port", ":8081",
	`Addr of the localhost server. Example: ":8081"`)

var replaceData []byte
var replaced = false
var m sync.Mutex
var manager = snapshotManager.NewManager()

func main() {
	fileName := "db.log"
	flag.Parse()

	replaced = false

	defer manager.Cancel()
	go manager.Manage()

	// Start the server and listen on port
	http.HandleFunc("/replace", handleReplace(fileName))
	http.HandleFunc("/get", handleGet(fileName))
	http.ListenAndServe(*port, nil)
}

func handleReplace(logFileName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		m.Lock()
		replaceData = body
		replaced = true

		err = manager.Put(string(body))
		if err != nil {
			log.Print(err)
		}

		m.Unlock()
		renameio.WriteFile(logFileName, body, 0644)
	}
}

func handleGet(logFileName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var data []byte

		m.Lock()
		defer m.Unlock()
		if replaced {
			data = replaceData
		} else {
			// Open a log file
			logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer logFile.Close()

			data = make([]byte, 1024)
			logFile.Seek(0, 0)
			n, err := logFile.Read(data)
			data = data[:n]
		}
		w.Write(data)
	}
}
