package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func getOrdersFromQueue() ([]Order, error) {
	ctx := context.Background()
	var orders []Order

	// Get queue name from environment variable
	orderQueueName := os.Getenv("ORDER_QUEUE_NAME")
	if orderQueueName == "" {
			log.Printf("ORDER_QUEUE_NAME is not set")
			return nil, errors.New("ORDER_QUEUE_NAME is not set")
	}

	// Check if using Workload Identity Authentication
	useWorkloadIdentityAuth := os.Getenv("USE_WORKLOAD_IDENTITY_AUTH")
	if useWorkloadIdentityAuth == "" {
			useWorkloadIdentityAuth = "false"
	}

	// Get Service Bus hostname
	orderQueueHostName := os.Getenv("ORDER_QUEUE_HOSTNAME")
	if orderQueueHostName == "" {
			log.Printf("ORDER_QUEUE_HOSTNAME is not set")
			return nil, errors.New("ORDER_QUEUE_HOSTNAME is not set")
	}

	if useWorkloadIdentityAuth == "true" {
			// Use Azure Workload Identity Authentication
			cred, err := azidentity.NewDefaultAzureCredential(nil)
			if err != nil {
					log.Fatalf("failed to obtain a workload identity credential: %v", err)
			}

			client, err := azservicebus.NewClient(orderQueueHostName, cred, nil)
			if err != nil {
					log.Fatalf("failed to create service bus client: %v", err)
			}
			defer client.Close(ctx)

			receiver, err := client.NewReceiverForQueue(orderQueueName, nil)
			if err != nil {
					log.Fatalf("failed to create receiver: %v", err)
			}
			defer receiver.Close(ctx)

			messages, err := receiver.ReceiveMessages(ctx, 10, nil)
			if err != nil {
					log.Fatalf("failed to receive messages: %v", err)
			}

			for _, message := range messages {
					log.Printf("message received: %s\n", string(message.Body))

					var order Order
					if err := json.Unmarshal(message.Body, &order); err != nil {
							log.Printf("failed to unmarshal message: %v", err)
							return nil, err
					}

					orders = append(orders, order)

					if err := receiver.CompleteMessage(ctx, message, nil); err != nil {
							log.Printf("failed to complete message: %v", err)
					}
			}
	} else {
			// Use Shared Access Policy Authentication
			connectionString := os.Getenv("AZURE_SERVICE_BUS_CONNECTION_STRING")
			if connectionString == "" {
					log.Printf("AZURE_SERVICE_BUS_CONNECTION_STRING is not set")
					return nil, errors.New("AZURE_SERVICE_BUS_CONNECTION_STRING is not set")
			}

			client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
			if err != nil {
					log.Fatalf("failed to create service bus client: %v", err)
			}
			defer client.Close(ctx)

			receiver, err := client.NewReceiverForQueue(orderQueueName, nil)
			if err != nil {
					log.Fatalf("failed to create receiver: %v", err)
			}
			defer receiver.Close(ctx)

			messages, err := receiver.ReceiveMessages(ctx, 10, nil)
			if err != nil {
					log.Fatalf("failed to receive messages: %v", err)
			}

			for _, message := range messages {
					log.Printf("message received: %s\n", string(message.Body))

					var order Order
					if err := json.Unmarshal(message.Body, &order); err != nil {
							log.Printf("failed to unmarshal message: %v", err)
							return nil, err
					}

					orders = append(orders, order)

					if err := receiver.CompleteMessage(ctx, message, nil); err != nil {
							log.Printf("failed to complete message: %v", err)
					}
			}
	}

	return orders, nil
}

func unmarshalOrderFromQueue(data []byte) (Order, error) {
	var order Order

	err := json.Unmarshal(data, &order)
	if err != nil {
		log.Printf("failed to unmarshal order: %v\n", err)
		return Order{}, err
	}

	// add orderkey to order
	order.OrderID = strconv.Itoa(rand.Intn(100000))

	// set the status to pending
	order.Status = Pending

	return order, nil
}
