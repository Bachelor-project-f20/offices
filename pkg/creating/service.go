package creating

import (
	"log"
	"time"

	ob "github.com/Bachelor-project-f20/go-outbox"
	models "github.com/Bachelor-project-f20/shared/models"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
)

type Service interface {
	CreateOffice(requestEvent models.Event) error
}

type service struct {
	outbox ob.Outbox
}

func NewService(outbox ob.Outbox) Service {
	return &service{outbox}
}

//Do something with address and such?
func (srv *service) CreateOffice(requestEvent models.Event) error {

	event := &models.CreateOffice{}
	err := proto.Unmarshal(requestEvent.Payload, event)
	office := event.Office

	if err != nil {
		log.Fatalf("Error with proto: %v \n", err)
		return err
	}

	officeCreatedEvent := &models.OfficeCreated{
		Office: office,
	}
	marshalEvent, err := proto.Marshal(officeCreatedEvent)
	if err != nil {
		log.Fatalf("Error with proto: %v \n", err)
		return err
	}

	id, _ := uuid.NewV4()
	idAsString := id.String()

	creationEvent := models.Event{
		ID:        idAsString,
		Publisher: models.OfficeService_OFFICES.String(),
		EventName: models.OfficeEvents_OFFICE_CREATED.String(),
		Timestamp: time.Now().UnixNano(),
		Payload:   marshalEvent,
	}

	err = srv.outbox.Insert(office, creationEvent)

	if err != nil {
		log.Fatal("Error during creation of office. Err: ", err)
	}
	return err
}
