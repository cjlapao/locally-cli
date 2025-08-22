# Command Line Interface reference for locally

## Table of Contents

- [Command Line Interface reference for locally](#command-line-interface-reference-for-locally)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [locally Functionality](#locally-functionality)
    - [Groups of commands](#groups-of-commands)
    - [Config](#config)
      - [list](#list)
      - [current-context](#current-context)
      - [set-context](#set-context)
    - [Certificates](#certificates)
    - [Docker](#docker)
      - [build](#build)
      - [rebuild](#rebuild)
      - [delete](#delete)
      - [up](#up)
      - [down](#down)
      - [start](#start)
      - [stop](#stop)
      - [pause](#pause)
      - [resume](#resume)
      - [status](#status)
      - [list](#list-1)
      - [logs](#logs)
      - [generate](#generate)
    - [Env](#env)
    - [Keyvault](#keyvault)
    - [Hosts](#hosts)
    - [Infrastructure](#infrastructure)
    - [Pipelines](#pipelines)
    - [Proxy](#proxy)
    - [Nuget](#nuget)
    - [Tools](#tools)

## Introduction

locally has grown into a very comprehensive tool providing a lot of functionality to manage local environments (create, start, stop), manage infrastructure required by the services in the environment, generate certificates, interface with Azure services, automatically adjust system files on engineers' computers, manage docker containers, etc. There is a wealth of functionality and it's growing.  

This document provides an extensive reference to commands implemented by locally.

## locally Functionality

The best way to explore what is supported by the tool is to run it with `--help` switch. The following command will print the list of operations the tool supports:

```bash
locally --help
```  

### Groups of commands

As this point the following locally groups of commands are supported (yes, the following is a _group_ of commands - each group actually has many commands related to the group):

- _config_          - Sets the configuration context and switches between them
- _certificates_    - Generates a self-signed valid chain certificates for local development
- _docker_          - Controls services docker operations from generation to lifecycle
- _env_             - Allows to query configuration variables
- _keyvault_        - Allows manual synchronization of an Azure keyvault into configuration
- _hosts_           - Controls system host file changes to help generate custom entries
- _infrastructure_  - Builds the required infrastructure for the services based on the stacks
- _pipelines_       - Runs locally specific integrated pipelines for easy management of services
- _proxy_           - Controls Caddy proxy service allowing to generate/update configuration
- _nuget_           - Builds NuGet packages and adds them to a local feed
- _tools_           - Some other useful developers tools
  
For each of these groups you can get a list of actual commands supported with some brief descriptions by executing:

```bash
locally [COMMAND-GROUP-NAME] --help
```

### Config

The config command allows us to check the configuration that locally is currently running on and also allows to change between contexts

There are different sub commands available to you, to list all the available commands you can run

```bash
locally config
```

#### list

**list** This command allows you to list all the available configuration in locally, this will show all the backend, frontend, pipelines, infrastructure, etc.

```bash
locally config list
```

you should see a list of available services in your context, this is also a good way to test if your configuration is correct as if there is any error the list command will issue an error instead of the list of services.  
You can also filter that list by adding an extra option, for example, imagine we would only want to list the pipelines, we could then run

```bash
locally config list pipelines
```

To view a full list of the filter you can add, run

```bash
locally config list --help
```

#### current-context

We can use this to show us what is the current selected context in locally, just run

```bash
locally config current-context
```

and this will show you what context is currently selected

#### set-context

The set-context sub command is used to change contexts in locally.
locally comes with a concept called a context, where we can have different sets of configuration that can create a multitude of environments, one of the use cases is if an engineer needs to move from a team to another team, he might decide to create a new context and change the configuration accordingly. he will then have the ability to move from one configuration to the other at ease.

**Attention:** Changing context will clean up some files like for example the caddy ones so you might need to run some of the generation or deployment commands. this happens due to you can only run a version of a container.  

To change context you can simple run

```bash
locally config set-context <name_of_context>
```

### Certificates

### Docker

The locally docker is target at the debugging side of the environment. It allows you to build and start containers locally, it also allows for example to stop a specific container so you can start the debugger and allow the proxy to forward the traffic there.

With this command and due to the complexity of some of the use cases we are going to take a different approach, Firstly we will describe how to use the commands and finally we then will be adding some use cases and how would you change your configuration to debug.

We are going to use the config-service as an example service in this document, this service contains two components, the config-service and the config-service-proxy component.

#### build

**build** is used to build the docker containers from the source code, it can only be used if you do have the source in your machine otherwise it will issue an error. you can build the whole of the components of a service or a specific component, you can also use the tags to run different chained commands.

To run it for all components in a service use:

```bash
locally docker build config-service
```

In this case it will build all of the component associated with it, two in this case

To target the build at a specific component you can do this

```bash
locally docker build config-server config-service-proxy
```

This will build exclusively the config-service-proxy container which is a component of the config-service

#### rebuild

**rebuild** is a similar command to the build we saw before, main difference is the rebuild will clean any cache docker has done and delete any existing image for the service in the local machine. While you can target a single component on the previous command, the rebuild will always rebuild all of the images, regardless of passing the component in the parameters. This is due to constrains in how docker-compose organizes internally

To run it you can use this:

```bash
locally docker rebuild config-service
```

While this is a legal command, it will still rebuild the whole service

```bash
locally docker rebuild config-server config-service-proxy
```

#### delete

**delete** is used to delete an image from the docker, this can be used for a cleanup process, it simple will delete the images related to a specific service, as the rebuild this cannot target a single component and it will delete all the images in a service

To run it you can use this:

```bash
locally docker delete config-service
```

While this is a legal command, it will still delete all the images for a service

```bash
locally docker delete config-server config-service-proxy
```

#### up

**up** is potentially one of the most used commands in the docker tooling, this not only will start your service but it can also build if it does not exist, so it works a bit like a chain of commands to get the docker running and healthy. there is not a lot to it like all of the other commands.

To run it for all components in a service use:

```bash
locally docker up config-service
```

In this case it will build (if necessary) and start all of the component associated with it, two in this case

To target the build at a specific component you can do this

```bash
locally docker up config-server config-service-proxy
```

This will build (if necessary) and start exclusively the config-service-proxy container which is a component of the config-service

#### down

**down** command is the reverse of the up, it will stop the service containers and it will clean up all of the images ready for you to start up again, because of the destructive nature of the command, if you only looking at stopping a service container you might want to use the [stop](#stop) command instead.  
As all the destructive commands before you cannot target a single component and it will always delete all of them.

To run it use:

```bash
locally docker down config-service
```

While this is a valid command it will still execute the down on all of the service components

```bash
locally docker down config-server config-service-proxy
```

#### start

**start** command allows you to start an existing service containers or a single container, this is useful if you for example stopped one for debugging purposes, or after a machine restart. This will not attempt to build so if the container does not exist in your docker it will just issue an error.

To run it for all components in a service use:

```bash
locally docker start config-service
```

To run it for a single component in a service use:

```bash
locally docker start config-server config-service-proxy
```

#### stop

**stop** command allows you to stop an existing service containers or a single container, this is useful if you for example if you want to stop a specific container to start a debugging session. If the specific container(s) does not exist an error will be issued

To run it for all components in a service use:

```bash
locally docker stop config-service
```

To run it for a single component in a service use:

```bash
locally docker stop config-server config-service-proxy
```

#### pause

**pause** command allows you to pause an existing service containers or a single container. One of the questions you might have is, what is the difference between pause and stop? While there is a lot of technicalities behind the difference, the easiest way to explain is, stop will make the container exit, refreshing all the memory and releasing all of the resources, pause will just pause the execution where it is and the resources like memory will still be in play. While different operating systems use different approaches this is the basic explanation.

To run it for all components in a service use:

```bash
locally docker pause config-service
```

To run it for a single component in a service use:

```bash
locally docker pause config-server config-service-proxy
```

#### resume

**resume** command allows you to resume an existing paused service containers or a single container.

To run it for all components in a service use:

```bash
locally docker resume config-service
```

To run it for a single component in a service use:

```bash
locally docker resume config-server config-service-proxy
```

#### status

**status** can be used to check on a service or component container status, this will show you if the service is running, when it was created and for how long has been running.

To run it for all components in a service use:

```bash
locally docker status config-service
```

To run it for a single component in a service use:

```bash
locally docker status config-server config-service-proxy
```

#### list

**list** can be use to list either all of the services currently running or to a more granular level of a service, it allows you to know how many container are also up in each of the service

To run it for all services use:

```bash
locally docker list
```

To run it for a single service use:

```bash
locally docker list config-service
```

#### logs

**logs** is used to get a specific service or container running logs, it will by default only show the recent history, but you can add the flag ```--follow``` and this will make locally keep displaying the following logs until you press ```ctrl + c```

To run it for all components of a service use:

```bash
locally docker log config-service
```

> add the --follow to keep the command from exiting and do a tail

To run it for a single component of a service use:

```bash
locally docker log config-service config-service-proxy
```

> add the --follow to keep the command from exiting and do a tail

#### generate

**generate** command is used to generate an ```docker-compose.override.yaml``` for a service. This is normally used if you are running a service from source and you also have a ```docker-compose.yaml``` present. In this particular case to avoid checking in code we generate a ```docker-compose.override.yaml``` to allow us to set environment variables while avoiding it to be checked in. This command will generate it for you based on the service configuration. you can also run the generate with the flag ```--all``` and this will just generate that files for all of your services in one go.

> When changing context this command is executed automatically as the **docker-compose.override.yaml** might have different values

To run it for all services use:

```bash
locally docker generate --all
```

To run it for a single service use:

```bash
locally docker generate config-service
```

To run it for a single component of a service use:

```bash
locally docker generate config-service config-service-proxy
```

### Env

### Keyvault

### Hosts

### Infrastructure

### Pipelines

### Proxy

### Nuget

### Tools
