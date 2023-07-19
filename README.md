<img src="./images/logo_full.svg" style="max-width: 350px; margin-bottom:30px"/>

Sybline is a message broker providing a reliable way to handle message queues. It uses gRPC to communicate with other services and applications. 

The Sybline broker aims to provide features such as message persistence, routing, and delivery guarantees. Designed to be lightweight and easy to deploy. Sybline uses Raft Consensus to ensure High Availability(HA).

## Current State
It is not production ready & probably should not be used in anything important. Below is the roadmap for feature that are planned to be coming to Sybline; though subject to change. Any other features are not in the pipeline and may not be added.

Likely to have bugs and be unstable.
### How Sybline is Developed

Sybline is developed in a container-first approach, meaning almost none of the development is done outside of a docker-compose system. This ensures that features are developed with High Availability in mind.

### Developed and Tested On:
Only ever developed/tested on MacOS within a docker compose cluster.

Versions:
- Docker Compose version v2.13.0
- Docker version 20.10.21, build baeda1f

Currently no plans to run/test Sybline outside of a linux-based container.

## Proto Generation Script For Golang
```
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    mq.proto
```
## Sybline Client Languages Offically supported 
Languages and links to offical repos:
- [Go]()
