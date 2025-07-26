package azure_cli

type AcrLoginResponse struct {
	AccessToken string `json:"accessToken"`
	LoginServer string `json:"loginServer"`
}
