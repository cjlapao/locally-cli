# Locally


locally is a command line tool design to help spin up a local environment including the infrastructure, the concept is easy, have configuration files created by each team that can be shareable and reproducible from machine to machine and deploy the bare minimum infrastructure.

## How to install

locally is not installed, it is just an executable with example/template configuration bundle so you can just download the latest release, unzip it to a folder and put it in the environment path.  


## Troubleshoot

We have a troubleshoot guide [here](./docs/troubleshooting.md) where we place the most common issues found by people, this will be constantly updated, so please be sure to read it before you ask questions.  
If you do not find an answer there you can use the locally channel to ask a question to the team [here](https://teams.microsoft.com/l/channel/19%3a98b5d070649f442ab23b247ec5858e16%40thread.skype/locally%2520-%2520Also%2520called%2520Locally?groupId=cd5ee759-4aef-4928-95f4-b8c658c5d0db&tenantId=e5208e76-dd12-47f0-9541-c9b45afaffe6).  


## Building locally locally

locally is written in Go and uses VSCODE to easily debug, you still need a few tools if you do not have in case you want to build it from source, or just debug it.

### Getting and installing go onto your PC

Download the latest Go from [here](https://go.dev/dl/), choose your operating system and then run. Note: if you are running Mac or Linux this needs to be unziped to a folder.  

Once this is done you can quickly test it by typing ```go version```

### Visual Studio Code with GO

There is a good setup guide in [here](https://code.visualstudio.com/docs/languages/go) this will use the extensions provided by google and allow intellisense in vscode
this is the [extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)

### How to build

If you want to test your changes can actually build but not generate the executable follow this:

1. Go to the ```src``` folder
2. Run ```go build -v ./...```

This will just build the code but with no output, if you want to generate an output you can do this

1. Go to the ```src``` folder
2. Run ```go build -o locally.exe```

This will be building the locally but will create a locally.exe as the output
