package main

import (
	"os"
	"fmt"
	"log"
	"errors"

	"net/http"
	"encoding/json"

	"gopkg.in/BurntSushi/toml.v0"
	"gopkg.in/Netwurx/routeros-api-go.v0"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	valid "github.com/asaskevich/govalidator"
)

type tomlConfig struct {
	Routers map[string]router
}

type router struct {
	Hostname string
	Port int
	Username string
	Password string
}

func ReadConfig() tomlConfig {
	var configfile = "config.toml"
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config tomlConfig
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	return config
}

func send(routerName string, command string) (routeros.Reply, error) {

	var err error

	ros, err := routeros.New("demo.mt.lv:8728")
	if err != nil {
		log.Print(err)
		return routeros.Reply{}, err
	}

	err = ros.Connect("admin", "")
	if err != nil {
		log.Print(err)
		return routeros.Reply{}, err
	}

	res, err := ros.Call(command, nil)

	ros.Close()

	return res, err

}

func jsonError(text string) string {
	message := map[string]string{"error": text}
	res, _ := json.Marshal(message)
	return string(res)
}

func commandHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	var router string = c.URLParams["router"]
	var command string = c.URLParams["host"]
	var param string = c.URLParams["host"]

	var err error
	var res routeros.Reply

	switch command {
		case "ping":
			res, err = cmdPing(router, param)
		case "tracert":
			res, err = cmdTracert(router, param)
	}

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, jsonError(err.Error()))
	}
}


func cmdPing(router string, param string) (routeros.Reply, error) {

	if (!valid.IsIP(param) && !valid.IsDNSName(param)) {
		return routeros.Reply{}, errors.New("This is not a hostname or an IP address.")
	}

	res, err := send(router, fmt.Sprintf("/ping %s", param))
	return res, err
}

func cmdTracert(router string, param string) (routeros.Reply, error) {

	if (!valid.IsIP(param) && !valid.IsDNSName(param)) {
		return routeros.Reply{}, errors.New("This is not a hostname or an IP address.")
	}

	res, err := send(router, fmt.Sprintf("/tool traceroute address=%s duration=40 count=3", param))
	return res, err
}

func handleInfo(c web.C, w http.ResponseWriter, r *http.Request) {
	var router string = c.URLParams["router"]

	res, err := send(router, "/system/resource/print")

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, jsonError(err.Error()))
	}

	log.FPrint(w, res)

}

func main() {

	config := ReadConfig()

	goji.Get("/:router/:command/:host", commandHandler)
	goji.Get("/:router/info", handleInfo)

	goji.Get("/*", http.FileServer(http.Dir("web")))

	goji.Serve()

}
