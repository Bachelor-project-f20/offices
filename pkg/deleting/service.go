package deleting

import (
	"log"
	"time"

	ob "github.com/Bachelor-project-f20/go-outbox"
	models "github.com/Bachelor-project-f20/shared/models"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
)

type Service interface {
	DeleteOffice(requestEvent models.Event) error
}

type service struct {
	ob ob.Outbox
}

func NewService(outbox ob.Outbox) Service {
	return &service{outbox}
}

func (srv *service) DeleteOffice(requestEvent models.Event) error {

	event := &models.DeleteOffice{}
	err := proto.Unmarshal(requestEvent.Payload, event)
	office := event.Office

	if err != nil {
		log.Fatalf("Error with proto: %v \n", err)
		return err
	}

	officeDeletedEvent := &models.OfficeDeleted{
		Office: office,
	}
	marshalEvent, err := proto.Marshal(officeDeletedEvent)

	if err != nil {
		log.Fatalf("Error with proto: %v \n", err)
		return err
	}

	id, _ := uuid.NewV4()
	idAsString := id.String()

	deletionEvent := models.Event{
		ID:        idAsString,
		Publisher: models.OfficeService_OFFICES.String(),
		EventName: models.OfficeEvents_OFFICE_DELETED.String(),
		Timestamp: time.Now().UnixNano(),
		Payload:   marshalEvent,
		ApiTag:    requestEvent.ApiTag,
	}

	err = srv.ob.Delete(office, deletionEvent)

	if err != nil {
		log.Fatal("Error during deletion of office. Err: ", err)
	}

	return err
}
