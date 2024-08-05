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


## How to set up your own PathfinderBeacon
At this moment it is not planed or advised to run your own PathfinderBeacon.  
I still need to do a lot of optimizations and security checks before being able to run it in a production environment that is not run by someone who knows the system well.  
Also a lot of features are still missing and i would like to have a more decentralized approach to the system if i ever find the time for this project to implement it.  
In the meantime pls use the public PathfinderBeacon at [pathfinderbeacon.net](https://pathfinderbeacon.net) that is also the default server for clients.


You can test the code if you change pathfinderbeacon.net to your localhost and disable the systemdresolver so the server can bind to port 53.
Optionally you can also build and run a docker container with the following command:

```bash
docker build -t pathfinderbeacon .
docker run -p 53:53/udp 53:53/tcp 80:80/tcp pathfinderbeacon
```