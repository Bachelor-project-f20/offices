package test

import (
	"fmt"
	"log"
	"testing"
	"time"

	etg "github.com/Bachelor-project-f20/eventToGo"
	"github.com/Bachelor-project-f20/offices/pkg/creating"
	"github.com/Bachelor-project-f20/offices/pkg/deleting"
	eventHandler "github.com/Bachelor-project-f20/offices/pkg/event"
	"github.com/Bachelor-project-f20/offices/pkg/updating"
	"github.com/Bachelor-project-f20/shared/config"
	models "github.com/Bachelor-project-f20/shared/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/golang/protobuf/proto"
)

var eventEmitter etg.EventEmitter
var eventListener etg.EventListener
var eventChan <-chan models.Event
var creatingService creating.Service

var updatingService updating.Service
var deletingService deleting.Service
var svc *sns.SNS

func TestServiceSetup(t *testing.T) {
	//AnonymousCredentials for the mock SNS instance
	//SSL disabled, because it's easier when testing
	//localhost:991 is where the fake SNS container should be running
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Credentials: credentials.AnonymousCredentials, Endpoint: aws.String("http://localhost:9911"), Region: aws.String("us-east-1"), DisableSSL: aws.Bool(true)},
	}))

	svc = sns.New(sess)

	incomingEvents := []string{
		models.OfficeEvents_CREATE_OFFICE.String(),
		models.OfficeEvents_DELETE_OFFICE.String(),
		models.OfficeEvents_UPDATE_OFFICE.String()}

	outgoingEvents := []string{
		models.OfficeEvents_OFFICE_CREATED.String(),
		models.OfficeEvents_OFFICE_UPDATED.String(),
		models.OfficeEvents_OFFICE_DELETED.String()}

	incomingAndOutgoingEvents := append(incomingEvents, outgoingEvents...)

	configRes, err := config.ConfigService(
		"configFile",
		config.ConfigValues{
			UseEmitter:        true,
			UseListener:       true,
			MessageBrokerType: etg.SNS,
			SNSClient:         svc,
			Events:            incomingAndOutgoingEvents,
			UseOutbox:         true,
			OutboxModels:      []interface{}{models.Office{}, models.Address{}},
		},
	)
	if err != nil {
		log.Fatalln("configuration failed, error: ", err)
		panic("configuration failed")
	}
	eventEmitter = configRes.EventEmitter
	eventListener = configRes.EventListener

	eventChan, _, err = eventListener.Listen(incomingAndOutgoingEvents...)

	if err != nil {
		fmt.Printf("Creation of subscriptions failed, error: %v \n", err)
		t.Error(err)
	}

	creatingService = creating.NewService(configRes.Outbox)
	updatingService = updating.NewService(configRes.Outbox)
	deletingService = deleting.NewService(configRes.Outbox)
}

func test(t *testing.T) {
	testingChan := make(chan eventHandler.TestObject)
	defer close(testingChan)
	go func() {
		eventHandler.TestingStartEventHandler(
			testingChan,
			eventChan,
			creatingService,
			updatingService,
			deletingService,
		)
	}()
	testResult := <-testingChan
	if !testResult.Ok {
		fmt.Println("ERROR")
		t.Error(testResult.Err)
	}
	testingChan <- eventHandler.TestObject{}
}

func TestCreateRequestHandling(t *testing.T) {
	fmt.Println("TestCreateRequestHandling")

	address := &models.Address{
		ID:       "test_address",
		RoadName: "test",
		Number:   1,
		ZipCode:  1111,
	}

	event := models.OfficeCreated{
		Office: &models.Office{
			ID:        "test",
			Name:      "test_office_created",
			Address:   address,
			AddressID: address.ID,
		},
	}

	marshalEvent, err := proto.Marshal(&event)

	if err != nil {
		fmt.Printf("Error marshalling new office, error: %v \n", err)
		t.Error(err)
	}

	creationRequest := models.Event{
		ID:        "test",
		Publisher: "offices_test",
		EventName: models.OfficeEvents_CREATE_OFFICE.String(),
		Timestamp: time.Now().UnixNano(),
		Payload:   marshalEvent,
	}

	eventEmitter.Emit(creationRequest)
	test(t)
	time.Sleep(2 * time.Second)
}

func TestUpdateRequestHandling(t *testing.T) {
	fmt.Println("TestUpdateRequestHandling")

	address := &models.Address{
		ID:       "test_address",
		RoadName: "test",
		Number:   1,
		ZipCode:  1111,
	}

	event := models.OfficeUpdated{
		Office: &models.Office{
			ID:        "test",
			Name:      "test_office_updated",
			Address:   address,
			AddressID: address.ID,
		},
	}

	marshalEvent, err := proto.Marshal(&event)

	if err != nil {
		fmt.Printf("Error marshalling new office, error: %v \n", err)
		t.Error(err)
	}

	updateRequest := models.Event{
		ID:        "test",
		Publisher: "offices_test",
		EventName: models.OfficeEvents_UPDATE_OFFICE.String(),
		Timestamp: time.Now().UnixNano(),
		Payload:   marshalEvent,
	}

	eventEmitter.Emit(updateRequest)
	test(t)
	time.Sleep(2 * time.Second)
}

func TestDeleteRequestHandling(t *testing.T) {
	fmt.Println("TestDeleteRequestHandling")

	address := &models.Address{
		ID:       "test_address",
		RoadName: "test",
		Number:   1,
		ZipCode:  1111,
	}

	event := models.OfficeDeleted{
		Office: &models.Office{
			ID:        "test",
			Name:      "test_office_updated",
			Address:   address,
			AddressID: address.ID,
		},
	}

	marshalEvent, err := proto.Marshal(&event)

	if err != nil {
		fmt.Printf("Error marshalling new office, error: %v \n", err)
		t.Error(err)
	}

	deletionRequest := models.Event{
		ID:        "test",
		Publisher: "offices_test",
		EventName: models.OfficeEvents_DELETE_OFFICE.String(),
		Timestamp: time.Now().UnixNano(),
		Payload:   marshalEvent,
	}

	eventEmitter.Emit(deletionRequest)
	test(t)
	time.Sleep(2 * time.Second)
}
