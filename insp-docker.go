package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Port struct {
	HostIp   string
	HostPort string
}

func getPublicPort(dockerSocket, hostname, localPort string) (publicPort Port, err error) {
	req, _ := http.NewRequest("GET", "/containers/"+hostname+"/json", nil)
	resp, err := SockReq(dockerSocket, req)
	if err != nil {
		log.Fatal("Cannot connect to Docker API through ", dockerSocket, " : ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Cannot get data from Docker API : ", err)
	}

	var data map[string]*json.RawMessage
	err = json.Unmarshal(body, &data)
	if err == nil {
		var networkSettings map[string]*json.RawMessage
		err = json.Unmarshal(*data["NetworkSettings"], &networkSettings)
		if err == nil {
			var ports map[string]*json.RawMessage
			err = json.Unmarshal(*networkSettings["Ports"], &ports)
			if err == nil {
				var publicPorts = make([]Port, 1, 1)
				err = json.Unmarshal(*ports[localPort+"/tcp"], &publicPorts)
				if err == nil {
					publicPort = publicPorts[0]
					return
				}
			}
		}
	}

	log.Fatal("Invalid JSON returned from Docker API : ", err)
	return
}
