package request

type Credentials struct {
	UserName string `json:"user_name,omitempty"`
	Password string `json:"password,omitempty"`
}

type Update struct {
	UserName string `json:"user_name,omitempty"`
	Status   string `json:"status,omitempty"`
}
