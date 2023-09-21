package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/fsm"
	"sybline/pkg/handler"
	"sybline/pkg/messages"

	"time"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverPort         string = "SERVER_PORT"
	raftNodeId         string = "RAFT_NODE_ID"
	TLS_ENABLED        string = "TLS_ENABLED"
	nodes              string = "NODES"
	addresses          string = "ADDRESSES"
	SNAPSHOT_THRESHOLD string = "SNAPSHOT_THRESHOLD"
	HOST_IP            string = "HOST_IP"

	ELECTION_TIMEOUT  string = "ELECTION_TIMEOUT"
	HEARTBEAT_TIMEOUT string = "HEARTBEAT_TIMEOUT"

	REDIS_IP       string = "REDIS_IP"
	TOKEN_DURATION string = "TOKEN_DURATION"
	REDIS_DATABASE string = "REDIS_DATABASE"
	REDIS_PASSWORD string = "REDIS_PASSWORD"

	NODE_TTL string = "NODE_TTL"
	SALT     string = "SALT"
)

var confKeys = []string{
	serverPort,
	raftNodeId,
	nodes,
	addresses,
	SNAPSHOT_THRESHOLD,
	HOST_IP,
	ELECTION_TIMEOUT,
	HEARTBEAT_TIMEOUT,
	REDIS_IP,
	TOKEN_DURATION,
	REDIS_DATABASE,
	REDIS_PASSWORD,
	TLS_ENABLED,
	SALT,
}

func main() {
	var v = viper.New()
	v.AutomaticEnv()

	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal(err)
		return
	}

	hostIP := v.GetString(HOST_IP)
	if hostIP == "" {
		log.Fatalf("HOST_IP is required")
	}

	raftId := v.GetUint64(raftNodeId)
	if raftId == 0 {
		log.Fatalf("RAFT_NODE_ID is required")
	}

	nodeTTL := v.GetInt64(NODE_TTL)
	if nodeTTL == 0 {
		nodeTTL = 5 * 60
	}

	redisIP := v.GetString(REDIS_IP)
	if redisIP == "" {
		log.Fatalf("REDIS_IP is required")
	}

	tokenDuration := v.GetInt64(TOKEN_DURATION)
	if tokenDuration == 0 {
		tokenDuration = 1800
	}

	salt := v.GetString(SALT)
	if salt == "" {
		log.Fatalf("SALT is required")
	}

	redisPassword := v.GetString(REDIS_PASSWORD)
	redisDatabase := v.GetInt(REDIS_DATABASE)

	sessHandler, err := auth.NewRedisSession(redisIP, redisDatabase, redisPassword)
	if err != nil {
		log.Fatalf("failed to start redis session: %v", err)
	}

	port := v.GetInt(serverPort)
	if port == 0 {
		port = 2221
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	messages.Initialise()

	var opts []grpc.ServerOption

	isTLSEnabled := v.GetBool(TLS_ENABLED)

	if isTLSEnabled {
		creds, err := credentials.NewServerTLSFromFile("./cert/certificate.crt", "./cert/private.key")
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	// log directory - Create a folder/directory at a full qualified path
	err = os.MkdirAll("node_data/logs", 0755)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Fatal(err)
	}

	queueMan := core.NewQueueManager(nodeTTL)
	broker := core.NewBroker(queueMan)
	consumer := core.NewConsumerManager(queueMan)

	tDur := time.Second * time.Duration(tokenDuration)
	authManger := auth.NewAuthManager(sessHandler, &auth.UuidGen{}, &auth.ByteGenerator{}, tDur)
	authManger.CreateUser("sybline", auth.GenerateHash("sybline", salt))

	electionTimeout := v.GetInt(ELECTION_TIMEOUT)
	if electionTimeout == 0 {
		electionTimeout = 2
	}

	heartbeatTimeout := v.GetInt(HEARTBEAT_TIMEOUT)
	if heartbeatTimeout == 0 {
		heartbeatTimeout = 2
	}

	snapshotThreshold := v.GetUint64(SNAPSHOT_THRESHOLD)
	if snapshotThreshold == 0 {
		snapshotThreshold = 10000
	}

	fsmStore, err := fsm.NewSyblineFSM(broker, consumer, authManger, queueMan)
	if err != nil {
		log.Fatal(err)
	}

	logStore, err := raft.NewLogStore(snapshotThreshold)
	if err != nil {
		log.Fatalf("failed to create logstore: %v", err)
	}
	grpcServer := grpc.NewServer(opts...)

	addresses := strings.Split(v.GetString(addresses), ",")

	servers := []raft.Server{}
	ids := strings.Split(v.GetString(nodes), ",")
	if len(ids) != len(addresses) {
		log.Fatal("addresses and ids must be same length")
	}

	for i, address := range addresses {
		id, err := strconv.ParseUint(ids[i], 10, 64)
		if err != nil {
			log.Fatal("invalid id passed in")
		}

		servers = append(servers, raft.Server{
			Address: address,
			Id:      id,
			Opts:    []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		})
	}

	raftServer := raft.NewRaftServer(fsmStore, logStore, grpcServer, &raft.Configuration{
		RaftConfig: &raft.RaftConfig{
			Servers: servers,
			Id:      raftId,
			ClientConf: &raft.ClientConfig{
				StreamBuildTimeout:  2 * time.Second,
				StreamBuildAttempts: 3,
				AppendTimeout:       3 * time.Second,
			},
		},
		ElectionConfig: &raft.ElectionConfig{
			ElectionTimeout:  time.Duration(electionTimeout) * time.Second,
			HeartbeatTimeout: time.Duration(heartbeatTimeout) * time.Second,
		},
	})

	raftServer.Start()

	log.Println("listening on port: ", port)
	grpcServer.RegisterService(&handler.MQEndpoints_ServiceDesc, handler.NewServer(authManger, raftServer, salt))
	grpcServer.Serve(lis)
}
