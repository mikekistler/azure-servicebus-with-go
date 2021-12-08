package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
)

var (
	client    *azservicebus.Client
	cred      *azidentity.DefaultAzureCredential
	namespace = os.Getenv("SERVICEBUS_NAMESPACE")
)

func init() {
	var err error
	cred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	client, err = azservicebus.NewClient(namespace, cred, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("\t foo send|receive queueName")
	os.Exit(1)
}

func main() {

	args := os.Args[1:] // exclude program
	if len(args) < 1 {
		usage()
	}

	switch args[0] {
	case "send":
		send(args[1:])
	case "receive":
		receive(args[1:])
	case "list-queues":
		list()
	default:
		usage()
	}
}

func send(argSlice []string) {
	if len(argSlice) < 1 {
		usage()
	}
	queueName := argSlice[0]

	qSender, err := client.NewSender(queueName, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	byteBuf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("%v", err)
	}

	message := azservicebus.Message{Body: byteBuf}
	err = qSender.SendMessage(context.TODO(), &message)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func receive(argSlice []string) {
	if len(argSlice) < 1 {
		usage()
	}
	queueName := argSlice[0]

	qReceiver, err := client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	messages, err := qReceiver.ReceiveMessages(context.TODO(), 1, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if len(messages) > 0 {
		message := messages[0]
		body, err := message.Body()
		if err != nil {
			log.Fatalf("%v", err)
		}
		fmt.Print(string(body))
		err = qReceiver.CompleteMessage(context.TODO(), message)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
}

func list() {
	adminClient, err := admin.NewClient(namespace, cred, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}
	queuePager := adminClient.ListQueues(nil)

	for queuePager.NextPage(context.TODO()) {
		for _, item := range queuePager.PageResponse().Items {
			fmt.Println(item.QueueName)
		}
	}

	if queuePager.Err() != nil {
		log.Fatalf("%v", queuePager.Err())
	}
}
