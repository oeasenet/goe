package contracts

type Config interface {
	Get(string) string
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetStringSlice(key string) []string
	GetIntSlice(key string) []int
	GetBoolSlice(key string) []bool

	GetOrDefaultString(key string, defaultValue string) string
	GetOrDefaultInt(key string, defaultValue int) int
	GetOrDefaultBool(key string, defaultValue bool) bool
}
