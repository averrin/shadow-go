package main

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
)

type Client struct {
	WID     xproto.Window
	Name    string
	Desktop uint
	Active  bool
}

func GetClients() []Client {
	clients := []Client{}
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	wids, err := ewmh.ClientListGet(X)
	if err != nil {
		log.Fatal(err)
	}
	a, _ := ewmh.ActiveWindowGet(X)
	for _, wid := range wids {
		name, err := ewmh.WmNameGet(X, wid)
		if name == "Shadow" {
			SHADOW = wid
			continue
		}
		if err != nil { // not a fatal error
			log.Println(err)
			name = ""
		}
		desk, _ := ewmh.WmDesktopGet(X, wid)
		clients = append(clients, Client{
			wid, name, desk, wid == a,
		})
	}
	return clients

}
