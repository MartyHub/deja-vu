package dejavu

import (
	"fmt"
	"time"
)

const (
	TickInterval = 5 * time.Second
	Timeout      = 5 * time.Minute
)

type Config struct {
	clock   Clock
	db      Database
	logger  Logger
	migs    Migrations
	tick    time.Duration
	timeout time.Duration
}

func NewConfig(db Database, migs Migrations) *Config {
	return &Config{
		db:      db,
		migs:    migs,
		tick:    TickInterval,
		timeout: Timeout,
	}
}

func (c *Config) Build() DejaVu {
	cfg := *c

	if cfg.clock == nil {
		cfg.clock = NewUtcClock()
	}

	if cfg.logger == nil {
		cfg.logger = LogLogger{}
	}

	c.logger.Log(cfg.String())

	return DejaVu{Config: cfg}
}

func (c *Config) WithClock(clock Clock) *Config {
	c.clock = clock

	return c
}

func (c *Config) WithLogger(logger Logger) *Config {
	c.logger = logger

	return c
}

func (c *Config) WithTick(value time.Duration) *Config {
	c.tick = value

	return c
}

func (c *Config) WithTimeout(value time.Duration) *Config {
	c.timeout = value

	return c
}

func (c *Config) String() string {
	return fmt.Sprintf("Config: clock=%v, db=%v, migs=%v, tick=%v, timeout=%v",
		c.clock,
		c.db,
		c.migs,
		c.tick,
		c.timeout,
	)
}
