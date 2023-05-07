package services

import "time"

type TimeService interface {
	Now() time.Time
}

type timeService struct {
}

func NewTimeService() TimeService {
	return &timeService{}
}

func (s *timeService) Now() time.Time {
	return time.Now()
}
