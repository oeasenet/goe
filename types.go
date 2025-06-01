package goe

type Map map[string]any

type WebResult struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
