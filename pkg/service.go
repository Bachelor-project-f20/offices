package pkg

import (
	"fmt"
	"log"
	"net/http"

	etg "github.com/Bachelor-project-f20/eventToGo"
	"github.com/Bachelor-project-f20/offices/pkg/creating"
	"github.com/Bachelor-project-f20/offices/pkg/deleting"
	handler "github.com/Bachelor-project-f20/offices/pkg/event"
	"github.com/Bachelor-project-f20/offices/pkg/updating"
	"github.com/Bachelor-project-f20/shared/config"
	models "github.com/Bachelor-project-f20/shared/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var configFile string = "configPath"

func Run() {

	// AnonymousCredentials for the mock SNS instance
	// SSL disabled, because it's easier when testing
	// localhost:991 is where the fake SNS container should be running
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Credentials: credentials.AnonymousCredentials, Endpoint: aws.String("http://localhost:9911"), Region: aws.String("us-east-1"), DisableSSL: aws.Bool(true)},
	}))

	svc := sns.New(sess)

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

	eventChan, _, err := configRes.EventListener.Listen(incomingEvents...)

	if err != nil {
		log.Fatalf("Creation of subscriptions failed, error: %v \n", err)
	}

	creatingService := creating.NewService(configRes.Outbox)
	updatingService := updating.NewService(configRes.Outbox)
	deletingService := deleting.NewService(configRes.Outbox)

	go func() {
		fmt.Println("Serving metrics API")

		h := http.NewServeMux()
		h.Handle("/metrics", promhttp.Handler())

		http.ListenAndServe(":9191", h)
	}()

	handler.StartEventHandler(
		eventChan,
		creatingService,
		updatingService,
		deletingService)
}
