package azure_cli

type AzureAdApps []AzureAdApp

type AzureAdApp struct {
	AddIns                            []interface{}            `json:"addIns"`
	API                               API                      `json:"api"`
	AppID                             string                   `json:"appId"`
	AppRoles                          []AppRole                `json:"appRoles"`
	ApplicationTemplateID             *string                  `json:"applicationTemplateId"`
	Certification                     interface{}              `json:"certification"`
	CreatedDateTime                   string                   `json:"createdDateTime"`
	DefaultRedirectURI                *string                  `json:"defaultRedirectUri"`
	DeletedDateTime                   interface{}              `json:"deletedDateTime"`
	Description                       interface{}              `json:"description"`
	DisabledByMicrosoftStatus         interface{}              `json:"disabledByMicrosoftStatus"`
	DisplayName                       string                   `json:"displayName"`
	GroupMembershipClaims             *string                  `json:"groupMembershipClaims"`
	ID                                string                   `json:"id"`
	IdentifierUris                    []string                 `json:"identifierUris"`
	Info                              Info                     `json:"info"`
	IsDeviceOnlyAuthSupported         *bool                    `json:"isDeviceOnlyAuthSupported"`
	IsFallbackPublicClient            *bool                    `json:"isFallbackPublicClient"`
	KeyCredentials                    []KeyCredential          `json:"keyCredentials"`
	Notes                             *string                  `json:"notes"`
	OptionalClaims                    *OptionalClaims          `json:"optionalClaims"`
	ParentalControlSettings           ParentalControlSettings  `json:"parentalControlSettings"`
	PasswordCredentials               []PasswordCredential     `json:"passwordCredentials"`
	PublicClient                      PublicClient             `json:"publicClient"`
	PublisherDomain                   *PublisherDomain         `json:"publisherDomain"`
	RequestSignatureVerification      interface{}              `json:"requestSignatureVerification"`
	RequiredResourceAccess            []RequiredResourceAccess `json:"requiredResourceAccess"`
	SamlMetadataURL                   interface{}              `json:"samlMetadataUrl"`
	ServiceManagementReference        interface{}              `json:"serviceManagementReference"`
	ServicePrincipalLockConfiguration interface{}              `json:"servicePrincipalLockConfiguration"`
	SignInAudience                    SignInAudience           `json:"signInAudience"`
	SPA                               PublicClient             `json:"spa"`
	Tags                              []interface{}            `json:"tags"`
	TokenEncryptionKeyID              interface{}              `json:"tokenEncryptionKeyId"`
	VerifiedPublisher                 VerifiedPublisher        `json:"verifiedPublisher"`
	Web                               Web                      `json:"web"`
}

type API struct {
	AcceptMappedClaims          interface{}             `json:"acceptMappedClaims"`
	KnownClientApplications     []interface{}           `json:"knownClientApplications"`
	Oauth2PermissionScopes      []Oauth2PermissionScope `json:"oauth2PermissionScopes"`
	PreAuthorizedApplications   []interface{}           `json:"preAuthorizedApplications"`
	RequestedAccessTokenVersion *int64                  `json:"requestedAccessTokenVersion"`
}

type Oauth2PermissionScope struct {
	AdminConsentDescription string      `json:"adminConsentDescription"`
	AdminConsentDisplayName string      `json:"adminConsentDisplayName"`
	ID                      string      `json:"id"`
	IsEnabled               bool        `json:"isEnabled"`
	Type                    TypeElement `json:"type"`
	UserConsentDescription  string      `json:"userConsentDescription"`
	UserConsentDisplayName  string      `json:"userConsentDisplayName"`
	Value                   Value       `json:"value"`
}

type AppRole struct {
	AllowedMemberTypes []TypeElement `json:"allowedMemberTypes"`
	Description        string        `json:"description"`
	DisplayName        string        `json:"displayName"`
	ID                 string        `json:"id"`
	IsEnabled          bool          `json:"isEnabled"`
	Origin             string        `json:"origin"`
	Value              interface{}   `json:"value"`
}

type Info struct {
	LogoURL             *string     `json:"logoUrl"`
	MarketingURL        interface{} `json:"marketingUrl"`
	PrivacyStatementURL interface{} `json:"privacyStatementUrl"`
	SupportURL          interface{} `json:"supportUrl"`
	TermsOfServiceURL   interface{} `json:"termsOfServiceUrl"`
}

