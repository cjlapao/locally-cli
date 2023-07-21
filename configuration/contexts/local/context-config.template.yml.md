# Context configuration Schema

The context configuration file contains all the necessary configuration to start a local environment and should be the only file you will need to fill in to run locally, while some values are already pre filled they can be changed if there is a need to, others will need to be filled in.

Pay close attention to the comments as they have hints on either how to source it or how to fill it in, some, specially the azure storage accounts have restrictions on how the value is.

```yaml
# this configures the context runtime and is mandatory
configuration:
  # this is your unified subdomain, the value here will be used by locally to generate the host entries and the
  # proxy rules to allow the traffic to be routed to the right places, so if you would use the local-cluster
  # as per example, the unified base url would be local-cluster.locally.team, the domains comes from
  # the locally-config.yaml and if changed there it would affect the base url
  rootUri: local-cluster
  # this folder is used by locally to dump its generated files or where it clones repos by default, this can be
  # be left empty, locally will then create a folder in the context folder  named .cache-data and use that
  # if you set a path then this will override the default and use that one instead
  # In some cases allowing locally to create this folder might bring issues with long paths, if that happens
  # you will need to override it and make it closer to your root drive 
  outputPath: ''
# this is where locally will store it's environment variables, some can be manually added like for example the
# global, but others are filled in by locally when it runs certain types of tasks.
environmentVariables:
  # this section makes variables to be available to the rest of the configuration by using the
  # ${{ global.var_name }} as the value, they can be used as complex values for example 
  #   foo: /test/${{ global.bar }}/example
  # the global section can also have variables that call others from the global section, these are called
  # nested variables., this works because locally will replace the value only on runtime, so by that time all
  # available variables will be there
  # by default we provide the required variables that are needed to start the local environment and these
  # will grow with time as more and more services start using it
  global:
  # keyvault will be filled in either by running the [locally keyvault sync name_of_keyvault] or by the
  # locally pipelines, in either case it should be left empty
  keyvault:
  # terraform object will be filled in when you run your infrastructure with all the outputs from it
  terraform:
# credentials is used by services and locally to authenticate to the different resources like for example azure
credentials:
  # this will have all the necessary credentials to use or create the azure resources
  azure:
    # add the client id if you already have one otherwise it will generate a new one for you
    # Attention: if you allow locally to create your service principal it might take a bit of time for
    # Microsoft to enable it and you might have an access denied on the first run of locally, to fix it
    # just restart the same command
    clientId: ''
    # add the client secret if you already have one otherwise it will generate a new one for you
    clientSecret: ''
    # your Azure Cloud "Visual Studio Enterprise" subscription id 
    # How to get it:
    #  1. Open Azure Portal Subscription page https://portal.azure.com/#view/Microsoft_Azure_Billing/SubscriptionsBlade
    #  2. Click on "Visual Studio Enterprise" subscription 
    #  3. Locate "Subscription ID" field and copy its value
    subscriptionId: ''
    # your Azure Active Directory tenant id 
    # How to get it:
    #  1. Open Azure Portal Azure Active Directory properties page https://portal.azure.com/#view/Microsoft_AAD_IAM/TenantPropertiesBlade
    #  2. Locate the "Tenant ID" field and copy its value
    tenantId: ''
# backendConfig is used to configure the infrastructure backend tfstate for terraform
backendConfig:
  # this will use the terraform azure backend to store the stacks
  azure: 
    # location where you will want to create the resource group to keep the states, it will be using the same
    # we are using for all the other resources
    location: ${{ global.azure_location}}
    # this will be the subscription we will want to use to create the resources on, it will be using the same
    # we have in the credentials
    subscriptionId: ${{ credentials.azure.subscription_id  }}
    # name of the resource group to keep the states, if it does not exist locally will create it, you can use
    # for example: local_stacks
    resourceGroupName: ''
    # name of the storage account to keep the states, if it does not exist locally will create it, this cannot
    # be longer than 24 chars, and the name needs to be globally unique so try to use your imagination
    storageAccountName: '' 
    # name of the container to keep the states,  if it does not exist locally will create it, example: stacks
    containerName: '' 
```
