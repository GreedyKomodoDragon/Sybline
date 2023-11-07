package e2e_test

import (
	"context"
	"fmt"
	"testing"

	client "github.com/GreedyKomodoDragon/sybline-go/handler"
	"github.com/cucumber/godog"
)

type messageBrokerCtxKey struct {
	client client.SyblineClient
}

func syblineClusterIsRunning(ctx context.Context) (context.Context, error) {
	// Implement logic to ensure the sybline cluster is running
	return ctx, nil
}

func thereIsAQueueWithNameConnectToBroker(ctx context.Context, queue, broker string) (context.Context, error) {
	// Implement logic to create a queue and connect it to the broker
	return ctx, nil
}

func messagesAreSentToBroker(ctx context.Context, messages int, broker string) (context.Context, error) {
	// Implement logic to send messages to the specified broker
	return ctx, nil
}

func consumerShouldBeAbleToGetAllMessagesFromQueue(ctx context.Context, expectedMessages int, queue string) error {
	// Implement logic to verify the consumer gets all messages from the specified queue
	// Use assertions or your broker's API to check this
	fmt.Printf("Consumer gets all messages from queue %s\n", queue)
	return nil
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^sybline cluster is running$`, syblineClusterIsRunning)
	sc.Step(`^there is a queue with name "([^"]*)" connect to broker "([^"]*)"$`, thereIsAQueueWithNameConnectToBroker)
	sc.Step(`^(\d+) messages are sent to "([^"]*)"$`, messagesAreSentToBroker)
	sc.Step(`^consumer should be able to get all (\d+) messages from queue "([^"]*)"$`, consumerShouldBeAbleToGetAllMessagesFromQueue)
}

func main() {
	TestFeatures(&testing.T{})
}
