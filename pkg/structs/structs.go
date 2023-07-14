package structs

type AckUpdate struct {
	QueueName  string
	Id         []byte
	ConsumerID []byte
}

type AddRoute struct {
	QueueName string
	RouteName string
}

type ChangeCredentials struct {
	Username    string
	OldPassword string
	NewPassword string
}

type Credentials struct {
	Username string
	Password string
}

type DeleteQueueInfo struct {
	QueueName string
}

type DeleteRoute struct {
	QueueName string
	RouteName string
}

type MessageCollection struct {
	Messages []*MessageData
}

type MessageData struct {
	Id   uint32
	Data []byte
}

type MessageInfo struct {
	Rk   string // RK = routingKey
	Data []byte
	Id   []byte
}

type QueueInfo struct {
	RoutingKey string
	Name       string
	Size       uint32
}

type RequestMessageData struct {
	QueueName  string
	Amount     uint32
	ConsumerID []byte
	Time       int64
}

type UserCreds struct {
	Username string
	Password string
}

type RemoveLocks struct {
	ConsumerID []byte
}

type Messages struct {
	RK   string
	Data [][]byte
	Ids  [][]byte
}

type BatchMessages struct {
	Messages []Messages
}

type UserInformation struct {
	Username string
}

type BatchAckUpdate struct {
	QueueName  string
	Ids        [][]byte
	ConsumerID []byte
}

type BatchNackUpdate struct {
	QueueName  string
	Ids        [][]byte
	ConsumerID []byte
}
