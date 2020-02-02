package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func subscriptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	var useWebsocket bool
	if r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket" {
		useWebsocket = true
	}

	ds := GetDatastore()

	if useWebsocket {
		// Websocket path
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("subscribe upgrade:", err)
			return
		}
		defer ws.Close()

		msgs, subChan, err := ds.RetrieveWithSubscription(topic)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		}
		for _, msg := range msgs {
			ws.WriteMessage(websocket.TextMessage, msg)
		}

		go func() {
			for {
				_, _, err := ws.ReadMessage()
				if err != nil {
					log.Println("subscribe ws read:", err)
					close(subChan.done)
					break
				}
			}
		}()

		for {
			select {
			case evt := <-subChan.msgs:
				if evt.topic == topic {
					err = ws.WriteMessage(websocket.TextMessage, evt.message)
					if err != nil {
						log.Println("subscribe ws write:", err)
						break
					}
				}
			}
		}
	} else {
		// Http path
		msgs, err := ds.RetrieveMessages(topic)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		var resp []byte
		for _, msg := range msgs {
			resp = append(resp, append(msg, byte('\n'))...)
		}
		w.Write(resp)
	}
}

func publishHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	var useWebsocket bool
	if r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket" {
		useWebsocket = true
	}

	ds := GetDatastore()

	if useWebsocket {
		// Websocket path
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("publish upgrade:", err)
			return
		}
		defer ws.Close()
		for {
			_, msgText, err := ws.ReadMessage()
			if err != nil {
				log.Println("publish ws read:", err)
				break
			}
			ds.StoreMessage(topic, msgText)
			err = ws.WriteMessage(websocket.TextMessage, []byte("ok"))
			if err != nil {
				log.Println("publish ws write:", err)
				break
			}
		}
	} else {
		// Http path
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("publish http read:", err)
			return
		}
		ds.StoreMessage(topic, body)
		w.Write([]byte("ok"))
	}
}
