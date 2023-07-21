# Running a component using source code, GitHub repository of a pre-built Docker image

## Introduction

The major use case for a local environment is to debug services you develop while running other dependencies seamlessly. Great deal of though had been put into making that possible with locally - you should be able to start/stop containers, run code in a debugger, run (and debug!) integration tests locally. The latter is a big deal - before you could only do that against a real environment, potentially used by others.

Default locally behavior is based on [Pipelines](../pipelines.md) concept, which is a way to package services as _third party_ - meaning locally downloads an image for the service from the container registry and runs that in your local Docker instance. This approach greatly simplifies running any dependencies you need with as little work as possible (especially with pipelines for all services being pre-created and made available). It is recommended to rely on this approach for anything you are not actively developing.

## Configuring where component executable comes from

All components in locally are described with definition YAML files, which are located under `contexts\ENV_NAME\services` directory - there are few subdirectories there:

```ascii
locally/
├─ contexts/
│  ├─ local/
│  │  ├─ services/
│  │  │  ├─ backends/
│  │  │  ├─ infrastructure/
│  │  │  ├─ mocks/
│  │  │  ├─ pipelines/
│  │  │  ├─ tenants/
│  │  │  ├─ webclients/
│  │  ├─ context-config.yml
├─ locally-config.yml
├─ locally.exe
```

Services are defined via YAML files in the following two directories - `backend` (for cloud services running business logic) and `webclients` (for containers serving WebClient Angular components). These definition files are where all of the rules and settings are defined.

Here is the definition file content for Configuration Service:

```yaml
backendServices:
  - name: config-service
    location:
      # rootFolder: C:\Code\example-Service
    repository:
      enabled: false
      url: https://github.com/org/example-service.git
      destination: ${{ config.path.sources }}/example-Service
      credentials:
        privateKeyPath: ${{ global.git_private_key_path}}
    dockerRegistry:
      enabled: true
      registry: ${{ global.docker_registry }}
      basePath: ${{ global.docker_base_manifest_path }}
      credentials:
        username: ${{ global.docker_username }}
        password: ${{ global.docker_password }}
    dockerCompose:
      services:
        config-service:
          volumes:
            - ${{ config.config_service.data.path }}:/app/config-service
          ports:
            - 5510:5000
        config-service-proxy:
          volumes:
            - ${{ config.config_service.data.path }}:/app/config-service
          ports:
            - 5511:5000
```

In the above document there are few points of interest for us now:

- `rootFolder` node under `location` is commented out
  - This makes locally unable to look for code on the machine to build the service binaries
- `dockerRegistry` is set to be **Enabled**, which instructs locally to use the container registry to obtain services image

Now, if we want our service to be built from a local source and not pulled from a container registry, we need to make simple changes:

- uncomment `rootFolder` node under `location` and point it to the root of the repo directory holding the source code
  - another option is if you want to pull source from GitHub and then have locally build the component from the code that it pulls from the repository
    - In this case you need to **enable** `repository` node and ensure that URL and Destination parameters are properly initialized
    - Obviously make sure to comment out `rootFolder` node under `location` in this case
- set `dockerRegistry` enabled flag to be **false**

```yaml
backendServices:
  - name: config-service
    location:
      rootFolder: C:\Code\example-service
    repository:
      enabled: false
      url: https://github.com/orgc/example-service.git
      destination: ${{ config.path.sources }}/example-service
      credentials:
        privateKeyPath: ${{ global.git_private_key_path}}
    dockerRegistry:
      enabled: false
      registry: ${{ global.docker_registry }}
      basePath: ${{ global.docker_base_manifest_path }}
      credentials:
        username: ${{ global.docker_username }}
        password: ${{ global.docker_password }}
    dockerCompose:
      services:
        config-service:
          volumes:
            - ${{ config.config_service.data.path }}:/app/config-service
          ports:
            - 5510:5000
        config-service-proxy:
          volumes:
            - ${{ config.config_service.data.path }}:/app/config-service
          ports:
            - 5511:5000
```

After these changes Configuration Service (or any other component for which you change where locally gets the executables from) is going to be run after being built from source code first. This change can be reverted anytime - everything is based on your use cases and how you want to run various services in your local environment. 
