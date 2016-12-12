package audit

import "time"

type Event struct {
	Timestamp int64
	Details map[string]string
}

type AuditEngine struct {
	Events []Event
}

func (p *AuditEngine) Init() {
	p.Events = make([]Event, 0)
}

func (p *AuditEngine) AddEvent(event map[string]string) {
	p.Events = append(p.Events, Event{Timestamp:time.Now().Unix(), Details:event})
}


//TODO: Make this method immutable and apply the filter
func (p *AuditEngine) ListEvents(filter map[string]string) []Event{
	res := p.Events
	return res
}

var Audit AuditEngine




