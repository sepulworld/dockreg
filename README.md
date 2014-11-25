# dockreg

## Proof of concept of Docker service registration in etcd

See [my blog post about this sample](http://adetante.github.io/articles/service-discovery-with-docker-2)

This process will request an application listen on localhost every 5 seconds. If it doesn't respond within 3 seconds, the corresponding key will be removed in etcd. Else, a key with the following path will be created on etcd server(s) : `/keys/{service}/{ip}:{mapped_port}`.


## Command line arguments

**\-\-etcd** : mandatory, the list of etcd servers for registration, with the following format: `--etcd http://host1:port1,http://host2:port2`  

**\-\-key** : optional, the name of the parent directory in etcd for keys created by the program. Default value is `service`.  

**\-\-port** : mandatory, local port of the application to register  

**\-\-localhost** : optional, localhost interface to run tcp healthcheck against.  Defaults to 'localhost' of Docker instances. Use this flag is you want to set it to 127.0.0.1 for example. format: `--localhost 127.0.0.1`  

**\-\-docker** : optional, path to the Docker UNIX socket to access the Docker API. Default value is `/var/run/docker.sock`.  

**\-\-ip** : optional, IP address of the Docker host to put in etcd. If not specified, the IP will be retrieved from Docker API on port mapping (which means you must specify it in the docker run command, as `docker run -p 8000::192.168.1.54`)  

