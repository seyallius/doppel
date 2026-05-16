package object

import (
	"complex/object/auth"

	"github.com/go-webauthn/webauthn/webauthn"
)

type User struct {
	Owner                string
	Name                 string
	CreatedTime          string
	UpdatedTime          string
	DeletedTime          string
	Id                   string
	ExternalId           string
	Type                 string
	Password             string
	PasswordSalt         string
	PasswordType         string
	DisplayName          string
	FirstName            string
	LastName             string
	Avatar               string
	AvatarType           string
	PermanentAvatar      string
	Email                string
	EmailVerified        bool
	Phone                string
	CountryCode          string
	Region               string
	Location             string
	Address              []string
	Addresses            []*Address
	Affiliation          string
	Title                string
	IdCardType           string
	IdCard               string
	RealName             string
	IsVerified           bool
	Homepage             string
	Bio                  string
	Tag                  string
	Language             string
	Gender               string
	Birthday             string
	BirthdayJalali       string
	Education            string
	Score                int
	Karma                int
	Ranking              int
	Balance              float64
	BalanceCredit        float64
	Currency             string
	BalanceCurrency      string
	IsDefaultAvatar      bool
	IsOnline             bool
	IsAdmin              bool
	IsForbidden          bool
	IsDeleted            bool
	SignupApplication    string
	Hash                 string
	PreHash              string
	RegisterType         string
	RegisterSource       string
	AccessToken          string
	OriginalToken        string
	OriginalRefreshToken string

	CreatedIp      string
	LastSigninTime string
	LastSigninIp   string

	GitHub          string
	Google          string
	QQ              string
	WeChat          string
	Facebook        string
	DingTalk        string
	Weibo           string
	Gitee           string
	LinkedIn        string
	Wecom           string
	Lark            string
	Gitlab          string
	Adfs            string
	Baidu           string
	Alipay          string
	Casdoor         string
	Infoflow        string
	Apple           string
	AzureAD         string
	AzureADB2c      string
	Slack           string
	Steam           string
	Bilibili        string
	Okta            string
	Douyin          string
	Kwai            string
	Line            string
	Amazon          string
	Auth0           string
	BattleNet       string
	Bitbucket       string
	Box             string
	CloudFoundry    string
	Dailymotion     string
	Deezer          string
	DigitalOcean    string
	Discord         string
	Dropbox         string
	EveOnline       string
	Fitbit          string
	Gitea           string
	Heroku          string
	InfluxCloud     string
	Instagram       string
	Intercom        string
	Kakao           string
	Lastfm          string
	Mailru          string
	Meetup          string
	MicrosoftOnline string
	Naver           string
	Nextcloud       string
	OneDrive        string
	Oura            string
	Patreon         string
	Paypal          string
	SalesForce      string
	Shopify         string
	Soundcloud      string
	Spotify         string
	Strava          string
	Stripe          string
	Telegram        string
	TikTok          string
	Tumblr          string
	Twitch          string
	Twitter         string
	Typetalk        string
	Uber            string
	VK              string
	Wepay           string
	Xero            string
	Yahoo           string
	Yammer          string
	Yandex          string
	Zoom            string
	MetaMask        string
	Web3Onboard     string
	Custom          string
	Custom2         string
	Custom3         string
	Custom4         string
	Custom5         string
	Custom6         string
	Custom7         string
	Custom8         string
	Custom9         string
	Custom10        string

	WebauthnCredentials []webauthn.Credential
	PreferredMfaType    string
	RecoveryCodes       []string
	TotpSecret          string
	MfaPhoneEnabled     bool
	MfaEmailEnabled     bool
	MfaMessengerEnabled bool
	MfaRadiusEnabled    bool
	MfaRadiusUsername   string
	MfaRadiusProvider   string
	MfaPushEnabled      bool
	MfaPushReceiver     string
	MfaPushProvider     string
	MultiFactorAuths    []*auth.MfaProps
	Invitation          string
	InvitationCode      string
	FaceIds             []*FaceId
	Cart                []ProductInfo

	Ldap       string
	Properties map[string]string

	Roles       []*auth.Role
	Permissions []*auth.Permission
	Groups      []string

	LastChangePasswordTime string
	LastSigninWrongTime    string
	SigninWrongTimes       int

	ManagedAccounts     []ManagedAccount
	MfaAccounts         []auth.MfaAccount
	MfaItems            []*auth.MfaItem
	MfaRememberDeadline string
	NeedUpdatePassword  bool
	ApplicationScopes   []ConsentRecord
	IpWhitelist         string

	DataIntegrityHash                      string
	CompletedStepUserCheck                 StepStatus
	CompletedStepTermsOfUse                StepStatus
	CompletedStepEhrazSabtAhval            StepStatus
	CompletedStepEhrazShahkar              StepStatus
	CompletedStepPhoneNumberCheck          StepStatus
	CompletedStepOTP                       StepStatus
	CompletedStepSetPassword               StepStatus
	CompletedStepEmployeeCheck             StepStatus
	CompletedStepVerifyLdapCredentials     StepStatus
	CompletedStepForceChangePassword       StepStatus
	CompletedStepFillAdditionalAccountItem StepStatus
	CompletedStepVerifyPassword            StepStatus
	CompletedStepForeignerDocTypeSelection StepStatus
	CompletedStepCollectForeignerInfo      StepStatus
	CompletedStepSetUsername               StepStatus
	CompletedCustomSteps                   map[StepName]StepStatus

	Identifier           string
	LdapUsername         string
	NationalId           string
	SocialSecurityNumber string
	Metadata             map[string]any
	IsEmployee           bool
	IsForeigner          bool

	IsPartial           bool
	IsRegistering       bool
	CompletedSmartLogin bool

	PasswordHistory    *auth.PasswordHistory
	NeedsResetPassword bool
	DisableMfa         bool
	EkycProperties     *EkycProperties
}

type EkycProperties struct {
	IsVerify bool
	ImageUrl string
	VideoUrl string
}
type Userinfo struct {
	Sub           string
	Iss           string
	Aud           string
	Name          string
	DisplayName   string
	Email         string
	EmailVerified bool
	Avatar        string
	Address       string
	Phone         string
	RealName      string
	IsVerified    bool
	Groups        []string
	Roles         []string
	Permissions   []string
}

type ManagedAccount struct {
	Application string
	Username    string
	Password    string
	SigninUrl   string
}

type Address struct {
	Tag         string
	SubLocality string
	Line1       string
	Line2       string
	City        string
	State       string
	ZipCode     string
	Region      string
}

type FaceId struct {
	Name       string
	FaceIdData []float64
	ImageUrl   string
}

type ConsentRecord struct {
	Application   string
	GrantedScopes []string
}
