package service

import (
	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"golang.org/x/net/context"
)

// Sports defines the interface for our sports service.
type Sports interface {
	// ListEvents will return a collection of events.
	ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error)
	// GetEvent will return a single event based on the provided filter.
	GetEvent(ctx context.Context, in *sports.GetEventRequest) (*sports.GetEventResponse, error)
}

// sportsService implements the Sports interface.
type sportsService struct {
	sportsRepo db.SportsRepo
}

// NewSportsService instantiates and returns a new sportsService.
func NewSportsService(sportsRepo db.SportsRepo) Sports {
	return &sportsService{sportsRepo}
}

// ListEvents will return a collection of events.
func (s *sportsService) ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	sportsList, err := s.sportsRepo.List(in.Filter)
	if err != nil {
		return nil, err
	}

	return &sports.ListEventsResponse{Events: sportsList}, nil
}

// GetEvent will return a single event based on the provided filter.
func (s *sportsService) GetEvent(ctx context.Context, in *sports.GetEventRequest) (*sports.GetEventResponse, error) {
	sport, err := s.sportsRepo.GetEvent(in.Filter)
	if err != nil {
		return nil, err
	}

	return &sports.GetEventResponse{Event: sport}, nil
}
