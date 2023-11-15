Feature: Message Broker

  Scenario: Can send Messages to a single queue
    Given sybline cluster is running
    And there is a queue connect to broker
    When 5 messages are sent to broker
    Then consumer should be able to get all 5 messages from queue

  Scenario: Can send Messages to multiple queues
    Given sybline cluster is running
    And there is a queue connect to broker
    And there is a another queue connect to the same broker
    When 5 messages are sent to broker
    Then consumer should be able to get all 5 messages from queue
    And consumer should be able to get all 5 messages from second queue

  