# Run Command

To test our `run` function, we can execute some commands, like:

* echo hello-world
* sh (interactive)

Test isolation in interactive shell:

- hostname
- ps -aux
- id

# Setup namespaces on Command-Start

Namespaces:
https://man7.org/linux/man-pages/man7/user_namespaces.7.html

Distinct kernel namespace for:
* Unix Timesharing System (hostname)
* Process IDs
* Mounts
* Network
* IPC
* User-ID / Group-ID

Try to change the hostname in sh!

# CLONE_NEWUSER (>= Linux 3.8) 

>      Starting in Linux 3.8, unprivileged processes can create user
       namespaces, and the other types of namespaces can be created with
       just the CAP_SYS_ADMIN capability in the caller's user namespace.

       When a nonuser namespace is created, it is owned by the user
       namespace in which the creating process was a member at the time
       of the creation of the namespace.  Privileged operations on
       resources governed by the nonuser namespace require that the
       process has the necessary capabilities in the user namespace that
       owns the nonuser namespace.

       If CLONE_NEWUSER is specified along with other CLONE_NEW* flags
       in a single clone(2) or unshare(2) call, the user namespace is
       guaranteed to be created first, giving the child (clone(2)) or
       caller (unshare(2)) privileges over the remaining namespaces
       created by the call.  Thus, it is possible for an unprivileged
       caller to specify this combination of flags.

* Let's checkout the code with CLONE_NEWUSER!
* And try to change the hostname! -> peng...


# UID (User-ID) mapping

Allows to 

>    User and group ID mappings: uid_map and gid_map
       When a user namespace is created, it starts out without a mapping
       of user IDs (group IDs) to the parent user namespace.  The
       /proc/[pid]/uid_map and /proc/[pid]/gid_map files (available
       since Linux 3.5) expose the mappings for user and group IDs
       inside the user namespace for the process pid.  These files can
       be read to view the mappings in a user namespace and written to
       (once) to define the mappings.

This allows us to map id 0 (root) in container to be XXX (unprivileged) outside container

* Check the setup of the uid_map in code
* Check /proc/XXX/uid_map (`echo $$` to get current PID)

# Process-ID Namespace

By specifing the `CLONE_NEWPID` flag, you can create a new PID namespace.

* Check the code for the added flag
* `echo $$` in shell
* `ps -x` in shell -> ?

# Give it an own root

Download an Alpine "Mini Root Filesystem" for our container

https://www.alpinelinux.org/downloads/
https://dl-cdn.alpinelinux.org/alpine/v3.15/releases/x86_64/alpine-minirootfs-3.15.4-x86_64.tar.gz

```sh
pushd /tmp
mkdir alpineroot
cd alpineroot
wget https://dl-cdn.alpinelinux.org/alpine/v3.15/releases/x86_64/alpine-minirootfs-3.15.4-x86_64.tar.gz
tar -xvf alpine-minirootfs-3.15.4-x86_64.tar.gz
```

