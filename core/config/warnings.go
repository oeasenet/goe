package config

import "go.oease.dev/goe/v2/contract"

// parseWarning is a value that failed to parse as its getter's type, or, with
// file set, an env file value whose quotes do not enclose it.
type parseWarning struct {
	key, value, kind string
	file             string
}

// parseFailed records that key's value does not parse as kind. The getter
// still returns its default; the warning is logged once per key and value.
func (c *config) parseFailed(key, value, kind string) {
	c.warn(parseWarning{key: key, value: value, kind: kind})
}

// unbalancedQuotes records that key's value in file was read without its outer
// quotes because they do not enclose it. The value is not logged: values with
// stray quotes are mostly passwords.
func (c *config) unbalancedQuotes(file, key string) {
	c.warn(parseWarning{key: key, file: file})
}

// warn logs w once, or keeps it until a logger is attached.
func (c *config) warn(w parseWarning) {
	c.warnMu.Lock()
	if c.warned == nil {
		c.warned = map[parseWarning]bool{}
	}
	if c.warned[w] {
		c.warnMu.Unlock()
		return
	}
	c.warned[w] = true
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
	if w.file != "" {
		logger.Warn("Env file value has unbalanced quotes; read without its outer quotes",
			"key", w.key,
			"file", w.file,
		)
		return
	}
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
