package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sybline/pkg/auth"
	"sybline/pkg/config"
	"sybline/pkg/core"
	"sybline/pkg/fsm"
	"sybline/pkg/handler"
	"sybline/pkg/rbac"
	"sybline/pkg/rest"
	"sybline/pkg/rpc"

	"time"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	serverPort string = "SERVER_PORT"
	REST_PORT  string = "PROM_PORT"
	raftNodeId string = "RAFT_NODE_ID"

	TLS_ENABLED     string = "TLS_ENABLED"
	TLS_VERIFY_SKIP string = "TLS_VERIFY_SKIP"

	nodes              string = "NODES"
	ADDRESSES          string = "ADDRESSES"
	SNAPSHOT_THRESHOLD string = "SNAPSHOT_THRESHOLD"

	ELECTION_TIMEOUT  string = "ELECTION_TIMEOUT"
	HEARTBEAT_TIMEOUT string = "HEARTBEAT_TIMEOUT"

	STREAM_BUILD_TIMEOUT  string = "STREAM_BUILD_TIMEOUT"
	STREAM_BUILD_ATTEMPTS string = "STREAM_BUILD_ATTEMPTS"
	APPEND_TIMEOUT        string = "APPEND_TIMEOUT"
	RAFT_BATCH_LIMIT      string = "RAFT_BATCH_LIMIT"

	TOKEN_DURATION string = "TOKEN_DURATION"

	NODE_TTL string = "NODE_TTL"
	SALT     string = "SALT"

	K8S_AUTO         string = "K8S_AUTO"
	STATEFULSET_NAME string = "STATEFULSET_NAME"
	REPLICA_COUNT    string = "REPLICA_COUNT"
)

var confKeys = []string{
	serverPort,
	raftNodeId,
	nodes,
	ADDRESSES,
	SNAPSHOT_THRESHOLD,
	ELECTION_TIMEOUT,
	HEARTBEAT_TIMEOUT,
	TOKEN_DURATION,
	TLS_ENABLED,
	SALT,
}

