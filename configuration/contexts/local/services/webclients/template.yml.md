# SPA Services Schema template

The SPA services is use to setup any webclient using a SPA webpage, this is very similar to the backend but it does not contain any components, there is also some extra properties that are only needed by the webclients.  
Another difference is that while the backend uses a unified endpoint for all the running services, the spa application will need a subdomain per application so we can route the traffic more effectively

```yaml
# All spa service definitions needs to start with this, the reason is locally is folder agnostic
# so while we places all the spa services definitions in the same folder it does necessarily
# needs to be so, adding that allows locally to set the configuration in the right place
spaServices:
  # A spa service is an array, you can define multiple spa services per file and they just
  # need to have unique names, they also have a naming rule, it needs to be alphanumeric and can 
  # contain dashes and underscores but it cannot contain spaces, for example example-definition
  - name: webclient-shell
    # this is normally set to false and only true to the webclient.shell but it indicates locally
    # what service should be the main ui, this can also be completely ignored
    default: false
    # location is used to define your repo location and this will override any other locations
    # folder this is normally filled in either automatically by locally if you use the repository
    # or the dockerRegistry or it will need to have the rootFolder setup on where the source
    # code is
    location:
      # This is the root folder to your repo, if this is setup then it will override all the
      # next settings
      rootFolder: C:\Code\WebClient
      # this is the folder where you keep the docker-compose file for this backend service
      # on single repos this does not need to be setup, but if you have multiple docker-compose
      # files in this repo then you can fine tune it by adding the folder where it is, this is
      # also relative to the rootFolder
      path: \src\docker-compose
      # distPath is used if you want to run the SPA from within the locally container, and this will
      # signal locally where is the compiled code for the SPA so it can copy it when generating the
      # container image, this might only be used by the owning teams and was mostly used before
      # the UI were containerized
      distPath: C:\Code\WebClient\dist
    # This is used if you want the code to be cloned for you, so if you already have your repo cloned 
    # this can be left disabled, if not you can use this to automatically clone the repo for you
    repository:
      # This will enable the auto cloning of the repo, if set to false it will not attempt to do that
      enabled: false
      # this is the repo url for the cloning tool
      url: https://github.com/org/example.git
      # this is where you want your repo to be cloned to, you can use the inbuilt source folder by adding
      # the ${{ config.path.sources }} or just the folder
      destination: ${{ config.path.sources }}/example
      # the credentials that git will use to clone the repo, you can either use a user/password, token
      # or ssh key methods, if more than one method is defined then locally will use the ssh
      # the possible combinations will be:
      # for user/pass you will need to fill in the:
      #   username:
      #   password:
      # for personal token access you need to fill in the:
      #   accessToken:
      # for the ssh you need to fill in the:
      #   privateKeyPath:
      # on all values you can use the environment variable replacer to fill it from the global section
      # like this, ${{ global.some_variable }}
      credentials:
        # github username to use for authentication
        username: 'example-user'
        # github password to use for authentication
        password: 'some_pass'
        # github access token with read/write access to the repo in question
        accessToken: 'abc'
        # location for the ssh private key to present to github for the cloning process
        privateKeyPath: 'some_path'
    # dockerRegistry definition is used to allow the consumption of a pre built container in your pc
    # a prime example will be other teams components that you need as requirement, for example loginapp
    # for this to work it will be mandatory to have the dockerCompose object as locally uses docker-compose
    # commands to start and stop containers
    dockerRegistry:
      # enables the dockerRegistry for this component
      enabled: true
      # url for the docker registry where the container is
      registry: ${{ global.docker_registry }}
      # base path for the container manifest, as an example can be /ci, this can also be left empty
      basePath: ${{ global.docker_base_manifest_path }}
      # the credentials object is used to authenticate in that docker registry and it is based in a 
      # user/password type, this might need to be requested from Ops or SRE's if it is CI/DevProd
      # as per other examples this can be replaced by a variable
      credentials:
        # Username used to login
        username: ${{ global.docker_username }}
        # password used to login
        password: ${{ global.docker_password }}
    # dockerCompose definition is used to instruct locally to generate a basic docker compose file if one
    # does not exist in the repo or if we are using the dockerRegistry definition, while we could
    # define almost all objects in this definition it is normally used for simpler cases and it should
    # not be used to entirely replace the use of the file in the repo
    # you can also use in almost all objects the environment variable replacer as in the example below
    dockerCompose:
      services:
        webclient-manifest:
          volumes:
            - ${{ config.config_service.data.path }}:/app/config-service
          ports:
            - 5540:5000
    #  uri will be the spa subdomain where it will be listening, each spa should have their own
    # subdomains with the sole exception of the webclient.shell where this will be just a dot. 
    # this uri will not be the full url but just a sub domain, for example if the uri would be example
    # then the spa service would be listening on example.locally.team
    uri: 'example'
    # reverseProxyUri is used by locally when it starts its proxy mode and this is used to move the traffic
    # to the container where the SPA is running, as with the backendServices you can use the same to
    # debug and you just need to start your debugger and attach to the same port as the reverseProxyUri
    # is listening to, and if the container is stopped then it will forward the traffic there.
    reverseProxyUri: 'host.docker.internal:5610'
    # useReverseProxy was used when the UI were not containerized and would allow to either run locally to
    # build the code or the container depending on the flag status, this should now always be true
    useReverseProxy: true
    # You can define the environment variables that will be passed down to the container, the same way
    # you define them in the helm charts, these can be defined as on the below example, you can also
    # use the global replacement to abstract users from changing the values
    environmentVariables:
      ENVIRONMENT_NAME: local
      HSTS_HEADER: max-age=2592000; includeSubDomains
      LEGACY_WEBAPP_URL: ${{ config.context.baseUrl }}
    # this is used for grouping commands with and allow to execute the same action in multiple services
    # for example, imagine you want to start all the services necessary for the UI, these can be a lot
    # or a few so instead of running the same command on each and every one of the services you can then
    # use one of the tags like ui and locally will the apply the command in each and every service that
    # contains that tag.
    # you can have as many tags as you want but they follow the same rule as before, they need to be an
    # alphanumeric word and can contain dashes and underscores
    tags:
      - ui
```
