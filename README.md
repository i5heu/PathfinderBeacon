<p align="center">
    <image src=".media/logo.webp" width="300" />
</p>

# PathfinderBeacon
Bootstrap your distributed system with the most scalable and robust dictionary on this planet: DNS

Since DNS caches information on multiple levels, PathfinderBeacon can provide node addresses fast, highly scalable and with minimal Bandwidth required.
If for example two searches in the same Network and for the same room are done, the chances are high that the second request will never leave the network and will be answered by the local DNS cache present in the router.

Powered by cryptographic signatures, you can create your own room without any registration requirements. You just have to either provide a RSA key or let one create by PathfinderBeacon's libraries and you are ready to publish your room and nodes addresses. The room address will be created from the key, which means that anyone with the same RSA key will automatically write to the same room.

## I want to use PathfinderBeacon for my nodes
You can use the PathfinderBeacon libraries to publish your nodes addresses to the PathfinderBeacon server and to find other your nodes.  

### Libraries
At the moment there is only the official library for Golang available that can be found here:  

[Official PathfinderBeacon Golang Library](https://github.com/i5heu/PathfinderBeacon-Client-Go)  


If you want to build a library for another language, hit me up and i will help you with the development and then link it here.


## Testing
You can test the code if you change pathfinderbeacon.net to your localhost and disable the systemdresolver so the server can bind to port 53.
Optionally you can also build and run a docker container with the following command:

```bash
docker build -t pathfinderbeacon .
docker run -p 53:53/udp 53:53/tcp 80:80/tcp pathfinderbeacon
```

## API

Please note that anyone can read room and nodes and anyone with the RSA private key can write to the room and nodes. (This might change)    
The Node hash is the SHA-224 of the IP that requested the API. (This might change)

### POST /register
Needs to contain following JSON:
```json
{
    "room": "<SHA-224 of the RSA Public Key encoded in Hex>",
    "publicKey": "<RSA Public Key Pem encoded in base64>",
    "roomSignature": "<RSA signed SHA-512 of the room name (SignPKCS1v15)>",
    "addresses": [
        {
            "protocol": "<Protocol of the node (tcp or udp (case sensitive))>",
            "ip": "<IP of the node>",
            "port": "<Port of the node>"
        }
    ]
}
```

**Pls note:**   
All addresses have an TTL of 3600 seconds (1 hour) and will be removed after that time.  
You can send the addresses again to reset the TTL.  
The node will be removed after no addresses exist for it anymore.  
The room will be removed after no nodes exist for it anymore.

### DNS
#### Rooms: room.pathfinderbeacon.net  
Will return a list of nodes in the room.

```bash
$ dig -t txt 04fed05f1e90bf24aa90c31742dff154074eac3ff0457c1785c7f001.room.pathfinderbeacon.net
04fed05f1e90bf24aa90c31742dff154074eac3ff0457c1785c7f001.room.pathfinderbeacon.net. 300 IN TXT "ebe9cf214d00031849fdaaea6174cf16d9ccc94a5f237ce4ab58bf5c"
```

#### Nodes: node.pathfinderbeacon.net
Will return a list of addresses for the node.

```bash
$ dig -t txt ebe9cf214d00031849fdaaea6174cf16d9ccc94a5f237ce4ab58bf5c.node.pathfinderbeacon.net
ebe9cf214d00031849fdaaea6174cf16d9ccc94a5f237ce4ab58bf5c.node.pathfinderbeacon.net. 3018 IN TXT "tcp://128.140.37.196:80"
ebe9cf214d00031849fdaaea6174cf16d9ccc94a5f237ce4ab58bf5c.node.pathfinderbeacon.net. 3018 IN TXT "tcp://fe80::42:37ff:fe24:2116:80"
ebe9cf214d00031849fdaaea6174cf16d9ccc94a5f237ce4ab58bf5c.node.pathfinderbeacon.net. 3018 IN TXT "tcp://192.168.1.42:80"
ebe9cf214d00031849fdaaea6174cf16d9ccc94a5f237ce4ab58bf5c.node.pathfinderbeacon.net. 3018 IN TXT "tcp://100.111.10.89:80"
```

## How to set up your own PathfinderBeacon
At this moment it is not planed or advised to run your own PathfinderBeacon.  
I still need to do a lot of optimizations and security checks before being able to run it in a production environment that is not run by someone who knows the system well.  
Also a lot of features are still missing and i would like to have a more decentralized approach to the system if i ever find the time for this project to implement it.  
In the meantime pls use the public PathfinderBeacon at [pathfinderbeacon.net](https://pathfinderbeacon.net) that is also the default server for clients.

## Potential Future Features and Ideas
- [ ] Have a shared cache for the DNS server, so we can do load balancing and failover via NS records
- [ ] Have private rooms in which the addresses are encrypted with the public key of the room
- [ ] Have another way to identify nodes so a node can have a static name that is not dependent on the IP

## License
PathfinderBeacon (c) 2024 Mia Heidenstedt and contributors  
   
SPDX-License-Identifier: AGPL-3.0