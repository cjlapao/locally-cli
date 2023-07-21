# locally main configuration file

This is the tool configuration file and it is used to setup locally as a tool and add the contexts that it will be using, we provide the bare minimum setup but you can adjust your one by looking at the available schema here

the file is called by default ```locally-config.yml``` and you can copy that to ```locally-config.personal.yml``` and that will be used as default, the repo is also set to ignore the changes on that personal file

```yaml
# This will contain the context's configurations, you can have as many contexts as you want and you
# will just need to point at the context configuration, this can also be in any place in your machine
contexts:
      # this will be the name of the context used when switching from one context to another, the name
      # can contains letter, numbers, dashes and underscores
    - name: local
      # path for where the context configuration is, this will be the starting point of the rest of the configuration and it needs to point to the file and the file needs to exist
      configPath: .\context\local\context.config.yml
# We use this definition to override the environment path for the tools locally uses, this is not
# mandatory and can be left out of the config file, or you can just set individual tools. the most
# common example would be the terraform or caddy as they are downloaded as exec files rather than
# having an installer
tools:
    azurecli:
        # this allows to override the azure cli tool by setting it's path to where the tool is, this
        # needs to point to the exact location of the exe rather than the folder, see below example
        path: c:\bin\az.exe
    caddy:
        # this allows to override the caddy tool by setting it's path to where the tool is, this needs
        # to point to the exact location of the exe rather than the folder, see below example
        path: c:\bin\caddy.exe
    docker:
      # this allows to set a retry for the docker build daemon in case it fails to build the container
      # by default it's 3
      buildRetries: 5
      # this allows to override the docker tool by setting it's path to where the tool is, this needs
      # to point to the exact location of the exe rather than the folder, see below example
      dockerPath: c:\bin\docker.exe
      # this allows to override the docker compose tool by setting it's path to where the tool is,
      # this needs to point to the exact location of the exe rather than the folder, see below example
      dockerComposePath: c:\bin\docker.exe
    nuget:
        # this allows to override the nuget tool by setting it's path to where the tool is, this needs
        # to point to the exact location of the exe rather than the folder, see below example
        path: c:\bin\nuget.exe
    dotnet:
        # this allows to override the dotnet tool by setting it's path to where the tool is, this needs
        # to point to the exact location of the exe rather than the folder, see below example
        path: c:\bin\dotnet.exe
    git:
        # this allows to override the git tool by setting it's path to where the tool is, this needs
        # to point to the exact location of the exe rather than the folder, see below example
        path: c:\bin\git.exe
    terraform:
        # this allows to override the git tool by setting it's path to where the tool is, this needs
        # to point to the exact location of the exe rather than the folder, see below example
        # ATTENTION: due to a stack requirement you will need to use a specific version of terraform
        # at the moment you need to use the v0.14.7 and this can be found on
        # https://releases.hashicorp.com/terraform/0.14.7/terraform_0.14.7_windows_amd64.zip
        path: c:\bin\terraform.exe
# we use the network definition to setup some network settings that are required for the proxy to run
network:
    # this should be left intact, it will use the default loopback as the card it will attach and generate the host files for
    localIp: 127.0.0.1
    # what is the domain we will be using, this due to restrictions on linucx containers we cannot use
    # a self signed certificate and for now we will be using the same as CI is using, that means using
    # the same domain, locally.team, in future we might change this to use a different one
    domainName: locally.team
    # because we are using the locally.team domain you will need to request the certificate private key and cert file, once you have those you can put it in any folder and fill in the path below as showed in the below example
    certPath: c:\ssl\cert.crt
    privateKeyPath: c:\ssl\cert.key
# this definition will setup the global cors for the locally proxy, normally this does not require any
# changes but you can fine tune it as you like, at the moment is pretty open and should allow all
# the use cases
cors:
    allowedMethods: OPTIONS,HEAD,GET,POST,PUT,PATCH,DELETE
    allowedHeaders: '*'
    allowedOrigins:
        - localhost
```
