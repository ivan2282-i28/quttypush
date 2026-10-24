package providers

type FCMProvider struct {
	*BaseProvider
}

func NewFCMProvider(config map[string]interface{}) Provider {
	base := ""
	if p, ok := config["provider"]; ok {
		if s, ok := p.(string); ok {
			base = s
		}
	}
	if base == "" {
		if v, ok := config["server_key"]; ok {
			base = "https://fcm.googleapis.com"
			_ = v
		} else {
			base = "https://fcm.googleapis.com"
		}
	}
	return &FCMProvider{BaseProvider: NewBaseProvider(base)}
}
