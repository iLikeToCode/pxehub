# pxehub
Dynamic iPXE booting and host tracking.

* Allows registering of hosts
* Allows setting an iPXE script per host
* Identifies hosts by mac address
* Combines DHCP and TFTP together
* Web UI for managing hosts and scripts

TODO:
* Web Authentication and Users

# Config Example
```
#HTTP_BIND=:80
#HTTP_BIND=192.168.1.1:80
INTERFACE=eth0
DHCP_RANGE_START=192.168.1.10
DHCP_RANGE_END=192.168.1.254
DHCP_MASK=255.255.255.0
DHCP_ROUTER=192.168.1.1
DNS_SERVERS=1.1.1.1,1.0.0.1
```