package sentinelone

type User struct {
	Email      string `json:"email"`
	Scope      string `json:"scope"`
	ID         string `json:"id"`
	ScopeRoles []Role `json:"scopeRoles"`
	FullName   string `json:"fullName"`
}

type ServiceUser struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
	ScopeRoles  []Role `json:"scopeRoles"`
	Scope       string `json:"scope"`
}

type Account struct {
	AccountType string `json:"accountType"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsDefault   bool   `json:"isDefault"`
}

type Site struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	SiteType    string `json:"siteType"`
	AccountID   string `json:"accountId"`
}

// Combination of predefined role and scope role.
type Role struct {
	AccountName string `json:"accountName,omitempty"`
	RoleName    string `json:"roleName,omitempty"`
	RoleID      string `json:"roleId,omitempty"`
	ID          string `json:"id"`
	Name        string `json:"name"`
}
