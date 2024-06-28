package config

import (
	"sync"
	"time"
)

type ProtectedIDs struct {
	ids     []int64
	rwMutex sync.RWMutex
}

// scheduledWeeks is the number of weeks for which the barbershop schedule is compiled.
const ScheduledWeeks byte = 26
const NonWorkingDay time.Weekday = time.Monday
const MaxAppointmentBookingMonths byte = 1
const InfiniteWorkDate = "3000-01-01"

// location is the time zone where the barbershop is located.
var Location *time.Location

var Barbers ProtectedIDs

func init() {
	Location = time.FixedZone("MSK", 3*60*60)
}

func InitBarberIDs(ids ...int64) {
	Barbers.ids = ids
}

func (p *ProtectedIDs) AddID(idToAdd int64) {
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()
	p.ids = append(p.ids, idToAdd)
}

func (p *ProtectedIDs) IDs() []int64 {
	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()
	return p.ids
}

func (p *ProtectedIDs) RemoveID(idToRemove int64) {
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()
	for i, id := range p.ids {
		if id == idToRemove {
			p.ids = append(p.ids[:i], p.ids[i+1:]...)
			return
		}
	}
}
