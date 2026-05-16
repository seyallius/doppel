package auth

type Role struct {
	Owner       string
	Name        string
	CreatedTime string
	DisplayName string
	Description string

	Users     []string
	Groups    []string
	Roles     []string
	Domains   []string
	IsEnabled bool `json:"isEnabled"`
}

type Permission struct {
	Owner       string
	Name        string
	CreatedTime string
	DisplayName string
	Description string

	Users   []string
	Groups  []string
	Roles   []string
	Domains []string

	Model        string
	Adapter      string
	ResourceType string
	Resources    []string
	Actions      []string
	Effect       string
	IsEnabled    bool `json:"isEnabled"`

	Submitter   string
	Approver    string
	ApproveTime string
	State       string
}
