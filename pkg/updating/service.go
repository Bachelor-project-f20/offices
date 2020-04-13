package updating

import (
	"log"
	"time"

	ob "github.com/Bachelor-project-f20/go-outbox"
	models "github.com/Bachelor-project-f20/shared/models"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
)

type Service interface {
	UpdateOffice(requestEvent models.Event) error
}

type service struct {
	ob ob.Outbox
}

func NewService(outbox ob.Outbox) Service {
	return &service{outbox}
}

func (srv *service) UpdateOffice(requestEvent models.Event) error {

	event := &models.CreateOffice{}
	err := proto.Unmarshal(requestEvent.Payload, event)
	office := event.Office

	if err != nil {
		log.Fatalf("Error with proto: %v \n", err)
		return err
	}

	officeUpdatedEvent := &models.OfficeUpdated{
		Office: office,
	}

	marshalEvent, err := proto.Marshal(officeUpdatedEvent)

	if err != nil {
		log.Fatalf("Error with proto: %v \n", err)
		return err
	}

	id, _ := uuid.NewV4()
	idAsString := id.String()

	updateEvent := models.Event{
		ID:        idAsString,
		Publisher: models.OfficeService_OFFICES.String(),
		EventName: models.OfficeEvents_OFFICE_UPDATED.String(),
		Timestamp: time.Now().UnixNano(),
		Payload:   marshalEvent,
	}

	err = srv.ob.Update(office, updateEvent)

	if err != nil {
		log.Fatal("Error during update of office. Err: ", err)
	}

	return err
}
