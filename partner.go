package megaport

type PartnerService interface {
}

func NewPartnerService(c *Client) *PartnerServiceOp {
	return &PartnerServiceOp{
		Client: c,
	}
}

// PartnerServiceOp handles communication with Partner methods of the Megaport API.
type PartnerServiceOp struct {
	Client *Client
}
