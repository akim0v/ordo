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

func main() {}
