package audit

type Event struct {
	timestamp int64
	params map[string]string
}

type AuditEngine struct {
	Events []Event
}

func (p *AuditEngine) Init() {
	p.Events = make([]Event, 0)
}

func (p *AuditEngine) AddEvent(event map[string]string) {
	var new_event = Event{}
	new_event.timestamp = 1000
	p.Events = append(p.Events, new_event)
}

//TODO: Make this method immutable and apply the filter
func (p *AuditEngine) ListEvents(filter map[string]string) []Event{
	return p.Events
}

var Audit AuditEngine




