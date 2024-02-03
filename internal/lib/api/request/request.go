package request

type Credentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Status struct {
	UserName string `json:"user_name,omitempty"`
	Status   string `json:"status,omitempty"`
}
