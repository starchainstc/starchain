package events

import (
	"sync"
	"errors"
)

type EventFunc func(v interface{})

type Subscriber chan interface{}

type Event struct {
	m           sync.RWMutex
	subscribers map[EventType]map[Subscriber]EventFunc
	//listeners	map[EventType]EventFunc
}

var event *Event
var initial bool

func NewEvent() *Event {
	return &Event{
		subscribers: make(map[EventType]map[Subscriber]EventFunc),
	}
}

//func (e *Event) AddListener(eventType EventType,eventFunc EventFunc){
//
//	e.listeners[eventType] = eventFunc
//}


func (e *Event) Subscribe(eventType EventType,eventFunc EventFunc) Subscriber{
	e.m.Lock()
	defer e.m.Unlock()
	sub := make(chan interface{})
	_,ok := e.subscribers[eventType]
	if !ok {
		e.subscribers[eventType] = make(map[Subscriber]EventFunc)
	}
	e.subscribers[eventType][sub] = eventFunc

	return sub
}

func (e *Event) UnSubscribe(eventtype EventType,subscriber Subscriber) (err error){
	e.m.Lock()
	defer e.m.Unlock()

	subEvent,ok := e.subscribers[eventtype]
	if !ok {
		err = errors.New("No event type.")
		return
	}

	delete(subEvent,subscriber)
	close(subscriber)

	return
}

//Notify subscribers that Subscribe specified event
func (e *Event) Notify(eventtype EventType,value interface{}) (err error){
	e.m.RLock()
	defer e.m.RUnlock()

	subs,ok := e.subscribers[eventtype]
	if !ok {
		err = errors.New("No event type.")
		return
	}

	for _, eventfunc := range subs {
		go e.NotifySubscriber(eventfunc,value)
	}
	return
}

func (e *Event) NotifySubscriber(eventfunc EventFunc, value interface{}) {
	if eventfunc == nil { return }

	//invode subscriber event func
	eventfunc(value)

}
