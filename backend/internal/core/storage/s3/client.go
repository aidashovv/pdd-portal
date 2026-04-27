package s3

type Client struct {
	Config Config
}

func NewClient(config Config) *Client {
	return &Client{Config: config}
}
