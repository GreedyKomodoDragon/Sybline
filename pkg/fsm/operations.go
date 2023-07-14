package fsm

type Operation uint32

const (
	CREATE_QUEUE Operation = iota
	DELETE_QUEUE
	SUBMIT_MESSAGE
	ADD_ROUTING_KEY
	DELETE_ROUTING_KEY
	GET_MESSAGES
	ACK
	NACK
	CREATE_ACCOUNT
	CHANGE_PASSWORD
	REMOVE_LOCKS
	SUBMIT_BATCH_MESSAGE
	DELETE_USER
	BATCH_ACK
	BATCH_NACK
)

// ApplyResponse response from Apply raft
type ApplyResponse struct {
	Error error
	Data  interface{}
}
