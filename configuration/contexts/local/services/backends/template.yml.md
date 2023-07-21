# Backend Services Schema template

The backend services template show you the full template and how you can set up a backend service

```yaml
# All backend service definitions needs to start with this, the reason is locally is folder agnostic
# so while we places all the backend services definitions in the same folder it does necessarily
# needs to be so, adding that allows locally to set the configuration in the right place
backendServices:
  # A backend service is an array, you can define multiple backend services per file and they just
  # need to have unique names, they also have a naming rule, it needs to be alphanumeric and can 
  # contain dashes and underscores but it cannot contain spaces, for example example-definition
  - name: example-definition
    # location is used to define your repo location and this will override any other locations
    # folder this is normally filled in either automatically by locally if you use the repository
    # or the dockerRegistry or it will need to have the rootFolder setup on where the source
    # code is
    location:
      # This is the root folder to your repo, if this is setup then it will override all the next settings
      rootFolder: C:\Code\example
      # this is the folder where you keep the docker-compose file for this backend service
      # on single repos this does not need to be setup, but if you have multiple docker-compose
      # files in this repo then you can fine tune it by adding the folder where it is, this is
      # also relative to the rootFolder
      path: \src\docker-compose
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
      # the credentials that git will use to clone the repo, you can either use a user/password, token or 
      # ssh key methods, if more than one method is defined then locally will use the ssh
      # the possible combinations will be:
      # for user/pass you will need to fill in the:
      #   username:
      #   password:
      # for personal token access you need to fill in the:
      #   accessToken:
      # for the ssh you need to fill in the:
      #   privateKeyPath:
      # on all values you can use the environment variable replacer to fill it from the global section like
      # this, ${{ global.some_variable }}
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
    # dockerCompose definition is used to instruct locally to generate a basic docker compose file if one does not
    # exist in the repo or if we are using the dockerRegistry definition, while we could define almost all
    # objects in this definition it is normally used for simpler cases and it should not be used to entirely
    # replace the use of the file in the repo
    # you can also use in almost all objects the environment variable replacer as in the example below
    dockerCompose:
      services:
        webclient-manifest:
          volumes:
            - ${{ config.config_service.data.path }}:/app/config-service
          ports:
            - 5540:5000
    # Most services do not require this parameter and it can be omitted, this is used to instead of attaching the service in question
    # to the unified subdomain, for example local-cluster.locally.team/something, to have it's own
    # subdomain. so in this case, instead of the service being in the unified subdomain it would be listening
    # in example-backend.locally.team
    uri: example-backend
    # if you have set the previous property for the uri you should then setup what are the allowed origins to
    # connect to the backend. when we do not specify the uri, this is automatically taken care by locally as it
    # knows what services are in the unified domain. to add an origin you can just add the url into the object
    # as an array of strings as we show on this example
    allowedOrigins:
      - some-frontend.locally.team
    # a backend service is normally composed of one or more components, these are effectively your containers in
    # the service, we need a definition to each one as they can have different parameters depending on their 
    # overall functionality
    components:
      # Name of the service, this is used in the command line to identify this component and has some rules,
      # it needs to be an alphanumeric word and can include dashes and underscores but cannot contain spaces
      # for example example-webhost
      - name: example-webhost
        # this is used in conjunction with the service dockerRegistry basePath to define where the docker image
        # is when generating the docker-compose.yaml file, while the basePath in the dockerRegistry is not 
        # mandatory, this one is. you can define the full path here or a partial if using the basePath, but if
        # none is present then we will not be able to download the container from the registry
        manifestPath: /webclient-manifest/webclient-manifest
        # If we are building the component from source, this is used to pass the build arguments to the docker
        # daemon, the most used is the FEED_ACCESSTOKEN as this is used to pass the credentials to download the

        # you can add as many arguments as you want and they can have two forms
        # first if you have a environment variable defined in your machine you can pass just the name of that
        # like in the example below, this will get the value from that variable
        # buildArguments:
        #   - NAME_OF_ENV_VAR
        # alternatively you can pass the value as a KVP that contain the argument and the value in it like this
        # buildArguments:
        #   - "NAME_OF_ENV_VAR=VALUE_OF_VAR"
        buildArguments:
          - FEED_ACCESSTOKEN
        # You can define the environment variables that will be passed down to the container, the same way you
        # define them in the helm charts, these can be defined as on the below example, you can also use the 
        # global replacement to abstract users from changing the values
        environmentVariables:
          ASPNETCORE_ENVIRONMENT: Development
          ASPNETCORE_URLS: http://*:5000
          identity__Ops__Issuer: ${{config.context.baseUrl}}/ops
        # reverseProxyUri is used by locally when it starts its proxy mode and allows you to have a single unified
        # sub-domain like we see in production, for example https://local-cluster.locally.team/api/config.
        # so in essence this is used so that locally proxy can route the traffic in a specific route in that
        # subdomain to a specific container in a port, this is the same principle that kubernetes uses, but
        # here in a smaller scale.
        # this also helps the debug process, imagine you have the reverse proxy for your container running on
        # port 5010, so if you would stop or pause that same container and start your visual studio project and
        # made it listen to the same port, then locally when forwarding the traffic, it will hit the project you
        # just run and allow you to get that traffic. once you done, you can start again the container and all
        # works as before.
        # the reverse proxy host will always be the same host.docker.internal the only thing different will be
        # the port, the only thing you need to be sure is that the port you use is unique
        reverseProxyUri: host.docker.internal:5540
        
        # These are the routes that locally proxy will use to forward the traffic to the container, these routes
        # will be very similar to what we use now as the istio virtual services, the only main difference is we
        # always will need the route to be a regex expression rather than a prefix.
        # you can have as many routes as you want but they will normally be the same as the ones found in your
        # virtual service. this will then be used by the locally proxy to generate the necessary configuration and
        # forward the correct traffic to this container
        routes:
          # name of the route, this is mandatory and needs to be unique. it will be used by the proxy to
          # generate configuration
          - name: root
            # the regex expression used to match the traffic in the proxy that should be forwarded to this
            # component, you can test your expression on https://regex101.com/ and select the golang section
            regex: ^\/api\/some-endpoint\/.*
            # you can use the replace in a way like the rewrite of the virtual service, while technically not
            # the same, they are used with the same intent. this allows you to capture the traffic in a route
            # and then change the path to match with what the endpoint is expecting.
            # in the background this does a string replacement and it does not deal for now with complex use
            # cases, as for the below example it simple replaces one string with another
            replace: 
              # the old string, normally the same as the regexp but can be just a word
              old: ^\/api\/some-endpoint\/.*
              # what will that word or words will be replaced with
              new: ^\/some-endpoint\/.*
            # this allows to add any specific headers to the request once the locally proxy captures it, the format
            # as before is a key value pair definition
            headers:
              - X-Forwarded-Prefix: /api/some-endpoint/
    # this is used for grouping commands with and allow to execute the same action in multiple services
    # for example, imagine you want to start all the services necessary for the UI, these can be a lot or a few
    # so instead of running the same command on each and every one of the services you can then use one of the
    # tags like ui and locally will the apply the command in each and every service that contains that tag.
    # you can have as many tags as you want but they follow the same rule as before, they need to be an
    # alphanumeric word and can contain dashes and underscores
    tags:
      - ui
```
