package auth

type MfaProps struct {
	Enabled            bool
	IsPreferred        bool
	MfaType            string
	Secret             string
	CountryCode        string
	URL                string
	RecoveryCodes      []string
	MfaRememberInHours int
}

type MfaItem struct {
	Name string
	Rule string
}

type MfaAccount struct {
	AccountName string
	Issuer      string
	SecretKey   string
	Origin      string
}
