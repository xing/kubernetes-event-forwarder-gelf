package src

import (
	"reflect"
	"time"

	"github.com/golang/glog"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/xing/event-forwarder-gelf/src/util"
)

// Controller listens for events and writes them as gelf to graylog
type Controller struct {
	Stop chan struct{}

	cluster        string
	eventAddedCh   chan *core.Event
	eventUpdatedCh chan *eventUpdateGroup
	host           string
	k8sFactory     informers.SharedInformerFactory
	stopCh         chan struct{}
	writer         gelf.Writer
}

type eventUpdateGroup struct {
	oldEvent *core.Event
	newEvent *core.Event
}

// NewController instanciates a new class of Controller
func NewController(writer gelf.Writer, cluster string) *Controller {
	k8sClient := util.Clientset()
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, time.Hour*24)
	host, _ := util.GetFQDN()

	controller := &Controller{
		cluster:        cluster,
		eventAddedCh:   make(chan *core.Event),
		eventUpdatedCh: make(chan *eventUpdateGroup),
		host:           host,
		k8sFactory:     k8sFactory,
		Stop:           make(chan struct{}),
		stopCh:         make(chan struct{}),
		writer:         writer,
	}

	eventsInformer := informers.SharedInformerFactory(k8sFactory).Core().V1().Events().Informer()
	eventsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event := obj.(*core.Event)
			glog.V(2).Infof("got event %s/%s", event.ObjectMeta.Namespace, event.ObjectMeta.Name)
			controller.eventAddedCh <- event
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.eventUpdatedCh <- &eventUpdateGroup{
				oldEvent: oldObj.(*core.Event),
				newEvent: newObj.(*core.Event),
			}
		},
	})

	return controller
}

// Run starts the loop
func (c *Controller) Run() {
	c.k8sFactory.Start(c.stopCh)
	c.updateLoop()
}

func (c *Controller) updateLoop() {
	for {
		select {
		case event := <-c.eventAddedCh:
			if isLoggable(event) {
				c.log(event)
			}
		case eventUpdate := <-c.eventUpdatedCh:
			c.evaluateEventUpdate(eventUpdate)
		case stop := <-c.Stop:
			c.stopCh <- stop
			c.writer.Close()
			return
		}
	}
}

func (c *Controller) evaluateEventUpdate(g *eventUpdateGroup) {
	if !isLoggable(g.newEvent) {
		return
	}
	if g.oldEvent == nil {
		return
	}
	if reflect.DeepEqual(g.oldEvent, g.newEvent) {
		return
	}
	if g.oldEvent.Count < g.newEvent.Count {
		glog.V(3).Infof("observed updated event %s/%s (count: %d)", g.newEvent.ObjectMeta.Namespace, g.newEvent.ObjectMeta.Name, g.newEvent.Count)
		c.log(g.newEvent)
	}
}

func (c *Controller) log(event *core.Event) {
	gelfMessage := &gelf.Message{
		Version:  "0.1",
		Level:    mapEventTypeToGelfLevel(event),
		Host:     c.host,
		Short:    event.Message,
		TimeUnix: float64(event.LastTimestamp.Unix()),
		Extra: map[string]interface{}{
			"cluster":        c.cluster,
			"component":      event.Source.Component,
			"event_name":     event.ObjectMeta.Name,
			"host_name":      event.Source.Host,
			"kind":           event.InvolvedObject.Kind,
			"namespace_name": event.InvolvedObject.Namespace,
			"pod_name":       event.InvolvedObject.Name,
			"event_type":     event.Type,
			"event_reason":   event.Reason,
		},
	}

	glog.V(1).Infof("Send message to graylog")
	glog.V(2).Infof("%+v\n\n", gelfMessage)
	err := c.writer.WriteMessage(gelfMessage)
	if err != nil {
		c.stopWithError(err)
	}
}

func (c *Controller) stopWithError(err error) {
	glog.Error(err)
	go func() { c.Stop <- struct{}{} }()
}

func isLoggable(event *core.Event) bool {
	// Throw away events that are older than 5 seconds. Probably duplicates due to a restart.
	if !event.CreationTimestamp.Add(5 * time.Second).After(time.Now()) {
		return false
	}

	return true
}

func mapEventTypeToGelfLevel(event *core.Event) int32 {
	switch event.Type {
	case "Normal":
		return gelf.LOG_INFO
	case "Warning":
		return gelf.LOG_WARNING
	default:
		return gelf.LOG_ERR
	}
}
