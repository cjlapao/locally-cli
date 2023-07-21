# Infrastructure Schema template

The infrastructure template show you the full template and how you can set up the stacks to run your component locally

```yaml
# All infrastructure definitions start with this, this will tell locally that this config is of the infrastructure
# type and will add it to the list of known infrastructure
infrastructure:
  stacks:
    - name: login-app-stack
      dependsOn:
        - config-stack
        - core-stack
        - security-stack
        - storage-stack
      requiredStates:
        - core-stack
        - security-stack
        - storage-stack
      location:
        # rootFolder: C:\Code\infra
        path: main\05_component_stack\05_login_app
      repository:
        enabled: true
        url: https://github.com/org/example.git
        destination: ${{ config.path.sources }}/infra-terraform
        credentials:
          privateKeyPath: ${{ global.git_private_key_path }}
      variables:
        validationEnvironment: true
      backend:
        stateFileName: loginApp.tfstate
      tags:
        - core
        - ui
```
