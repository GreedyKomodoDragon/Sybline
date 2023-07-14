# Sybline

This messaging broker, called Sybline, is written in Go and provides reliable way to handle message queues. It uses gRPC to communicate with other services and applications. 

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


## Feature Road Map

Below is a running list of features implemented along with release targets. Avoids having to set up task tracking for this project since it is currently a side project.
### Features left on Road Map for v0.1.0
There is no timeline for this road map.
- [ ] TLS confirmed working
    - [ ] Redis Session
- [ ] Add Validation to each endpoint

### Feature Road Map v0.2.0
- [ ] Enable incremental Snapshots to disk
- [ ] Redis Cluster Support

### Known Bugs
- Snapshot randomly cannot find log

### Unconfirmed Features/Idea list
- [ ] User can whitelist a queue's access
- [ ] Change users to have access levels e.g. seperate infra, publisher & consumer
- [ ] Can read from yaml config to start server (leader uses this)
    - May require some work
- [ ] Some way to peform log compression, remove logs that no longer needed
    - May not be an issue with log compression after incremental snapshots are implemented
- [ ] Manual node managment via clients 
    - [ ] Can add nodes
    - [ ] Can delete nodes

## Testing Required
Not currently close to any release, but this section will be filled with more in-depth testing then.
- [ ] Nodes can fail and comeback online in more realistic environment

## Tech Debt
- [ ] Change Auth to take token not MetaData value
- [ ] Use object pools for queue nodes
- [ ] Improve memory management/configurations
    - Many structs/objects are thrown away but could be re-used e.g. in object pools



