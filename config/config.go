package config

import "encoding/json"

// Conf is a collection of config key/value pairs
type Conf struct {
	c map[string]string
}

// New returns a conf collection from the passed argument.
func New(conf map[string]string) *Conf {
	return &Conf{
		c: conf,
	}
}

// NewFromJSON returns a conf collection, parsed from the passed JSON blob.
func NewFromJSON(data []byte) (*Conf, error) {
	var c map[string]string
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &Conf{c: c}, nil
}

// IsSet returns true if key is set.
func (c *Conf) IsSet(key string) bool {
	_, ok := c.c[key]
	return ok
}

// GetString returns the value as a string.
func (c *Conf) GetString(key string) string {
	s, _ := c.c[key]
	return s
}
