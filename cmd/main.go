package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/fsm"
	"sybline/pkg/handler"
	"sybline/pkg/rbac"

	"time"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	PROM_PORT  string = "PROM_PORT"
	raftNodeId string = "RAFT_NODE_ID"

	TLS_ENABLED     string = "TLS_ENABLED"
	TLS_VERIFY_SKIP string = "TLS_VERIFY_SKIP"

	TLS_ENABLED_PROM     string = "TLS_ENABLED_PROM"
	TLS_VERIFY_SKIP_PROM string = "TLS_VERIFY_SKIP_PROM"

	nodes              string = "NODES"
	addresses          string = "ADDRESSES"
	SNAPSHOT_THRESHOLD string = "SNAPSHOT_THRESHOLD"

	ELECTION_TIMEOUT  string = "ELECTION_TIMEOUT"
	HEARTBEAT_TIMEOUT string = "HEARTBEAT_TIMEOUT"

	STREAM_BUILD_TIMEOUT  string = "STREAM_BUILD_TIMEOUT"
	STREAM_BUILD_ATTEMPTS string = "STREAM_BUILD_ATTEMPTS"
	APPEND_TIMEOUT        string = "APPEND_TIMEOUT"

	TOKEN_DURATION string = "TOKEN_DURATION"

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

	var v = viper.New()
	v.AutomaticEnv()

	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal().Err(err)
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

	caCertFile := "./cert/ca-cert.pem"
	certFile := "./cert/cert.pem"
	keyFile := "./cert/key.pem"

	skipVerification := v.GetBool(TLS_VERIFY_SKIP)
	isTLSEnabled := v.GetBool(TLS_ENABLED)

	skipVerificationProm := v.GetBool(TLS_VERIFY_SKIP_PROM)
	isTLSEnabledProm := v.GetBool(TLS_ENABLED_PROM)

	sessHandler := auth.NewSessionHandler()

	port := v.GetInt(serverPort)
	if port == 0 {
		port = 2221
	}

	promPort := v.GetInt(PROM_PORT)
	if promPort == 0 {
		promPort = 8080
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

	// Expose Prometheus metrics endpoint.
	// Create an HTTP server for /metrics with TLS
	metricsServer := &http.Server{
		Addr:    ":" + strconv.Itoa(promPort), // Choose the desired port
		Handler: promhttp.Handler(),
	}

	if isTLSEnabledProm {
		tlsConfig, err := createTLSConfig(caCertFile, certFile, keyFile, skipVerificationProm)
		if err != nil {
			log.Fatal().Err(err)
		}

		metricsServer.TLSConfig = tlsConfig
	}

	go func() {
		log.Info().Int("port", promPort).Msg("Metrics server is running")

		if isTLSEnabledProm {
			if err := metricsServer.ListenAndServeTLS("", ""); err != nil {
				log.Fatal().Err(err).Msg("Failed to start metrics server with TLS")
			}
			return
		}

		if err := metricsServer.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start metrics server")
		}
	}()

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			// Order matters e.g. tracing interceptor have to create span first for the later exemplars to work.
			otelgrpc.UnaryServerInterceptor(),
			srvMetrics.UnaryServerInterceptor(),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			otelgrpc.StreamServerInterceptor(),
			srvMetrics.StreamServerInterceptor(),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	}

	if isTLSEnabled {
		tlsConfig, err := createTLSConfig(caCertFile, certFile, keyFile, skipVerification)
		if err != nil {
			log.Fatal().Msg("Failed to generate credentials: " + err.Error())
		}
		opts = []grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsConfig))}
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

		clientOpts := []grpc.DialOption{}
		if isTLSEnabled {
			// Create a TLS configuration.
			tlsConfig, err := createTLSConfig(caCertFile, certFile, keyFile, skipVerification)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to create TLS config")
			}

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

	grpcServer.RegisterService(&auth.Session_ServiceDesc, auth.NewSessionServer(sessHandler))

	tDur := time.Second * time.Duration(tokenDuration)
	authManger, err := auth.NewAuthManager(sessHandler, &auth.UuidGen{}, &auth.ByteGenerator{}, tDur, servers)
	if err != nil {
		log.Fatal().Err(err)
	}

	authManger.CreateUser("sybline", auth.GenerateHash("sybline", salt))

	rbacManager := rbac.NewRoleManager()

	// Gives sybline all the permissions
	rbacManager.AssignRole("sybline", "ROOT")

	fsmStore, err := fsm.NewSyblineFSM(broker, consumer, authManger, queueMan, rbacManager)
	if err != nil {
		log.Fatal().Err(err)
	}

	logStore, err := raft.NewLogStore(snapshotThreshold)
	if err != nil {
		log.Fatal().Msg("failed to create logstore: " + err.Error())
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
	grpcServer.RegisterService(&handler.MQEndpoints_ServiceDesc, handler.NewServer(rbacManager, authManger, raftServer, salt))
	grpcServer.Serve(lis)
}
