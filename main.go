package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

var addr = flag.String("addr", ":7070", "http service address")

func serveHome(rm *RoomManager, w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rooms := rm.getRooms()
	data := struct {
		Rooms []string
	}{
		Rooms: rooms,
	}

	tmpl, err := template.ParseFiles("rooms.html")
	if err != nil {
		log.Fatal("ParseFiles: ", err)
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Fatal("Execute: ", err)
	}
}

func serveRoom(rm *RoomManager, w http.ResponseWriter, r *http.Request) {
	// Getting path from url which has the form /rooms/<roomName>/ws
	roomName := mux.Vars(r)["room"]
	fmt.Println("room name: ", roomName)

	hub, err := rm.getRoom(roomName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	hub.runOnce()
	serveWs(hub, w, r)
}

func main() {
	flag.Parse()
	router := mux.NewRouter()
	roomManager := newRoomManager()

	go roomManager.run()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveHome(roomManager, w, r)
	})

	router.Handle("/rooms", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roomManager.createRoom <- r.FormValue("room")
	})).Methods(http.MethodPost)

	router.Handle("/rooms/{room}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("serving room ", r.URL.Path)
		roomName := mux.Vars(r)["room"]
		_, err := roomManager.getRoom(roomName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		tmpl, err := template.ParseFiles("room.html")
		if err != nil {
			log.Fatal("ParseFiles: ", err)
		}
		err = tmpl.Execute(w, struct {
			Room string
		}{
			Room: roomName,
		})
		if err != nil {
			log.Fatal("Execute: ", err)
		}
	})).Methods(http.MethodGet)

	router.Handle("/rooms/{room}/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("serving room ws", r.URL.Query().Get("room"))
		fmt.Println("serving room ws", r.URL.Path)
		serveRoom(roomManager, w, r)
	})).Methods(http.MethodGet)

	err := http.ListenAndServe(*addr, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
