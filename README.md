# Docker Name Resolver

Resolve docker container's name to its ip address on host machine.

In docker container, you can get the ip address of another container using its name or id.
But the name or id can not be used on the host. Because docker's embedded DNS server only
bind to its container's network namespace.

This program can resolve container's name or id to its ip address on host, so you can connect to container using its
name.

## Feature

* Resolve container's name or id to its IP address
* Support remote docker host

## Install

1. Download from release or build `main.go` manual
   ```shell
   go build main.go
   ```

2. Start `docker-name-resolver`
   ```shell
   ./docker-name-resolver
   ```

3. Depends on your DNS resolver, you have to modify its config.
   * for systemd-resolved:
      1. Create `docker-name-resolver.conf` in `/etc/systemd/resolved.conf.d`, and write the following text:
         ```text
         [Resolve]
         DNS=127.0.0.11
         Domains=~docker.
         ```
         You can change `~docker.` to any suffix you like. For example, if you want to resolve `<container-name>.container`
         to `<container-name>`'s ip, you can use `Domains=~container.`
      2. restart systemd-resolved
         ```shell
         systemctl restart systemd-resolved
         ```
   * for `dnsmasq`
      1. Append this line to your dnsmasq config file. (Typically `/etc/dnsmasq.conf`)
         ```text
         server=/docker/127.0.0.11
         ```
      2. restart dnsmasq
         ```shell
         systemctl restart dnsmasq
         ```

4. Test `<container-name>.docker`
```
$ docker run --name hello -d ubuntu /bin/sleep 100
987173b0d8f7fba5319e93cea25a2797dd14428516a5f008c527cf5bd59097f9
$ ping hello.docker
PING hello.docker (192.168.123.3) 56(84) bytes of data.
64 bytes from 192.168.123.3: icmp_seq=1 ttl=64 time=0.041 ms

```

## Usage
Arguments are passing by environment

| Variable     | Description                                                                              | Example                |
|--------------|------------------------------------------------------------------------------------------|------------------------|
| DOCKER_HOST  | The docker host to query containers. (See `-H` option docker(1) manual) (Default: local) | tcp://192.168.1.2:2376 |
| BIND_ADDRESS | Address to bind DNS server (Default: 127.0.0.11)                                         | 127.0.0.11             |


## Example

* Resolve container on remote host

You can specify remote host by environment variable `DOCKER_HOST`
```shell
export DOCKER_HOST=tcp://192.168.1.2:
./docker-name-resolver
```


## License
Apache