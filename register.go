package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const ttl = 20
const dialTimeout = 3 * time.Second
const rwTimeout = 3 * time.Second
const loop = 5 * time.Second

var hostname, _ = os.Hostname()
var etcdCli *etcd.Client
var httpCli = NewTimeoutClient(dialTimeout, rwTimeout)


func reg(key, hostPort, hostIp string) {
	_, err := etcdCli.Set(fmt.Sprintf("%v/%v:%v", key, hostIp, hostPort), "running", uint64(ttl))
	if err != nil {
		log.Println(err)
	}
}

func unreg(key, hostPort, hostIp string) {
	etcdCli.Delete(fmt.Sprintf("%v/%v:%v", key, hostIp, hostPort), true)
}

func healthCheck(localPort, localHosttoCheck string) bool {
	hostCheck := fmt.Sprintf("http://%v:", localHosttoCheck)
	resp, err := httpCli.Get(hostCheck + localPort)
	if err != nil {
		log.Printf("Cannot access to %v", hostCheck +localPort+" : %v\n", err)
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
	// Parse command line arguments
	pEtcdServer := flag.String("etcd", "", "etcd server for registration")
	pEtcdKey := flag.String("key", "service", "etcd key for this service")
	pLocalPort := flag.String("port", "", "Local port to listen to")
	pLocalHosttoCheck := flag.String("localhost", "localhost", "Localhost Docker interface to run health check against")
	pDockerSock := flag.String("docker", "/var/run/docker.sock", "Unix socket to call Docker API")
	pPublicIp := flag.String("ip", "", "Public IP for this service")
	flag.Parse()

	etcdServers := strings.Split(*pEtcdServer, ",")
	hostname, _ = os.Hostname()

	// Create a client for etcd
	etcdCli = etcd.NewClient(etcdServers)

	// Get container info through Docker API
	port, _ := getPublicPort(*pDockerSock, hostname, *pLocalPort)

	// If -ip argument is passed, use this IP address for registration, otherwise use docker mapped IP.
	registerIp := port.HostIp
	if len(*pPublicIp) > 0 {
		registerIp = *pPublicIp
	}

	log.Printf("Starting register on %v, watching for local port %v...\n", hostname, *pLocalPort)
	log.Printf("Local port %v is mapped to public port %v on host %v", *pLocalPort, port.HostPort, registerIp)

	stop := make(chan os.Signal, 1)
	ticker := time.NewTicker(loop)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// At each tick, checks the health of the service and update etcd key
		for {
			<-ticker.C
			if healthCheck(*pLocalPort, *pLocalHosttoCheck) {
				log.Println("Service is UP!")
				reg(*pEtcdKey, port.HostPort, registerIp)
			} else {
				log.Println("Service is DOWN!")
				unreg(*pEtcdKey, port.HostPort, registerIp)
			}
		}
	}()

	<-stop
	log.Println("Register stopped!")
	unreg(*pEtcdKey, port.HostPort, registerIp)
	ticker.Stop()
}
