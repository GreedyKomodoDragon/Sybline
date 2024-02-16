<p align="center">
<img src="./images/logo_full.svg" width="450"/>
</p>

<p align="center">
    Providing a reliable way to handle message queues
</p>

<p align="center">
    <img src="https://img.shields.io/badge/license-AGPL-blue.svg" alt="AGPL">
    <img src="https://img.shields.io/badge/Go-00ADD8?style=&logo=go&logoColor=white" alt="AGPL">
</p>

# Table of Contents
- [Introduction](#Introduction)
- [Current State](#Current-State)
- [Features](#Features)
- [Client Languages supported](#Client-Languages-supported)
- [Supported APIs](#Supported-APIs)
- [Quick Start](#Quick-Start)
- [Installation](#Installation)
- [Development](#Development)
- [Contributing](#contributing)
- [License](#license)

# Introduction

The Sybline broker aims to provide features such as message persistence, routing, and delivery guarantees. Designed to be lightweight and easy to deploy. Sybline uses Raft Consensus to ensure High Availability(HA).

Sybline has not been designed to be a kafka-like replacement for extremely high speed messaging, it has been designed to simplify microservice architectures and ensure messages are not lost.

# Current State

It is not production ready but is become increasing ready for alpha use. Cannot ensure that there will be no application breaking bugs; use with caution currently!

It does not offically support any single node implementations, you must run at least 3 nodes; see docker-compose for developer template files.

# Features
See docs for more information: [Sybline Docs](https://www.sybline.com)

The list of features are:

* Direct Message Routing   
* Batch Messaging
    * Can batch Submit messages to a routing key
    * Can batch ack and nack messages
* Built-in Consumer Management
    * Ensures only one consumer per message
    * No duplicate deliveries of messages!
* Security:
    * Account Management
    * Session Handling
    * mTLS Communication (Non-Account Level)
    * Role based Managed
        * JSON Based Roles
        * Allow and Deny Actions
            * Denys take precedence
        * Restrict Access to Specific Queues or Route Key
* High Avalibility
    * Write First Logging built-on Raft
    * Snapshotting built in
        * Incremental Snapshot to disk
        * S3 Options Planned
* Metrics
    * /metrics endpoint for prometheus (mTLS authentication only)
* Kubernetes Auto-Configuration:
    * With Statefulset+Services, the instance is able to find the other pods within the same namespace

## Role Based Access Control

Below is an example of a JSON role that could be assigned to a user. The example is for an admin only role.

```js
{
    "role": "ADMIN-ONLY", // Name of the role
    "actions": {
        "GetMessages": "deny:*",
        "SubmitMessage": "deny:*",
        "SubmitBatchedMessages": "deny:*",
        "CreateQueue": "allow",
        "ChangePassword": "allow",
        "Ack": "deny:*",
        "BatchAck": "deny:*",
        "DeleteQueue": "allow",
        "CreateUser": "allow",
        "DeleteUser": "allow",
        "CreateRole": "allow",
        "DeleteRole": "allow",
        "AssignRole": "allow",
        "UnassignRole": "allow",
        "CreateRole": "allow"
    }
}
```

If you have used anything like AWS IAM you should feel relatively comfortable with how Sybline handles access control via roles.

# Client Languages supported 

Languages and links to offical repos:
- [Go](https://github.com/GreedyKomodoDragon/sybline-go/tree/main)

# Supported APIs

Sybline has two APIs:

* gRPC 
* REST

Each has functionality that is not implemented in the other e.g. REST can grab metadata whereas gRPC cannot yet.

APIs will be functionally the same before v1 launch.

# Quick Start

If you want to just try out Sybline, here is a quick and easy docker-compose file to run a cluster.

```yaml
version: '3.1'

services:
  node_1:
    image: greedykomodo/sybline:latest
    hostname: node_1
    mem_limit: 1000m
    environment:
      - SERVER_PORT=2221
      - RAFT_NODE_ID=1
      - TLS_ENABLED=false
      - TLS_VERIFY_SKIP=false
      - NODES=2,3
      - ADDRESSES=node_2:2221,node_3:2221
      - SNAPSHOT_THRESHOLD=100000
      - HOST_IP=node_1
      - TOKEN_DURATION=1800
      - SALT=salty
    ports:
      - 2221:2221
      - 7878:7878
    networks:
      webnet:
        ipv4_address: 10.5.0.4

  node_2:
    image: greedykomodo/sybline:latest
    hostname: node_2
    mem_limit: 1000m
    environment:
      - SERVER_PORT=2221
      - RAFT_NODE_ID=2
      - TLS_ENABLED=false
      - TLS_VERIFY_SKIP=false
      - NODES=1,3
      - ADDRESSES=node_1:2221,node_3:2221
      - SNAPSHOT_THRESHOLD=100000
      - HOST_IP=node_2
      - TOKEN_DURATION=1800
      - SALT=salty
    ports:
      - 2222:2221
      - 7879:7878
    networks:
      webnet:
        ipv4_address: 10.5.0.5

  node_3:
    image: greedykomodo/sybline:latest
    hostname: node_3
    mem_limit: 1000m
    environment:
      - SERVER_PORT=2221
      - RAFT_NODE_ID=3
      - TLS_ENABLED=false
      - TLS_VERIFY_SKIP=false
      - NODES=1,2
      - ADDRESSES=node_1:2221,node_2:2221
      - SNAPSHOT_THRESHOLD=100000
      - HOST_IP=node_3
      - TOKEN_DURATION=1800
      - SALT=salty

    ports:
      - 2223:2221
      - 7880:7878
    networks:
      webnet:
        ipv4_address: 10.5.0.6

networks:
  webnet:
    driver: bridge
    ipam:
      config:
        - subnet: 10.5.0.0/16
          gateway: 10.5.0.1
```

# Installation

## Binary

### Prerequisites

* [Go - 1.21](https://go.dev/dl/)

### Clone the Repository

Clone the repository:

```
https://github.com/GreedyKomodoDragon/Sybline.git
```

### Building

Run the following to build the application:
```
go mod tidy && go build --ldflags '-w -s' -o sybline cmd/main.go
```

## Docker

There are two offical images:
* greedykomodo/sybline
* greedykomodo/sybline-ubi

We do not recommend running individual containers and connecting them up, we recommend using docker-compose to handling the networking when developing with Sybline.

### Docker Compose

For development purposes, we have a series of prepared docker-compose under `infra/docker`.

## Helm 

Not started, coming soon...

# Development

## How Sybline is currently Developed

Sybline is developed in a container-first approach, meaning almost none of the development is done outside of a docker-compose system. This ensures that features are developed with High Availability in mind.

## Developed and Tested On

Only ever developed/tested on MacOS within a docker compose cluster.

Versions:
- Docker Compose version v2.13.0
- Docker version 20.10.21, build baeda1f

Currently no plans to run/test Sybline outside of a linux-based container.

## Proto Generation Script For Golang

Script used to generate auto-generated code for grpc server and client.

Server:
```sh
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/handler/mq.proto
```

Session Handler:
```sh
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/auth/session.proto
```


# Contributing

Sybline (or any sybline related projects) are open to contributions whether these are new features or bug-fixes.

Please note, if the feature does not align with the original goal of Sybline it will sadly not be accepted; we don't want the scope of Sybline to become too unmaintainable.

If you are interested in the project but have no/little technical experience, please have a look at the [documentation repo](https://github.com/GreedyKomodoDragon/sybline-docs), it always needs changes or translations!

# License

Sybline has been released under GNU Affero General Public License v3.0. 
* This is a copyleft License