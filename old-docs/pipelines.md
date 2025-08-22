# locally Pipelines

- [locally Pipelines](#locally-pipelines)
  - [Purpose](#purpose)
  - [Pipeline Definition Schema](#pipeline-definition-schema)
  - [Supported Jobs Types](#supported-jobs-types)
    - [Bash Worker](#bash-worker)
    - [Curl Worker](#curl-worker)
    - [Docker Worker](#docker-worker)
    - [Dotnet Worker](#dotnet-worker)
    - [.NET Entity Framework Migrations Worker](#net-entity-framework-migrations-worker)
    - [EMS Worker](#ems-worker)
    - [Git Worker](#git-worker)
    - [Infrastructure Worker](#infrastructure-worker)
    - [KeyVault Worker](#keyvault-worker)
    - [SQL Worker](#sql-worker)

locally has a concept called pipelines, these are very similar to what ADO pipelines are or even GitHub Actions. In their basic form locally pipelines are a form of automation for running specific tasks in order.

## Purpose

Sys can be a complex system and deploying or running services can be a simple process or a very complex one, each service has a set of requirements, and those requirements can vary extremely from each other, while some have processes in common. To abstract this from the consumers of locally we built the pipelines where we give the teams the ability to automate the necessary steps to get that specific service up and running on a reproducible form on each machine regardless of the operating system.  

Pipelines rely of a concept called `worker` - something that performs the actions that pipeline orchestrates. There are several different type of workers provided and pipelines set an execution order of the workers, which are needed to get a service up and running.


## Pipeline Definition Schema

The [schema document](../configuration/contexts/local/services/pipelines/template.yml.md) explains the basic of what pipeline's structure and parameters are. Please review the document before attempting to create your own pipelines.

## Supported Jobs Types

Pipeline definition contains a collection of Jobs, which group actions needed to be carried out by a `worker` of a specific type that performs specific functionality.  

Below is a review of available worker types with details on how to define and configure them.

### Bash Worker

While it is called bash, it does not use bash to execute commands, it is rather a cross OS execution of commands.

**Attention**: the command carried out by the worker needs to be compatible with **all** supported operating systems, if not please let the users know what operating system it is compatible with.

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: check_version
    # the worker will be of type bash
    type: bash
    # it will take the following inputs
    inputs:
      # this will be the command that should be executed
      command: 'terraform'
      # arguments will be a list of arguments you will want to pass, you will need to pass in the
      # arguments that are separated with a space in the command line
      arguments:
        - "version"
      # this is the working directory where the command will be executed from
      workingDir: ''
```

### Curl Worker

Curl worker is an HTTP client that allows you to make HTTP calls to a particular endpoint. There are no restrictions on HTTP verbs here, plus the request can contain a payload.

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: test
    # the worker will be of type curl
    type: curl
    # it will take the following inputs
    inputs:
      # host for the call
      host: 'http://localhost'
      # http verb ot method to use for the call
      verb: 'GET'
      # Headers to add to the request, for example authorization
      headers:
        Authorization: 'Basic dXNlcjpwYXNz'
      # if your request requires a body this is where you can define it
      content:
        contentType: 'application/json'
        # you can have either a url encoded body
        urlEncoded:
          scope: 'api1'
        # or a json type of body
        json: '{ "scope": "api1" }'
      # you can ask the worker to retry x amount of time if the response is invalid, for example while
      # waiting for a service to be available
      retryCount: 3
      # the amount of seconds to wait between retry calls
      waitFor: 10
```

### Docker Worker

The docker worker uses the internal docker command, which allows automation of locally docker commands in sequence, like for example the pull of an image or the build of one

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: example
    # the worker will be of type docker
    type: docker
    # it will take the following inputs
    inputs:
      # command is to tell the locally docker what commands to execute, you can get a full list by running
      # locally docker --help
      command: 'pull'
      # the docker registry url
      registry: ''
      # the username for the docker registry
      username: ''
      # the password for the docker registry
      password: ''
      # base path for the image manifest, this can be left empty and use only the imagePath property
      basePath: ''
      # image path for the manifest
      imagePath: ''
      # this can be left empty, if empty it will try to download the latest and fix it at that
      imageTag: ''
      # this is the configuration of the locally backend service or spa service, this will be used to pull
      # the rest of the configuration for the service allowing the pipeline to be more lean
      configName: ''
      # arguments to pass to the locally docker, you can get a full list by running locally docker --help
      arguments: 
        - ''
      # environment variables to pass to the locally docker
      environmentVars: 
        - ''
```

### Dotnet Worker

The Dotnet worker is a special worker executing `dotnet` tool. The worker generates a custom docker container into which it clones a repository and executes the `dotnet` tool inside the container. This enables predictable and cross machine execution.

Use this for

- running specific tooling you need to get your service up and running
- running integration tests or smoke tests for the components you work on

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: example
    # the worker will be of type dotnet
    type: dotnet
    # it will take the following inputs
    inputs:
      # what dotnet command to run
      command: run
      # the context from where to build the docker container, this is usually a dot
      context: '.'
      # baseImage what is the image the docker compose should use, if left empty it will use the focal
      # dotnet 6
      baseImage: ''
      # github access token with enough permissions to clone the repo in question
      repoAccessToken: ${{ global.git_access_token}}
      # the github url for your project so it can be cloned
      repoUrl: ''
      # the project path where the target project is
      projectPath: ''
      # arguments to be passed to the docker container file
      arguments: 
        FEED_ACCESSTOKEN: ${{ global.feed_accesstoken}}
      # environment variables that will be passed to the docker container on execution
      environmentVars:
        Scope: api1
      # arguments passed to the dotnet command that will be executed
      buildArguments:
        - "connString=something"
```

### .NET Entity Framework Migrations Worker

Focuses on supporting .NET Entity Framework DB migration functionality. Is similar to the `dotnet` worker but is more specialized so it can run EF Migrations actions, which makes migrations tasks easier to perform.

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: example
    # the worker will be of type migrations
    type: migrations
    # it will take the following inputs
    inputs:
      # baseImage what is the image the docker compose should use, if left empty it will use the focal
      # dotnet 6
      baseImage: ''
      # github access token with enough permissions to clone the repo in question
      repoAccessToken: ${{ global.git_access_token}}
      # the github url for your project so it can be cloned
      repoUrl: ''
      # the project path where the target project is
      projectPath: ''
      # this is the startup project that will evoked the migrations, depending on how the migration was
      # build this can be left empty
      startupProjectPath: ''
      # arguments to be passed to the docker container file
      arguments: 
        FEED_ACCESSTOKEN: ${{ global.feed_accesstoken}}
      # environment variables that will be passed to the docker container on execution
      environmentVars:
        Scope: api1
```


### Git Worker

git worker, as the name implies, helps to checkout a repo locally

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: example
    # the worker will be of type git
    type: git
    # it will take the following inputs
    inputs:
      # the github url for your project so it can be cloned
      repoUrl: ''
      # destination is where do you want the repo to be cloned to
      destination: ''
      # if you set the clean to true it will delete the content of the repo on the next run if it exists
      clean: false
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
```

### Infrastructure Worker

Infrastructure is a worker that leverages the locally infrastructure command in the pipelines and can be used to run the infrastructure command to any of the stacks including the dependencies

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: example
    # the worker will be of type infrastructure
    type: infrastructure
    # it will take the following inputs
    inputs:
      # command is to tell the locally infrastructure what commands to execute, you can get a full list
      # by running locally infrastructure --help
      command: 'up'
      # what is the stack to apply the operation
      stackName: ''
      # list of arguments to pass to the command, can be left empty
      arguments:
        - "test"
      # with this enabled we will be running the stack all all of it's dependencies
      buildDependencies: true
      # this is the working directory where the command will be executed from
      workingDir: ''
```

### KeyVault Worker

Keyvault worker is used to synchronize Azure KeyVault instance with locally Environment Vaults and make those variables available to the rest of the pipeline

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: example
    # the worker will be of type keyvault
    type: keyvault
    # it will take the following inputs
    inputs:
      # the keyvault url, this normally will be the one that the stacks will output and in most cases
      # will be the environment variable ${{ terraform.core-stack.globalKvUrl}}
      keyvaultUrl: ${{ terraform.core-stack.globalKvUrl}}
      # this will attempt to base64 decode the value returned by the keyvault, by default this should be
      # left set to true
      base64Decode: true
```

### SQL Worker

SQL worker executes a SQL command directly on a SQL server. Could be used to create databases before a service is deployed.

```yaml
    # name of the worker, this is used mostly for logging purpose
  - name: example
    # the worker will be of type keyvault
    type: keyvault
    # it will take the following inputs
    inputs:
      # connection string to the sql server you wish the command to be executed on
      connectionString: ''
      # sql query that you want to be executed, this can spread across multiple lines by adding | to the
      # start of the query
      query: ''
```
