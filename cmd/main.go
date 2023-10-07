package main

import (
	"context"
	"fmt"
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	serverPort         string = "SERVER_PORT"
	raftNodeId         string = "RAFT_NODE_ID"
	TLS_ENABLED        string = "TLS_ENABLED"
	nodes              string = "NODES"
	addresses          string = "ADDRESSES"
	SNAPSHOT_THRESHOLD string = "SNAPSHOT_THRESHOLD"

	ELECTION_TIMEOUT  string = "ELECTION_TIMEOUT"
	HEARTBEAT_TIMEOUT string = "HEARTBEAT_TIMEOUT"

	STREAM_BUILD_TIMEOUT  string = "STREAM_BUILD_TIMEOUT"
	STREAM_BUILD_ATTEMPTS string = "STREAM_BUILD_ATTEMPTS"
	APPEND_TIMEOUT        string = "APPEND_TIMEOUT"

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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var v = viper.New()
	v.AutomaticEnv()

	if err := v.BindEnv(confKeys...); err != nil {
		// log.Fatal(err)
		return
	}

	raftId := v.GetUint64(raftNodeId)
	if raftId == 0 {
		log.Fatal().Msg("RAFT_NODE_ID is required")
	}

	nodeTTL := v.GetInt64(NODE_TTL)
	if nodeTTL == 0 {
		nodeTTL = 5 * 60
	}

	redisIP := v.GetString(REDIS_IP)
	if redisIP == "" {
		log.Fatal().Msg("REDIS_IP is required")
	}

	tokenDuration := v.GetInt64(TOKEN_DURATION)
	if tokenDuration == 0 {
		tokenDuration = 1800
	}

	streamBuildTimeout := v.GetInt(STREAM_BUILD_TIMEOUT)
	if streamBuildTimeout == 0 {
		streamBuildTimeout = 2
	}

	streamBuildAttempts := v.GetInt(STREAM_BUILD_ATTEMPTS)
	if streamBuildAttempts == 0 {
		streamBuildAttempts = 3
	}

	appendTimeout := v.GetInt(APPEND_TIMEOUT)
	if appendTimeout == 0 {
		appendTimeout = 3
	}

	salt := v.GetString(SALT)
	if salt == "" {
		log.Fatal().Msg("SALT is required")
	}

	redisPassword := v.GetString(REDIS_PASSWORD)
	redisDatabase := v.GetInt(REDIS_DATABASE)

	sessHandler, err := auth.NewRedisSession(redisIP, redisDatabase, redisPassword)
	if err != nil {
		log.Fatal().Msg("failed to start redis session: " + err.Error())
	}

	port := v.GetInt(serverPort)
	if port == 0 {
		port = 2221
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Msg("failed to listen: " + err.Error())
	}

	messages.Initialise()

	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)

	panicsTotal := promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})

	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		return status.Errorf(codes.Internal, "%s", p)
	}

	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			// Order matters e.g. tracing interceptor have to create span first for the later exemplars to work.
			otelgrpc.UnaryServerInterceptor(),
			srvMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			otelgrpc.StreamServerInterceptor(),
			srvMetrics.StreamServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	}

	isTLSEnabled := v.GetBool(TLS_ENABLED)

	if isTLSEnabled {
		creds, err := credentials.NewServerTLSFromFile("./cert/certificate.crt", "./cert/private.key")
		if err != nil {
			log.Fatal().Msg("Failed to generate credentials: " + err.Error())
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	// log directory - Create a folder/directory at a full qualified path
	err = os.MkdirAll("node_data/logs", 0755)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Fatal().Err(err)
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
		log.Fatal().Err(err)
	}

	logStore, err := raft.NewLogStore(snapshotThreshold)
	if err != nil {
		log.Fatal().Msg("failed to create logstore: " + err.Error())
	}
	grpcServer := grpc.NewServer(opts...)

	addresses := strings.Split(v.GetString(addresses), ",")

	servers := []raft.Server{}
	ids := strings.Split(v.GetString(nodes), ",")
	if len(ids) != len(addresses) {
		log.Fatal().Msg("addresses and ids must be same length")
	}

	for i, address := range addresses {
		id, err := strconv.ParseUint(ids[i], 10, 64)
		if err != nil {
			log.Fatal().Msg("invalid id passed in")
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
				StreamBuildTimeout:  time.Duration(streamBuildTimeout) * time.Second,
				StreamBuildAttempts: streamBuildAttempts,
				AppendTimeout:       time.Duration(appendTimeout) * time.Second,
			},
		},
		ElectionConfig: &raft.ElectionConfig{
			ElectionTimeout:  time.Duration(electionTimeout) * time.Second,
			HeartbeatTimeout: time.Duration(heartbeatTimeout) * time.Second,
		},
	})

	raftServer.Start()

	log.Info().Int("port", port).Msg("listening on port")
	grpcServer.RegisterService(&handler.MQEndpoints_ServiceDesc, handler.NewServer(authManger, raftServer, salt))
	grpcServer.Serve(lis)
}
