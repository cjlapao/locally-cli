# Pipelines Schema template

Pipelines is a concept in locally to allow automation of different steps to get the specific service running. this is because a service is more than just it's infrastructure or its container, there is for example the migrations for the database or the ems registration and a lot more. for this we build a pipeline schema that is similar to the ADO pipelines and where you can add ```workers``` to do operations, these can range from a clone to for example running dotnet ef migrations.  
We will only show the schema .

```yaml
# All pipelines definitions needs to start with this, the reason is locally is folder agnostic
# so while we place all the pipelines definitions in the same folder it does necessarily
# needs to be so, adding that allows locally to set the configuration in the right place
pipelines:
  # A pipeline is an array, you can define multiple pipelines per file and they just
  # need to have unique names, they also have a naming rule, it needs to be alphanumeric and can 
  # contain dashes and underscores but it cannot contain spaces, for example example-service
  - name: example-service
    # each pipeline is constituted by a set of jobs, you can have as many jobs as you want and they will
    # be run in sequence
    jobs:
        # each job do need to have unique name, it also have a naming rule, it needs to be alphanumeric
        # and can contain dashes and underscores but it cannot contain spaces, for example example-job 
      - name: example-job
        # each job can be disabled, this will signal the locally that it should not run it
        disabled: false
        # each job can have an array of steps, steps are like tasks and they are like jobs run in sequence, the step will take a type and calls a worker that does a specific job
        steps:
            # each job do need to have unique name, it also have a naming rule, it needs to be
            # alphanumeric and can contain dashes and underscores but it cannot contain spaces,
            # for example checkout
          - name: checkout
            # we need to define a type of worker, there are a few available and more to come, in this
            # example we are going to use the git worker, this is as the name implies the command git
            type: git
            # while inputs is mandatory the properties will be different from worker to worker and you
            # will need to check the documentation for each worker to better understand it
            inputs:
              # in this case the git takes these inputs and the inputs can be driven by the global
              # variables as their values
              repoUrl: https://github.com/org/example-service.git
              clean: false
              credentials:
                privateKeyPath: ${{ global.git_private_key_path}}
```
