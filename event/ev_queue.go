package event

type EventChan chan EventMsg
type EventQueue struct {
	ev_queue EventChan
}

func NewEventQueue() *EventQueue {

	queue := &EventQueue{
		ev_queue: make(EventChan, 65535),
	}
	return queue
}

func (this *EventQueue) Push(event EventMsg) {
	this.ev_queue <- event
}

func (this *EventQueue) Get() EventMsg {
	event := <-this.ev_queue
	return event
}

func (this *EventQueue) GetChan() EventChan {
	return this.ev_queue
}
