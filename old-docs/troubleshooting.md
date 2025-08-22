# Troubleshooting locally

## locally is failing but I don't know why
locally can provide additional log levels, please add `--debug` to your command.

## locally is only handling the stack I specified
locally on purpose does only the provided stack. If you want to have the whole dependency tree being managed, specify the flag `--build-dependencies`.

## locally doesn't update the stacks repository
locally does not automatically update your local version of the infrastructure git repositories in order to avoid breaking your local version. If you are sure you want to use the latest version of the infrastructure repositories, please use `--clean-repo` flag on your command line.

## Executing "infrastructure" command I get an error while downloading modules

This could happen if your path is too long, try adding this using an administrative command line:

```bash
git config --system core.longpaths true
```

This could also happen if the output folder for locally is too *deep*, you can also try to move it closer to the root

## An error while accessing my SQL server

- Make sure you have sql network access enable
- Make sure your user has access to the SQL service using network
- Check if you are not using localdb, this is not accessible using network

## locally proxy cannot attach to port 80

- Check if you have IIS running
- Check if any other application is using port 80
  - On windows you can try:
      ```netsh http show servicestate view=requestq verbose=yes```


## When running an image I get a mount access error from docker

- In your docker go to settings, resources and check if you have an option to add file share, if you do add that specific folder to that list and restart docker

## When running locally I get an error message "unable to open tcp connection with host 'host.docker.internal:1433': dial tcp 192.168.68.112:1433: i/o timeout"

- This happens if you have switched network, for example going from home to the office or vice versa, restarting docker daemon should fix the issue.

## When running locally proxy run I get error: loading initial config: loading new config: loading http app module: provision http: getting tls app: loading tls app module: provision tls: loading certificates: tls: failed to parse private key 

This can happen if in the ```locally-config.yml``` your certificates path is not with double quotes, try to make the path like this:

```yaml
    certPath: "c:\\somepath\\cert.crt"
    privateKeyPath: "c:\\somepath\\cert.key"
```

## When running infrastructure initialization failed with downloading providers

- check if your output folder is very long, for example if your output folder is c:\folder1\folder2\folder3\output, this makes the paths very long and it might break terraform, try to move the folder closer to the root and try again
