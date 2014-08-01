package main

import (
	"net"
	"net/http"
	"net/http/httputil"
)

func SockReq(host string, req *http.Request) (resp *http.Response, err error) {
	conn, _ := net.Dial("unix", host)
	client := httputil.NewClientConn(conn, nil)
	defer client.Close()
	return client.Do(req)
}
