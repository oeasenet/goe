package msearch

type SortOrder int

const (
	ASC  SortOrder = 1
	DESC SortOrder = -1
)

type IndexAttributeItem struct {
	SearchableFields []string `json:"searchable_fields"`
	FilterableFields []string `json:"filterable_fields"`
	SortableFields   []string `json:"sortable_fields"`
	DisplayedFields  []string `json:"displayed_fields"`
}
type configDataMap map[string]*IndexAttributeItem
type IndexConfigs struct {
	ConfigData *configDataMap `json:"config_data"`
}
