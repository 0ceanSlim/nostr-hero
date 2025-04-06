package types

type SubscriptionFilter struct {
	IDs     []string            `json:"ids,omitempty"`
	Authors []string            `json:"authors,omitempty"`
	Kinds   []int               `json:"kinds,omitempty"`
	Tags    map[string][]string `json:"#,omitempty"`
	Since   *int64              `json:"since,omitempty"`
	Until   *int64              `json:"until,omitempty"`
	Limit   *int64              `json:"limit,omitempty"`
}
