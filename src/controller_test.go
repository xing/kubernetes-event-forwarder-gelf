package src

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/admissionregistration"
	"k8s.io/client-go/informers/apps"
	"k8s.io/client-go/informers/autoscaling"
	"k8s.io/client-go/informers/batch"
	"k8s.io/client-go/informers/certificates"
	informers_core "k8s.io/client-go/informers/core"
	"k8s.io/client-go/informers/events"
	"k8s.io/client-go/informers/extensions"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/informers/networking"
	"k8s.io/client-go/informers/coordination"
        "k8s.io/client-go/informers/discovery"
        "k8s.io/client-go/informers/node"
	"k8s.io/client-go/informers/flowcontrol"
        "k8s.io/client-go/informers/apiserverinternal"
	"k8s.io/client-go/informers/policy"
	"k8s.io/client-go/informers/rbac"
	"k8s.io/client-go/informers/scheduling"
	"k8s.io/client-go/informers/storage"
	"k8s.io/client-go/tools/cache"
)

func TestStop(t *testing.T) {
	c := controller()

	go func() {
		time.Sleep(time.Millisecond * 100)
		c.Stop <- struct{}{}
	}()
	c.Run()

	assert.Equal(t, []string{"Close"}, dummy(c).calls)
}

func TestWriteMessage(t *testing.T) {
	c := controller()
	c.log(&core.Event{})
	assert.Equal(t, []string{"WriteMessage"}, dummy(c).calls)
}

func TestWriteMessageFromUpdate(t *testing.T) {
	c := controller()

	go func() {
		time.Sleep(time.Millisecond * 100)
		meta := metav1.ObjectMeta{CreationTimestamp: metav1.Time{Time: time.Now()}}
		c.eventUpdatedCh <- &eventUpdateGroup{
			oldEvent: &core.Event{
				ObjectMeta: meta,
				Reason:     "Foo",
				Count:      1,
			},
			newEvent: &core.Event{
				ObjectMeta:    meta,
				Reason:        "Foo",
				Count:         2,
				LastTimestamp: metav1.Time{Time: time.Now()},
			},
		}
		c.Stop <- struct{}{}
	}()
	c.Run()

	assert.Equal(t, []string{"WriteMessage", "Close"}, dummy(c).calls)
}

func TestSkipMessageFromUpdateDueToEquality(t *testing.T) {
	c := controller()

	go func() {
		time.Sleep(time.Millisecond * 100)
		meta := metav1.ObjectMeta{CreationTimestamp: metav1.Time{Time: time.Now()}}
		event := &core.Event{
			ObjectMeta:    meta,
			Reason:        "Foo",
			Count:         1,
			LastTimestamp: metav1.Time{Time: time.Now()},
		}
		c.eventUpdatedCh <- &eventUpdateGroup{
			oldEvent: event,
			newEvent: event,
		}
		c.Stop <- struct{}{}
	}()
	c.Run()

	assert.Equal(t, []string{"Close"}, dummy(c).calls)
}

func TestSkipMessageFromUpdateDueToNewness(t *testing.T) {
	c := controller()

	go func() {
		time.Sleep(time.Millisecond * 100)
		meta := metav1.ObjectMeta{CreationTimestamp: metav1.Time{Time: time.Now()}}
		event := &core.Event{
			ObjectMeta:    meta,
			Reason:        "Foo",
			Count:         1,
			LastTimestamp: metav1.Time{Time: time.Now()},
		}
		c.eventUpdatedCh <- &eventUpdateGroup{
			oldEvent: nil,
			newEvent: event,
		}
		c.Stop <- struct{}{}
	}()
	c.Run()

	assert.Equal(t, []string{"Close"}, dummy(c).calls)
}

func TestStopWithError(t *testing.T) {
	c := controller()
	dummy(c).err = errors.New("dummy")

	go func() {
		time.Sleep(time.Millisecond * 100)
		c.eventAddedCh <- &core.Event{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Time{Time: time.Now()}}}
	}()
	c.Run()

	assert.Equal(t, []string{"WriteMessage", "Close"}, dummy(c).calls)
}

func controller() *Controller {
	return &Controller{
		eventAddedCh:   make(chan *core.Event),
		eventUpdatedCh: make(chan *eventUpdateGroup),
		host:           "",
		k8sFactory:     &dummyK8sFactory{},
		Stop:           make(chan struct{}),
		stopCh:         make(chan struct{}),
		writer:         &dummyWriter{},
	}
}

func dummy(c *Controller) *dummyWriter {
	return c.writer.(*dummyWriter)
}

type dummyWriter struct {
	err   error
	calls []string
}

func (d *dummyWriter) Close() error {
	d.calls = append(d.calls, "Close")
	return d.err
}

func (d *dummyWriter) Write(b []byte) (int, error) {
	d.calls = append(d.calls, "Write")
	return len(b), d.err
}

func (d *dummyWriter) WriteMessage(m *gelf.Message) error {
	d.calls = append(d.calls, "WriteMessage")
	return d.err
}

type dummyK8sFactory struct{}

func (d *dummyK8sFactory) Start(ch <-chan struct{})                                      { go func() { <-ch }() }
func (d *dummyK8sFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool { return nil }
func (d *dummyK8sFactory) Admissionregistration() admissionregistration.Interface        { return nil }
func (d *dummyK8sFactory) Internal() apiserverinternal.Interface                         { return nil }
func (d *dummyK8sFactory) Apps() apps.Interface                                          { return nil }
func (d *dummyK8sFactory) Autoscaling() autoscaling.Interface                            { return nil }
func (d *dummyK8sFactory) Batch() batch.Interface                                        { return nil }
func (d *dummyK8sFactory) Certificates() certificates.Interface                          { return nil }
func (d *dummyK8sFactory) Coordination() coordination.Interface                          { return nil }
func (d *dummyK8sFactory) Core() informers_core.Interface                                { return nil }
func (d *dummyK8sFactory) Discovery() discovery.Interface                                { return nil }
func (d *dummyK8sFactory) Events() events.Interface                                      { return nil }
func (d *dummyK8sFactory) Extensions() extensions.Interface                              { return nil }
func (d *dummyK8sFactory) Flowcontrol() flowcontrol.Interface                            { return nil }
func (d *dummyK8sFactory) Networking() networking.Interface                              { return nil }
func (d *dummyK8sFactory) Node() node.Interface                                          { return nil }
func (d *dummyK8sFactory) Policy() policy.Interface                                      { return nil }
func (d *dummyK8sFactory) Rbac() rbac.Interface                                          { return nil }
func (d *dummyK8sFactory) Scheduling() scheduling.Interface                              { return nil }
func (d *dummyK8sFactory) Storage() storage.Interface                                    { return nil }
func (d *dummyK8sFactory) ForResource(resource schema.GroupVersionResource) (informers.GenericInformer, error) {
	return nil, nil
}
func (d *dummyK8sFactory) InformerFor(obj runtime.Object, newFunc internalinterfaces.NewInformerFunc) cache.SharedIndexInformer {
	return nil
}