type KeyCredential struct {
	CustomKeyIdentifier string      `json:"customKeyIdentifier"`
	DisplayName         string      `json:"displayName"`
	EndDateTime         string      `json:"endDateTime"`
	Key                 interface{} `json:"key"`
	KeyID               string      `json:"keyId"`
	StartDateTime       string      `json:"startDateTime"`
	Type                string      `json:"type"`
	Usage               string      `json:"usage"`
}

type OptionalClaims struct {
	AccessToken []interface{} `json:"accessToken"`
	IDToken     []interface{} `json:"idToken"`
	Saml2Token  []Saml2Token  `json:"saml2Token"`
}

type Saml2Token struct {
	AdditionalProperties []interface{} `json:"additionalProperties"`
	Essential            bool          `json:"essential"`
	Name                 string        `json:"name"`
	Source               interface{}   `json:"source"`
}

type ParentalControlSettings struct {
	CountriesBlockedForMinors []interface{}     `json:"countriesBlockedForMinors"`
	LegalAgeGroupRule         LegalAgeGroupRule `json:"legalAgeGroupRule"`
}

type PasswordCredential struct {
	CustomKeyIdentifier *CustomKeyIdentifier `json:"customKeyIdentifier"`
	DisplayName         *string              `json:"displayName"`
	EndDateTime         string               `json:"endDateTime"`
	Hint                *string              `json:"hint"`
	KeyID               string               `json:"keyId"`
	SecretText          interface{}          `json:"secretText"`
	StartDateTime       string               `json:"startDateTime"`
}

type PublicClient struct {
	RedirectUris []interface{} `json:"redirectUris"`
}

type RequiredResourceAccess struct {
	ResourceAccess []ResourceAccess `json:"resourceAccess"`
	ResourceAppID  string           `json:"resourceAppId"`
}

type ResourceAccess struct {
	ID   string             `json:"id"`
	Type ResourceAccessType `json:"type"`
}

type VerifiedPublisher struct {
	AddedDateTime       interface{} `json:"addedDateTime"`
	DisplayName         interface{} `json:"displayName"`
	VerifiedPublisherID interface{} `json:"verifiedPublisherId"`
}

type Web struct {
	HomePageURL           *string               `json:"homePageUrl"`
	ImplicitGrantSettings ImplicitGrantSettings `json:"implicitGrantSettings"`
	LogoutURL             *string               `json:"logoutUrl"`
	RedirectURISettings   []RedirectURISetting  `json:"redirectUriSettings"`
	RedirectUris          []string              `json:"redirectUris"`
}

type ImplicitGrantSettings struct {
	EnableAccessTokenIssuance bool `json:"enableAccessTokenIssuance"`
	EnableIDTokenIssuance     bool `json:"enableIdTokenIssuance"`
}

type RedirectURISetting struct {
	Index *int64 `json:"index"`
	URI   string `json:"uri"`
}

type TypeElement string

const (
	User TypeElement = "User"
)

type Value string

const (
	AccessAsUser      Value = "access_as_user"
	ReadAccess        Value = "ReadAccess"
	UserImpersonation Value = "user_impersonation"
)

type LegalAgeGroupRule string

const (
	Allow LegalAgeGroupRule = "Allow"
)

type CustomKeyIdentifier string

const (
	The5YAGIAYQBjAA     CustomKeyIdentifier = "//5yAGIAYQBjAA=="
	VABlAHMAdABLAGUAEQA CustomKeyIdentifier = "VABlAHMAdABLAGUAeQA="
)

type PublisherDomain string

const (
	LandeskincOnmicrosoftCOM PublisherDomain = "landeskinc.onmicrosoft.com"
)

type ResourceAccessType string

const (
	Role  ResourceAccessType = "Role"
	Scope ResourceAccessType = "Scope"
)

type SignInAudience string

const (
	AzureADMyOrg                       SignInAudience = "AzureADMyOrg"
	AzureADandPersonalMicrosoftAccount SignInAudience = "AzureADandPersonalMicrosoftAccount"
)

type AzureAdAppList []string
