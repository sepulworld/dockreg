package main

import (
	"flag"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

/*
 * TODO :
 *		-> re-use client
 */

const ttl = 10
const dialTimeout = 3 * time.Second
const rwTimeout = 3 * time.Second
const loop = 3 * time.Second

var hostname, _ = os.Hostname()
var etcdCli *etcd.Client

func reg(hostPort, hostIp string) {
	_, err := etcdCli.Set("services/backend/"+hostname+"/port", hostPort, uint64(ttl))
	if err != nil {
		log.Println(err)
	}
	_, err = etcdCli.Set("services/backend/"+hostname+"/ip", hostIp, uint64(ttl))
	if err != nil {
		log.Println(err)
	}
}

func unreg() {
	etcdCli.Delete("services/backend/"+hostname, true)
}

func healthCheck(localPort string) bool {
	client := NewTimeoutClient(dialTimeout, rwTimeout)
	resp, err := client.Get("http://localhost:" + localPort)
	if err != nil {
		log.Printf("Cannot access to http://localhost:"+localPort+" : %v\n", err)
		return false
	} else {
		defer resp.Body.Close()
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Service responds with status code %v\n", resp.StatusCode)
		return false
	} else {
		return true
	}
}

func main() {

	pEtcdServer := flag.String("etcd", "", "etcd server for registration")
	pLocalPort := flag.String("port", "", "Local port to listen to")
	pDockerSock := flag.String("docker", "/var/run/docker.sock", "Unix socket to call Docker API")
	flag.Parse()

	etcdServers := strings.Split(*pEtcdServer, ",")
	hostname, _ = os.Hostname()

	etcdCli = etcd.NewClient(etcdServers)

	port, _ := getPublicPort(*pDockerSock, hostname, *pLocalPort)

	log.Printf("Starting register on %v, watching for local port %v...\n", hostname, *pLocalPort)
	log.Printf("Local port %v is mapped to public port %v on host %v", *pLocalPort, port.HostPort, port.HostIp)

	stop := make(chan os.Signal, 1)
	ticker := time.NewTicker(loop)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			<-ticker.C
			if healthCheck(*pLocalPort) {
				log.Println("Service is UP!")
				reg(port.HostPort, port.HostIp)
			} else {
				log.Println("Service is DOWN!")
				unreg()
			}
		}
	}()

	<-stop
	log.Println("Register stopped!")
	unreg()
	ticker.Stop()
}
