package github

type PushPayload struct {
	Ref        string      `json:"ref"`
	Before     string      `json:"before"`
	After      string      `json:"after"`
	Repository *Repository `json:"repository"`
}

type Repository struct {
	FullName string `json:"full_name"`
}
