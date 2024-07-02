package contracts

type Mailer interface {
	Send() error
}
