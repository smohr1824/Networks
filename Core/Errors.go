package Core


type NetworkArgumentError struct {
	message string
}
func NewNetworkArgumentError(message string) *NetworkArgumentError {
	return &NetworkArgumentError{
		message: message,
	}
}
func (e *NetworkArgumentError) Error() string {
	return e.message
}

//
type NetworkArgumentNullError struct {
	message string
}
func NewNetworkArgumentNullError(message string) *NetworkArgumentNullError {
	return &NetworkArgumentNullError{
		message: message,
	}
}
func (e *NetworkArgumentNullError) Error() string {
	return e.message
}
