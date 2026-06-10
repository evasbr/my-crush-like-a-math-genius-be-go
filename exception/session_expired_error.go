package exception

type SessionExpiredError struct {
	Message string
}

func (e SessionExpiredError) Error() string {
	return e.Message
}
