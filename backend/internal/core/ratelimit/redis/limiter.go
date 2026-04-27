package redis

type Limiter struct {
	Config Config
}

func NewLimiter(config Config) *Limiter {
	return &Limiter{Config: config}
}
