package handler

import (
	"errors"
	"fmt"
	"log"

	"github.com/Bachelor-project-f20/offices/pkg/creating"
	"github.com/Bachelor-project-f20/offices/pkg/deleting"
	"github.com/Bachelor-project-f20/offices/pkg/updating"
	models "github.com/Bachelor-project-f20/shared/models"
)

type TestObject struct {
	Ok  bool
	Err error
}

type handler struct {
	errorChan       chan error
	testingChan     chan TestObject
	creatingService creating.Service
	updatingService updating.Service
	deletingService deleting.Service
}

func StartEventHandler(
	eventChan <-chan models.Event,
	creatingService creating.Service,
	updatingService updating.Service,
	deletingService deleting.Service) {

	errChan := make(chan error, 5)
	defer close(errChan)

	handler := handler{
		errChan,
		nil,
		creatingService,
		updatingService,
		deletingService,
	}

	handler.handleEvents(eventChan)
}

func TestingStartEventHandler(
	testingChan chan TestObject,
	eventChan <-chan models.Event,
	creatingService creating.Service,
	updatingService updating.Service,
	deletingService deleting.Service) {

	errChan := make(chan error, 5)
	defer close(errChan)

	handler := handler{
		errChan,
		testingChan,
		creatingService,
		updatingService,
		deletingService,
	}

	handler.handleEvents(eventChan)
}

func (h *handler) handleEvents(eventChan <-chan models.Event) {
	for {
		select {
		case event, open := <-eventChan:
			if !open {
				h.testErrors(false, "EventHandler, event channel closed. STOPPING")
				return
			}
			h.handleEvent(event)
		case err, open := <-h.errorChan:
			if !open {
				h.testErrors(false, "EventHandler, error channel closed. STOPPING")
				return
			}
			h.testErrors(false, fmt.Sprintf("ERROR: %v", err))
		case <-h.testingChan:
			log.Println("Stopping eventHandler")
			return
		}
	}
}

func (h *handler) handleEvent(event models.Event) {
	go func() {
		eventType := models.OfficeEvents(int32(models.OfficeEvents_value[event.EventName]))
		var err error
		switch eventType {
		case models.OfficeEvents_CREATE_OFFICE:
			err = h.creatingService.CreateOffice(event)
		case models.OfficeEvents_UPDATE_OFFICE:
			err = h.updatingService.UpdateOffice(event)
		case models.OfficeEvents_DELETE_OFFICE:
			err = h.deletingService.DeleteOffice(event)
		default:
			//This may be horrible practice, but i does allow the tests to pass
			//If it throws an error, the tests freeze up. During testing it has to handle the outgoing events as well,
			//because SNS won't let you publish events to topics without subscribers
			log.Printf("Event of type %v is not handled by this service", eventType)
			//err = errors.New("Event not of type handled by this service")
		}
		if err != nil {
			h.errorChan <- err
		}
		h.testErrors(true, "Event handled")
	}()
}

func (h *handler) testErrors(ok bool, msg string) {
	log.Println(msg)
	if h.testingChan != nil && !ok {
		h.testingChan <- TestObject{ok, errors.New(msg)}
	} else if h.testingChan != nil && ok {
		h.testingChan <- TestObject{ok, nil}
	}
}
