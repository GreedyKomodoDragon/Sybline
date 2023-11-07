Feature: Message Broker

  Scenario: Can send Messages to a single queue
    Given sybline cluster is running
    And there is a queue with name "queue" connect to broker "broker"
    When 5 messages are sent to "broker"
    Then consumer should be able to get all 5 messages from queue "queue"

  Scenario: Can send Messages to multiple queues
    Given sybline cluster is running
    And there is a queue with name "queue" connect to broker "broker"
    And there is a queue with name "queue2" connect to broker "broker"
    When 5 messages are sent to "broker"
    Then consumer should be able to get all 5 messages from queue "queue"
    And consumer should be able to get all 5 messages from queue "queue2"

  