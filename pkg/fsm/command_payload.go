package fsm

type CommandPayload struct {
	Username string    `json:"u"`
	Op       Operation `json:"o"`
	Data     []byte    `json:"d"`
}

func (c *CommandPayload) Reset() {
	c.Op = 0
	c.Data = []byte{}
	c.Username = ""
}
