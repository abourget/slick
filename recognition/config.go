package recognition

type Config struct {
	DomainRestriction string `json:"domain_restriction"` // Only accept up votes from people with emails ending with this value.
	Channel           string `json:"channel"`            // Name of the channel where recognitions will be shouted to.
}
