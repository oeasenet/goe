package config

import "go.oease.dev/goe/v2/contract"

// parseWarning is a value that failed to parse as its getter's type.
type parseWarning struct {
	key, value, kind string
}

// parseFailed records that key's value does not parse as kind. The getter
// still returns its default; the warning is logged once per key and value.
func (c *config) parseFailed(key, value, kind string) {
	c.warnMu.Lock()
	if c.warned == nil {
		c.warned = map[string]bool{}
	}
	if c.warned[key+"="+value] {
		c.warnMu.Unlock()
		return
	}
	c.warned[key+"="+value] = true
	w := parseWarning{key: key, value: value, kind: kind}
	logger := c.logger
	if logger == nil {
		c.pending = append(c.pending, w)
	}
	c.warnMu.Unlock()

	if logger != nil {
		logParseWarning(logger, w)
	}
}

func logParseWarning(logger contract.Logger, w parseWarning) {
	value := w.value
	if len(value) > 64 {
		value = value[:64] + "…"
	}
	logger.Warn("Config value does not parse; using the default",
		"key", w.key,
		"value", value,
		"type", w.kind,
	)
}

// AttachLogger gives cfg a logger for values that fail to parse — a duration
// without a unit, say, which the getters turn into zero — and logs the ones
// read before. GOE calls it once its logger exists. Configurations not built by
// this package are left alone.
func AttachLogger(cfg contract.Config, logger contract.Logger) {
	switch c := cfg.(type) {
	case *config:
		c.warnMu.Lock()
		c.logger = logger
		pending := c.pending
		c.pending = nil
		c.warnMu.Unlock()
		for _, w := range pending {
			logParseWarning(logger, w)
		}
	case *ConfigWrapper:
		AttachLogger(c.base, logger)
	}
}