func createTLSConfig(caCertFile, certFile, keyFile string, skipVerification bool) (*tls.Config, error) {
	// Load the CA certificate to validate the server's certificate.
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Create a certificate pool and add the CA certificate.
	certPool := x509.NewCertPool()
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}
	certPool.AppendCertsFromPEM(caCert)

	// Create a TLS configuration with the certificate and key.
	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            certPool,
		InsecureSkipVerify: skipVerification,
	}, nil
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	v := viper.New()
	v.AutomaticEnv()

	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal().Err(err)
		return
	}

	nodeTTL := v.GetInt64(NODE_TTL)
	if nodeTTL == 0 {
		nodeTTL = 5 * 60
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

	raftBatchLimit := v.GetInt(RAFT_BATCH_LIMIT)
	if raftBatchLimit == 0 {
		raftBatchLimit = 1000
	}

	salt := v.GetString(SALT)
	if salt == "" {
		log.Fatal().Msg("SALT is required")
	}

	port := v.GetInt(serverPort)
	if port == 0 {
		port = 2221
	}

	caCertFile := "./cert/ca-cert.pem"
	certFile := "./cert/cert.pem"
	keyFile := "./cert/key.pem"

	skipVerification := v.GetBool(TLS_VERIFY_SKIP)
	isTLSEnabled := v.GetBool(TLS_ENABLED)

	var addresses []string
	var ids []string
	var raftId uint64

	k8sAuto := v.GetBool(K8S_AUTO)
	if k8sAuto {
		statefulsetName := v.GetString(STATEFULSET_NAME)
		if statefulsetName == "" {
			log.Fatal().Msg("statefulsetName is required if using K8S_AUTO")
		}

		replicaCount := v.GetInt(REPLICA_COUNT)
		if replicaCount < 1 {
			log.Fatal().Msg("REPLICA_COUNT must be greater than 0")
		}

		addresses, ids, raftId = config.KubernetesAutoConfig(replicaCount, os.Getenv("HOSTNAME"), statefulsetName, port)

	} else {
		addresses = strings.Split(v.GetString(ADDRESSES), ",")

		ids = strings.Split(v.GetString(nodes), ",")
		if len(ids) != len(addresses) {
			log.Fatal().Msg("addresses and ids must be same length")
		}

		raftId = v.GetUint64(raftNodeId)
		if raftId == 0 {
			log.Fatal().Msg("RAFT_NODE_ID is required")
		}

	}

	// Create a TLS configuration.
	var tlsConfig *tls.Config
	if isTLSEnabled {
		tlsConf, err := createTLSConfig(caCertFile, certFile, keyFile, skipVerification)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create TLS config")
		}

		tlsConfig = tlsConf
	}

	sessHandler := auth.NewSessionHandler()

	restPort := v.GetInt(REST_PORT)
	if restPort == 0 {
		restPort = 7878
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Msg("failed to listen: " + err.Error())
	}

	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	prometheus.MustRegister(srvMetrics)
	panicsTotal := promauto.NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})

	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		return status.Errorf(codes.Internal, "%s", p)
	}

	servers := []raft.Server{}
	for i, address := range addresses {
		id, err := strconv.ParseUint(ids[i], 10, 64)
		if err != nil {
			log.Fatal().Msg("invalid id passed in")
		}

		clientOpts := []grpc.DialOption{}
		if isTLSEnabled {
			clientOpts = append(clientOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		} else {
			clientOpts = append(clientOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		servers = append(servers, raft.Server{
			Address: address,
			Id:      id,
			Opts:    clientOpts,
		})
	}

	tDur := time.Second * time.Duration(tokenDuration)
	authManger, err := auth.NewAuthManager(sessHandler, &auth.UuidGen{}, &auth.ByteGenerator{}, tDur, servers, salt)
	if err != nil {
		log.Fatal().Err(err)
	}

	// log directory - Create a folder/directory at a full qualified path
	err = os.MkdirAll("node_data/logs", 0755)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Fatal().Err(err)
	}

	queueMan := core.NewQueueManager(nodeTTL)
	broker := core.NewBroker(queueMan)
	consumer := core.NewConsumerManager(queueMan)

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

	rbacManager := rbac.NewRoleManager()

	// Gives sybline all the permissions
	rbacManager.AssignRole("sybline", "ROOT")

	fsmStore, err := fsm.NewSyblineFSM(broker, consumer, authManger, queueMan, rbacManager)
	if err != nil {
		log.Fatal().Err(err)
	}

	logStore, err := raft.NewLogGlobStore(snapshotThreshold)
	if err != nil {
		log.Fatal().Msg("failed to create logstore: " + err.Error())
	}

	raftServer := raft.NewRaftServer(fsmStore, logStore, &raft.Configuration{
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

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			rpc.IsLeader(raftServer),
			rpc.AuthenticationInterceptor(authManger),
			// Order matters e.g. tracing interceptor have to create span first for the later exemplars to work.
			otelgrpc.UnaryServerInterceptor(),
			srvMetrics.UnaryServerInterceptor(),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	}

	if isTLSEnabled {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcServer := grpc.NewServer(opts...)
	grpcServer.RegisterService(&auth.Session_ServiceDesc, auth.NewSessionServer(sessHandler))

	authManger.CreateUser("sybline", auth.GenerateHash("sybline", salt))

	raftServer.Start(grpcServer)

	batcher := fsm.NewBatcher(raftServer, raftBatchLimit)
	go batcher.ProcessLogs()

	hand := handler.NewHandler(rbacManager, authManger, raftServer, batcher, salt)
	grpcAPI := rpc.NewServer(hand)

	app := rest.NewRestServer(broker, authManger, rbacManager, queueMan, raftServer, hand)
	if isTLSEnabled {
		ln, _ := net.Listen("tcp", ":"+strconv.Itoa(restPort))
		ln = tls.NewListener(ln, tlsConfig)
		go app.Listener(ln)
	} else {
		go app.Listen(":" + strconv.Itoa(restPort))
	}

	log.Info().Int("port", port).Msg("listening on port")
	grpcServer.RegisterService(&rpc.MQEndpoints_ServiceDesc, grpcAPI)
	grpcServer.Serve(lis)
}
