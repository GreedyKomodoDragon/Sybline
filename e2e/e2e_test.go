package e2e_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/GreedyKomodoDragon/sybline-go/handler"
	client "github.com/GreedyKomodoDragon/sybline-go/handler"
	"github.com/cucumber/godog"
)

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

type messageBrokerCtxKey struct {
	client   client.SyblineClient
	queueOne string
	queueTwo string
	broker   string
}

var broker messageBrokerCtxKey

func syblineClusterIsRunning(ctx context.Context) (context.Context, error) {
	passwordManager := client.NewUnsecurePasswordManager()
	passwordManager.SetPassword("sybline", "sybline")

	cli, err := client.NewBasicSyblineClient([]string{"localhost:2221", "localhost:2222", "localhost:2223"}, passwordManager, handler.Config{
		TimeoutSec:      5,
		TimeoutAttempts: 3,
	})
	if err != nil {
		return ctx, err
	}

	if err := cli.Login(ctx, "sybline"); err != nil {
		return ctx, err
	}

	broker = messageBrokerCtxKey{
		client: cli,
	}

	return ctx, nil
}

func thereIsAQueueWithNameConnectToBroker(ctx context.Context) (context.Context, error) {
	broker.broker = randomString(10)
	broker.queueOne = randomString(10)

	if err := broker.client.CreateQueue(ctx, broker.broker, broker.queueOne, 5, 5, false); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func messagesAreSentToBroker(ctx context.Context, messages int) (context.Context, error) {
	msgs := []client.Message{}
	for i := 0; i < messages; i++ {
		msgs = append(msgs, client.Message{
			Rk:   broker.broker,
			Data: []byte("message_" + strconv.Itoa(i)),
		})
	}

	if err := broker.client.SubmitBatchMessage(ctx, msgs); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func consumerShouldBeAbleToGetAllMessagesFromQueue(ctx context.Context, expectedMessages int) error {
	msgs, err := broker.client.GetMessages(ctx, broker.queueOne, uint32(expectedMessages))
	if err != nil {
		return err
	}

	if len(msgs) != expectedMessages {
		return fmt.Errorf("was only able to get %v messages", len(msgs))
	}

	return nil
}

func thereIsAQueueWithNameConnectToSameBroker(ctx context.Context) error {
	broker.queueTwo = randomString(10)

	return broker.client.CreateQueue(ctx, broker.broker, broker.queueTwo, 5, 5, false)
}

func consumerShouldBeAbleToGetAllMessagesFromSecondQueue(ctx context.Context, expectedMessages int) error {
	msgs, err := broker.client.GetMessages(ctx, broker.queueTwo, uint32(expectedMessages))
	if err != nil {
		return err
	}

	if len(msgs) != expectedMessages {
		return fmt.Errorf("was only able to get %v messages", len(msgs))
	}

	return nil
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^sybline cluster is running$`, syblineClusterIsRunning)
	sc.Step(`^there is a queue connect to broker$`, thereIsAQueueWithNameConnectToBroker)
	sc.Step(`^(\d+) messages are sent to broker$`, messagesAreSentToBroker)
	sc.Step(`^consumer should be able to get all (\d+) messages from queue$`, consumerShouldBeAbleToGetAllMessagesFromQueue)
	sc.Step(`^there is a another queue connect to the same broker$`, thereIsAQueueWithNameConnectToSameBroker)
	sc.Step(`^consumer should be able to get all (\d+) messages from second queue$`, consumerShouldBeAbleToGetAllMessagesFromSecondQueue)
}
