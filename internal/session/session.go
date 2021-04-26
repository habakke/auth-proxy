package session

type Data struct {
	ID         string `json:"id,omitempty"`
	Email      string `json:"email,omitempty"`
	Name       string `json:"name,omitempty"`
	Authorised bool   `json:"authorised"`
}
