package main

import (
	"context"
	"encoding/json"
	"errors" // Add this import
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func fetchOrders(ctx context.Context, client *azservicebus.Client, queueName string) ([]Order, error) {
	var orders []Order

	receiver, err := client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		log.Fatalf("failed to create receiver: %v", err)
		return nil, err
	}
	defer receiver.Close(ctx)

	messages, err := receiver.ReceiveMessages(ctx, 10, nil)
	if err != nil {
		log.Fatalf("failed to receive messages: %v", err)
		return nil, err
	}

	for _, message := range messages {
		log.Printf("message received: %s\n", string(message.Body))

		var order Order
		if err := json.Unmarshal(message.Body, &order); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			continue
		}

		orders = append(orders, order)

		if err := receiver.CompleteMessage(ctx, message, nil); err != nil {
			log.Printf("failed to complete message: %v", err)
		}
	}

	return orders, nil
}

func getOrdersFromQueue() ([]Order, error) {
	ctx := context.Background()

	connectionString := os.Getenv("AZURE_SERVICE_BUS_CONNECTION_STRING")
	if connectionString == "" {
		log.Printf("AZURE_SERVICE_BUS_CONNECTION_STRING is not set")
		return nil, errors.New("AZURE_SERVICE_BUS_CONNECTION_STRING is not set")
	}

	queueName := os.Getenv("ORDER_QUEUE_NAME")
	if queueName == "" {
		log.Printf("ORDER_QUEUE_NAME is not set")
		return nil, errors.New("ORDER_QUEUE_NAME is not set")
	}

	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Fatalf("failed to create service bus client: %v", err)
		return nil, err
	}
	defer client.Close(ctx)

	return fetchOrders(ctx, client, queueName)
}
