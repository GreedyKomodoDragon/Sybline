package fsm

type CommandPayload struct {
	Op   Operation   `json:"o"`
	Data interface{} `json:"d"`
}

func (c *CommandPayload) Reset() {
	c.Op = 0
	c.Data = []byte{}
}
