package main

// Notifier has more than one implementation.
type Notifier interface {
	Name() string
}

type emailNotifier struct{}

func NewEmailNotifier() *emailNotifier { return &emailNotifier{} }
func (*emailNotifier) Name() string    { return "email" }

type smsNotifier struct{}

func NewSMSNotifier() *smsNotifier { return &smsNotifier{} }
func (*smsNotifier) Name() string  { return "sms" }

// Dispatcher consumes every Notifier.
type Dispatcher struct {
	notifiers []Notifier
}

// NewDispatcher takes a slice of the interface.
func NewDispatcher(notifiers []Notifier) *Dispatcher {
	return &Dispatcher{notifiers: notifiers}
}

func (d *Dispatcher) Names() []string {
	names := make([]string, 0, len(d.notifiers))
	for _, n := range d.notifiers {
		names = append(names, n.Name())
	}
	return names
}

func main() {}
