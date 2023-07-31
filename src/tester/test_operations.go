package tester

import (
	"fmt"

	"github.com/cjlapao/locally-cli/azure_cli"
	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/lanes"
	"github.com/cjlapao/locally-cli/notifications"
	"github.com/cjlapao/locally-cli/vaults/azure_keyvault"

	"github.com/cjlapao/common-go/helper"
)

var notify = notifications.Get()

func TestOperations(subCommand string) {
	automationService := lanes.Get()
	switch subCommand {
	case "azure_acr":
		action := common.VerifyCommand(helper.GetArgumentAt(2))
		switch action {
		case "token_exchange":
			tool := azure_cli.Get()
			name := helper.GetFlagValue("name", "")
			subscriptionId := helper.GetFlagValue("sid", "")
			tenantId := helper.GetFlagValue("tid", "")
			token, err := tool.ExchangeRefreshTokenForAccessToken(name, "", subscriptionId, tenantId)
			if err != nil {
				notify.FromError(err, "getting token")
			}
			notify.Info(token)
		}
	case "pipelines":
		action := common.VerifyCommand(helper.GetArgumentAt(2))
		pipeline := common.VerifyCommand(helper.GetArgumentAt(3))
		switch action {
		case "run":
			if err := automationService.Validate(pipeline); err == nil {
				if err := automationService.Run(pipeline); err != nil {
					notify.Error("There was an error executing the requested pipeline %s", pipeline)
				}
			} else {
				notify.Error("There was an error validating the requested pipeline %s", pipeline)
			}
		case "validate":
			if err := automationService.Validate(pipeline); err != nil {
				notify.Error("There was an error validating the requested pipeline %s", pipeline)
			}
		}
	case "env":
		action := common.VerifyCommand(helper.GetArgumentAt(2))
		switch action {
		case "config":
			notify.Info("Testing the env system")
			env := environment.Get()

			env.Add("test", "carlos", "mad")
			env.Add("test", "carlos1", "good")
			env.Add("terraform", "carlos", "amazing")
			env.Add("simple", "carlos", "testing")
			env.Add("complex", "carlos.amazing.replacer", "testing complex")

			env.Remove("test", "carlos1")
			r := env.Get("test", "carlos")
			x := env.Replace("${{ simple.carlos }}")
			y := env.Replace("${{complex.carlos.amazing.replacer }}")
			z := env.Replace("${{ simple.carlos1 }}")
			config := env.Replace("${{ config.path.base }}")
			notify.Success("%s", r)
			notify.Success("%s", x)
			notify.Success("%s", y)
			notify.Success("%s", z)
			notify.Success("With replace: %s", config)
			notify.Success("Direct: %s", env.Get("config", "path.base"))
			notify.Success("Direct: %s", env.Get("config", "path.sources"))
		case "keyvault":
			env := environment.Get()
			env.Register(azure_keyvault.New("global", &azure_keyvault.AzureKeyVaultOptions{
				KeyVaultUri:  "https://cjlocal-global-kv.vault.azure.net/",
				DecodeBase64: true,
			}))
			// env.SyncVault("keyvault")
		}
	case "az":
		action := common.VerifyCommand(helper.GetArgumentAt(2))
		switch action {
		case "login":
			wrapper := azure_cli.GetWrapper()
			credentials := azure_cli.WrapperCredentials{
				ServicePrincipal: helper.GetFlagSwitch("sp", false),
				SubscriptionId:   helper.GetFlagValue("sid", ""),
				UseDeviceCode:    helper.GetFlagSwitch("device-code", false),
				Username:         helper.GetFlagValue("user", ""),
				Password:         helper.GetFlagValue("pass", ""),
				TenantId:         helper.GetFlagValue("tid", ""),
			}

			if err := wrapper.Login(&credentials); err != nil {
				notify.Error(err.Error())
			}
			if helper.GetFlagSwitch("double", false) {
				if err := wrapper.Login(&credentials); err != nil {
					notify.Error(err.Error())
				}
			}
		case "logout":
			wrapper := azure_cli.GetWrapper()
			credentials := azure_cli.WrapperCredentials{
				ServicePrincipal: helper.GetFlagSwitch("sp", false),
				SubscriptionId:   helper.GetFlagValue("sid", ""),
				UseDeviceCode:    helper.GetFlagSwitch("device-code", false),
				Username:         helper.GetFlagValue("user", ""),
				Password:         helper.GetFlagValue("pass", ""),
				TenantId:         helper.GetFlagValue("tid", ""),
			}

			if err := wrapper.Login(&credentials); err != nil {
				notify.Error(err.Error())
			}

			if err := wrapper.Logout(); err != nil {
				notify.Error(err.Error())
			}
			if helper.GetFlagSwitch("double", false) {
				if err := wrapper.Logout(); err != nil {
					notify.Error(err.Error())
				}
			}
		case "list":
			wrapper := azure_cli.GetWrapper()
			// credentials := azure_cli.WrapperCredentials{
			// 	ServicePrincipal: helper.GetFlagSwitch("sp", false),
			// 	SubscriptionId:   helper.GetFlagValue("sid", ""),
			// 	UseDeviceCode:    helper.GetFlagSwitch("device-code", false),
			// 	Username:         helper.GetFlagValue("user", ""),
			// 	Password:         helper.GetFlagValue("pass", ""),
			// 	TenantId:         helper.GetFlagValue("tid", ""),
			// }

			// if err := wrapper.Login(&credentials); err != nil {
			// 	notify.Error(err.Error())
			// }

			appName := helper.GetFlagValue("appName", "")
			if appList, err := wrapper.ListApps(appName); err != nil {
				notify.Error(err.Error())
			} else {
				fmt.Printf("%v", appList)
			}
		case "create-sp":
			wrapper := azure_cli.GetWrapper()
			appName := helper.GetFlagValue("appName", "")
			subscriptionId := helper.GetFlagValue("sid", "")
			if appList, err := wrapper.CreateServicePrincipal(appName, subscriptionId); err != nil {
				notify.Error(err.Error())
			} else {
				fmt.Printf("%v", appList)
			}
		case "create-resource-group":
			wrapper := azure_cli.GetWrapper()
			rgName := helper.GetFlagValue("rgName", "")
			subscriptionId := helper.GetFlagValue("sid", "")
			location := helper.GetFlagValue("location", "")

			if err := wrapper.UpsertResourceGroup(rgName, subscriptionId, location); err != nil {
				notify.Error(err.Error())
			}
		case "create-storage-account":
			wrapper := azure_cli.GetWrapper()
			name := helper.GetFlagValue("name", "")
			rgName := helper.GetFlagValue("rgName", "")
			subscriptionId := helper.GetFlagValue("sid", "")

			if err := wrapper.UpsertStorageAccount(name, rgName, subscriptionId); err != nil {
				notify.Error(err.Error())
			}
		case "get-storage-account-key":
			wrapper := azure_cli.GetWrapper()
			name := helper.GetFlagValue("name", "")
			rgName := helper.GetFlagValue("rgName", "")
			subscriptionId := helper.GetFlagValue("sid", "")

			if _, err := wrapper.GetStorageAccountKey(name, rgName, subscriptionId); err != nil {
				notify.Error(err.Error())
			}
		case "create-storage-account-container":
			wrapper := azure_cli.GetWrapper()
			name := helper.GetFlagValue("name", "")
			rgName := helper.GetFlagValue("rgName", "")
			subscriptionId := helper.GetFlagValue("sid", "")
			sa := helper.GetFlagValue("sa", "")
			var key string
			var err error

			if key, err = wrapper.GetStorageAccountKey(sa, rgName, subscriptionId); err != nil {
				notify.Error(err.Error())
			}

			if err := wrapper.UpsertStorageAccountContainer(name, sa, key); err != nil {
				notify.Error(err.Error())
			}
		case "user-login":
			wrapper := azure_cli.GetWrapper()
			subscriptionId := helper.GetFlagValue("sid", "")
			tenantId := helper.GetFlagValue("tid", "")
			if err := wrapper.UserLogin(subscriptionId, tenantId); err != nil {
				notify.Error(err.Error())
			}
		}
	}
}
