package providers

type AppleProvider struct {
	*BaseProvider
}

func NewAppleProvider(config map[string]interface{}) Provider {
	base := ""
	if p, ok := config["provider"]; ok {
		base = p.(string)
	}
	return &AppleProvider{BaseProvider: NewBaseProvider(base)}
}
