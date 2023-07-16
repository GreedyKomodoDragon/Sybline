package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/fsm"
	"sybline/pkg/handler"
	"sybline/pkg/messages"

	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	serverPort               string = "SERVER_PORT"
	raftNodeId               string = "RAFT_NODE_ID"
	raftPort                 string = "RAFT_PORT"
	raftVolDir               string = "RAFT_VOL_DIR"
	TLS_ENABLED              string = "TLS_ENABLED"
	nodes                    string = "NODES"
	addresses                string = "ADDRESSES"
	BATCH_LIMIT              string = "BATCH_LIMIT"
	SNAPSHOT_THRESHOLD       string = "SNAPSHOT_THRESHOLD"
	SNAPSHOT_RETENTION_COUNT string = "SNAPSHOT_RETENTION_COUNT"
	CACHE_LIMIT              string = "CACHE_LIMIT"
	HOST_IP                  string = "HOST_IP"

	ELECTION_TIMEOUT     string = "ELECTION_TIMEOUT"
	HEARTBEAT_TIMEOUT    string = "HEARTBEAT_TIMEOUT"
	LEADER_LEASE_TIMEOUT string = "HEARTBEAT_TIMEOUT"

	REDIS_IP       string = "REDIS_IP"
	TOKEN_DURATION string = "TOKEN_DURATION"
	REDIS_DATABASE string = "REDIS_DATABASE"
	REDIS_PASSWORD string = "REDIS_PASSWORD"

	NODE_TTL string = "NODE_TTL"

	SALT string = "SALT"
)

var confKeys = []string{
	serverPort,
	raftNodeId,
	raftPort,
	raftVolDir,
	nodes,
	addresses,
	BATCH_LIMIT,
	SNAPSHOT_THRESHOLD,
	SNAPSHOT_RETENTION_COUNT,
	CACHE_LIMIT,
	HOST_IP,
	ELECTION_TIMEOUT,
	HEARTBEAT_TIMEOUT,
	LEADER_LEASE_TIMEOUT,
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

	raftId := v.GetString(raftNodeId)
	if raftId == "" {
		log.Fatalf("RAFT_NODE_ID is required")
	}

	nodeTTL := v.GetInt64(NODE_TTL)
	if nodeTTL == 0 {
		nodeTTL = 5 * 60
	}

	raftPortNum := v.GetInt(raftPort)
	if raftPortNum == 0 {
		raftPortNum = 1111
	}

	volDir := v.GetString(raftVolDir)

	if volDir == "" {
		log.Fatalf("RAFT_VOL_DIR is required")
	}

	redisIP := v.GetString(REDIS_IP)
	if raftId == "" {
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

	lgr := hclog.New(&hclog.LoggerOptions{
		Name:  "sybline-logger",
		Level: hclog.LevelFromString("debug"),
	})

	queueMan := core.NewQueueManager(nodeTTL)
	broker := core.NewBroker(queueMan, lgr)
	consumer := core.NewConsumerManager(queueMan, lgr)

	tDur := time.Second * time.Duration(tokenDuration)
	authManger := auth.NewAuthManager(sessHandler, &auth.UuidGen{}, &auth.ByteGenerator{}, tDur)
	authManger.CreateUser("sybline", auth.GenerateHash("sybline", salt))

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(raftId)

	electionTimeout := v.GetInt(ELECTION_TIMEOUT)
	if electionTimeout == 0 {
		electionTimeout = 2
	}

	heartbeatTimeout := v.GetInt(HEARTBEAT_TIMEOUT)
	if heartbeatTimeout == 0 {
		heartbeatTimeout = 2
	}

	leaderLeaseTimeout := v.GetInt(LEADER_LEASE_TIMEOUT)
	if leaderLeaseTimeout == 0 {
		leaderLeaseTimeout = 2
	}

	raftConfig.ElectionTimeout = time.Second * time.Duration(electionTimeout)
	raftConfig.HeartbeatTimeout = time.Second * time.Duration(heartbeatTimeout)
	raftConfig.LeaderLeaseTimeout = time.Second * time.Duration(leaderLeaseTimeout)

	snapshotThreshold := v.GetUint64(SNAPSHOT_THRESHOLD)
	if snapshotThreshold == 0 {
		snapshotThreshold = 10000
	}

	raftConfig.SnapshotThreshold = snapshotThreshold

	snapshotRetentionCount := v.GetInt(SNAPSHOT_RETENTION_COUNT)
	if snapshotRetentionCount == 0 {
		snapshotRetentionCount = 3
	}

	snapshotStore, err := raft.NewFileSnapshotStore(volDir, snapshotRetentionCount, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	batchLimit := v.GetUint64(BATCH_LIMIT)
	if batchLimit == 0 {
		batchLimit = 1000
	}

	// Wrap the store in a LogCache to improve performance.
	cacheLimit := v.GetInt(CACHE_LIMIT)
	if cacheLimit == 0 {
		cacheLimit = 100
	}

	store := fsm.NewStableStore(batchLimit, uint64(cacheLimit))

	fsmStore, err := fsm.NewSyblineFSM(broker, consumer, authManger, queueMan, store)
	if err != nil {
		log.Fatal(err)
	}

	cacheStore, err := raft.NewLogCache(cacheLimit, store)
	if err != nil {
		log.Fatal(err)
	}

	raftBinAddr := fmt.Sprintf("%s:%d", hostIP, raftPortNum)

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftBinAddr)
	if err != nil {
		log.Fatal(err)
	}

	transport, err := raft.NewTCPTransport(raftBinAddr, tcpAddr, 3, 10*time.Second, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	raftConfig.Logger = lgr

	raftServer, err := raft.NewRaft(raftConfig, fsmStore, cacheStore, store, snapshotStore, transport)
	if err != nil {
		log.Fatal(err)
	}

	nodes := strings.Split(v.GetString(nodes), ",")
	addresses := strings.Split(v.GetString(addresses), ",")

	servers := []raft.Server{
		{
			ID:      raft.ServerID(raftConfig.LocalID),
			Address: transport.LocalAddr(),
		},
	}

	for i, node := range nodes {
		servers = append(servers, raft.Server{
			Suffrage: raft.Voter,
			ID:       raft.ServerID(node),
			Address:  raft.ServerAddress(addresses[i]),
		})
	}

	// always starts single server as a leader
	configuration := raft.Configuration{
		Servers: servers,
	}

	raftServer.BootstrapCluster(configuration)
	grpcServer := grpc.NewServer(opts...)

	log.Println("listening on port: ", port)
	grpcServer.RegisterService(&handler.MQEndpoints_ServiceDesc, handler.NewServer(authManger, raftServer, salt))
	grpcServer.Serve(lis)
}
