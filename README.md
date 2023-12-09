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
- [Installation](#Installation)
- [Development](#Development)
- [Contributing](#contributing)
- [License](#license)

# Introduction

The Sybline broker aims to provide features such as message persistence, routing, and delivery guarantees. Designed to be lightweight and easy to deploy. Sybline uses Raft Consensus to ensure High Availability(HA).

Sybline has not been designed to be a kafka-like replacement for extremely high speed messaging, it has been designed to simplify microservice architectures and ensure messages are not lost.

# Current State

It is not production ready & probably should not be used in anything important. Below is the roadmap for feature that are planned to be coming to Sybline; though subject to change. Any other features are not in the pipeline and may not be added.

It does not offically support any single node implementations, you must run at least 3 nodes; see docker-compose for developer template files.

Likely to have bugs and be unstable.

# Features
See docs for more information: [Sybline Docs]()

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
        "UnAssignRole": "allow",
        "CreateRole": "allow"
    }
}
```

If you have used anything like AWS IAM you should feel relatively comfortable with how Sybline handles access control via roles.

# Client Languages supported 

Languages and links to offical repos:
- [Go](https://github.com/GreedyKomodoDragon/sybline-go/tree/main)

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

If you are interested in the project but have no/little technical experience, please have a look at the [documentation repo](), it always needs changes or translations!

# License

Sybline has been released under GNU Affero General Public License v3.0. 
* This is a copyleft License