Use `chroot` (https://linux.die.net/man/1/chroot) to specify a new root `/` for the container.

* The code has to be modified a bit, because we can't pass the new root when starting a process (like NS).
* We need to invoke the `chroot` before invoking our command
* But we don't wanna mess around with the state of our main process
* So we create a process tree main -> child -> run

Test running the following:
* `ls /`
* `ps -a`

# Mount proc

We have a distinct mount ns (`CLONE_NEWNS`), but have not mounted anything
If you check `/proc/??/mounts` you'll the system mounts
If you check the Alpine Root Image, you see /proc in an empty dir (prepared for mounting)

Check the Code for the `syscall.Mount` of the `proc` filesystem!
And test `ps -a` again!

We chould do this for /tmp, /dev, ... etc.

# Use overlayfs

All changes to the filesystem are written directly to our Alpine root.
So we can't isolate container instances.

For such an isolation, we use a differenting filesystem, like overlayfs.
https://www.kernel.org/doc/html/latest/filesystems/overlayfs.html

> This document describes a prototype for a new approach to providing overlay-filesystem functionality in Linux (sometimes referred to as union-filesystems). An overlay-filesystem tries to present a filesystem which is the result over overlaying one filesystem on top of the other.

The "lowerdir" is the readonly baseline. For multiple layers, you can have multiple lowerdirs
The "upperdir" is the writable layer above. All changes to the resulting filesystem are written to the "upperdir".

Our setup:

lowerdir=Alpine root
upperdir=tempoary location, this will contain all changes we make to the fs
workdir=tempoary location, this is needed by overlay for some operations (like atomic move)

* Check the code for our overlay fs
* Try to create new files, and check upperdir
* Try to delete existing file, and check upperdir
* Restart to start from scratch
* Anaylse docker images, containers, and overlayfs

# Limit the container with cgroups

Using cgroups in rootless mode is troublesome (https://statswork.wiki/engine/security/rootless/#known-limitations). It works on Debian and Ubuntu, because of systemd support.

CGROUP documentation:

https://www.kernel.org/doc/Documentation/cgroup-v1/
https://www.kernel.org/doc/Documentation/cgroup-v1/cpusets.txt
https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt

https://en.wikipedia.org/wiki/Cgroups


## Show docker cpu set

https://docs.docker.com/engine/reference/run/#cpuset-constraint

docker run -it --rm --cpuset-cpus="1" ubuntu 

apt install stress
stress -c 8

## Use Vagrant for an easy VM setup

``` sh
# https://www.vagrantup.com/

vagrant init ubuntu/xenial64
vagrant up

# Start a terminal session
vagrant ssh

# check network interfaces
ip a
export IF=enp0s3 # variable to point to primary interface
export PS1="HOST $PS1"

```

# Create a network namespace

```sh
# we could just use `unshare` like in part 1, like:
unshare --net bash
ip a # only shows loopback, down
ip netns list # no result, anonymous netns
exit

# create a named netns in host terminal
export NETNS=ns1
ip netns add netns1
ip netns list

# exec shell in container terminal
export NETNS=ns1
ip netns exec $NETNS bash
export PS1="CONT $PS1"
ip a # only loopback

# test loopback interface
ping 127.0.0.1  # Network is unreachable

ip link set dev lo up
ping 127.0.0.1  # Network is up
```

# Connect the Network namespace with a veth pair

```sh
# test connectivity in host
ping 8.8.8.8 # ok
ip route # check routes, and test with traceroute -> gateway default route

# we need to create a veth pair to link between namespaces in host terminal
export VETHHOST=veth-$NETNS-host
export VETHCONT=veth-$NETNS-cont
ip link add $VETHHOST type veth peer name $VETHCONT
ip link # show interfaces

# set the namespace of one side of the pair to our container namespace
ip link set $VETHCONT netns $NETNS

# check address again in HOST and CONTAINER
ip link

# retry
ping 8.8.8.8


# set link up in HOST terminal
ip link set dev $VETHHOST up
ip link # check status


# set link up in CONT terminal
ip link set dev $VETHCONT up
ip link # check status
ping 8.8.8.8 # retry


# assign a IP to the pair in HOST terminal
export CIP_HOST="192.168.0.1"
ip addr add $CIP_HOST/24 dev $VETHHOST

# assign a IP to the pair in CONT terminal
export CIP_CONT="192.168.0.2"
ip addr add $CIP_CONT/24 dev $VETHCONT


# ping between the addresses
ping $CIP_HOST
ping $CIP_CONT


# ping external in CONTAINTER
ping 8.8.8.8


# add route in CONTAINER
ip route add default via $CIP_HOST


# ping external in CONTAINTER
ping 8.8.8.8
```


# Connect the container with NAT

What we need:
* Network Address Translation (NAT) https://en.wikipedia.org/wiki/Network_address_translation 
* IP Tables https://wiki.ubuntuusers.de/iptables/
* IP Masquerading https://www.linux.com/training-tutorials/what-ip-masquerading-and-when-it-use/

```sh
# enable ip_forward on HOST
echo 1 > /proc/sys/net/ipv4/ip_forward

# check the FORWARD chain
iptables -L FORWARD  # (policy ACCEPT)
iptables -P FORWARD DROP # (policy DROP per default)

# check the NAT table
iptables -t nat -L

# append a rule to the POSTROUTING chain to MASQUERADE
iptables -t nat -A POSTROUTING -s 192.168.0.0/255.255.255.0 -o $IF -j MASQUERADE

# add both directions to the FORWARD chain
iptables -A FORWARD -i $IF -o $VETHHOST -j ACCEPT
iptables -A FORWARD -o $IF -i $VETHHOST -j ACCEPT

# ping external in CONTAINTER
ping 8.8.8.8



```


# Use a Linux Bridge for Container to Container Communications

We we want to comminucate between containers, but don't want to NAT or create a vast amount of pairs.
A Linux Bridge is like a switch, where mutltiple Interfaces can be connected

```sh

# reset VM
vagrant destroy
vagrant up

# create the bridge
export BRIDGE_IP=192.168.1.1
ip link add name the-bridge type bridge
ip addr add $BRIDGE_IP/24 dev the-bridge
ip link set the-bridge up

# create container 1
export NETNS=ns1
export CIP=192.168.1.10
export VETHHOST=veth-$NETNS-host
export VETHCONT=veth-$NETNS-cont

ip netns add $NETNS
ip link add $VETHHOST type veth peer name $VETHCONT
ip link set $VETHHOST up
ip link set $VETHCONT netns $NETNS
ip netns exec $NETNS ip link set lo up
ip netns exec $NETNS ip link set $VETHCONT up
ip netns exec $NETNS ip addr add $CIP/24 dev $VETHCONT
ip link set $VETHHOST master the-bridge       # assign the bridge
ip netns exec $NETNS ip route add default via $BRIDGE_IP

# test pings from HOST
ping $CIP
ping $BRIDGE_IP

# test pings from CONTAINER
ip netns exec $NETNS ping $CIP
ip netns exec $NETNS ping $BRIDGE_IP

# create container 2
export NETNS=ns2
export VETHHOST=veth-$NETNS-host
export VETHCONT=veth-$NETNS-cont
export CIP=192.168.1.11
... same scripts like container 1



# external access (NAT, MASQUERADE) only needed for bridge
echo 1 > /proc/sys/net/ipv4/ip_forward
iptables -t nat -A POSTROUTING -s $BRIDGE_IP/24 ! -o the-bridge -j MASQUERADE
```


# Check how docker uses all this

We use different scenarios and inspect `ip a`, `bridge link`, ...

```sh
vagrant snapshot restore docker
ip a


# host networking
docker run --net=host -it alpine ip a

# bridge networking (=default)
docker run --net=bridge -it alpine ip a

# custom network (=used in docker-compose)
docker network create mynet

# use different terminals here
docker run -it --rm --name=cont1 --network=mynet alpine
docker run -it --rm --name=cont2 --network=mynet alpine
docker network delete mynet

# container attached networking (different terminals)
docker run -it --rm --name=cont1 alpine
docker run -it --rm --name=cont2 --network=container:cont1 alpine



ls -l /proc/$(docker inspect cont1 -f="{{.State.Pid}}")/ns
```


# Use port-forwarding to run a server in the container


We use a simple netcat server `nc -l PORT` and client `nc HOST PORT'

```sh
# start a server in CONTAINER termial listening on port 8080
while true; do { echo -e 'CONTAINER 1'; } | nc -l 8080; done

# test if we could listen in HOST without EADDRINUSE
nc -l 8080

# test client in HOST terminal for container-ip 192.168.0.2
nc $CIP 8080

# setup DNAT (Destination-NAT) from 6200 to 192.168.0.2:8080
export DPORT=6200
iptables -t nat -A PREROUTING -p tcp -i $IF --dport $DPORT -j DNAT --to-destination $CIP:8080
```

# Rootless Containers

https://github.com/rootless-containers/rootlesskit
https://github.com/rootless-containers/slirp4netns
https://faun.pub/podman-rootless-container-networking-1cb5a1973b4b
