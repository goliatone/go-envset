# envset

`envset` run commands in an environment defined using a [ini][ini] configuration file.

---

## Environment level configuration

Application configuration is (usually) specific to an environment and will change between different build environments- e.g. app secrets for a staging environment are different than your production secrets.

The [12 factor app][12factor] guidelines suggest you store your application's configuration in the environment.

Environment variables enable us to manage application configuration outside of our application code.

Application configuration usually are small and sensitive data such as API keys or tokens, database credentials, etc. However not all environment configuration have to be secrets, there might be build distribution specific values such as the application's base URL to build OAuth callbacks, a dependent service endpoint or anything that changes between development and production environments.

`envset` helps you manage environment variables for multiple build distributions.

Is as simple as calling:

```console
$ envset development -- node server.js
```

This will load the variables defined in the `[development]` section of a local `.envset` in the shell environment and execute the command after the `--`, in this instance `node server.js`.

See the [examples](#examples) section for more details.

## Similar Tools

Inspired by [daemontools][dtools]' tool [envdir][envdir] and tools such as [dotenv](https://github.com/bkeepers/dotenv).

* Distributed as a single binary
* No dependencies in your codebase
    * e.g. `dotenv-rails` and `dotenv`<sup>[1](#node-dotenv)</sup> for Node.js require you to use a library
* Support multiple environments in a single file
* Generates an example file with your current env vars to keep documentation updated.
* Interpolation of variable using POSIX variable expansion.
* Command expansion
* (required) Define required variables and exit with error if not set
* (isolated) By default the shell environment is not loaded in the context

Instead of having an `.env` file per environment you can have one single `.envset` file with one section per environment. 

<a name="node-dotenv">1</a>: You an actually require the library outside of your project with the `node -r` flag.

## Examples

### Executing a command

An **.envset** file could look like this:

```ini
[production]
NODE_AWS_SECRET_ACCESS_KEY=FS40N0QY22p2bpciAh7wuAeHjJURgXIBQ2cGodpJD3FRjw2EyYGjyXpi73Ld8zWO
NODE_AWS_ACCESS_KEY_ID=LSLhv74Q1vH8auQKUt5pFwnix0FUl0Ml
NODE_HONEYBADGER_KEY=LCgZgqsxKfhO
NODE_POSTGRES_ENDPOINT=50.23.54.25
NODE_POSTGRES_DATABASE=myproject
NODE_POSTGRES_PSWD=Pa$sW03d
NODE_POSTGRES_USER=myproject

[development]
NODE_AWS_SECRET_ACCESS_KEY=HN5ITok3lDaA1ASHxtEpy1U9XCjZwThmfgoIYZk8bkOqc5yk6sT7AWd3ooNeRFV9
NODE_AWS_ACCESS_KEY_ID=m35W4PGGVZfj1gOxYvztFxBD5F2605B3
NODE_HONEYBADGER_KEY=f3MNPUhZoki6
NODE_POSTGRES_ENDPOINT=localhost
NODE_POSTGRES_DATABASE=postgres
NODE_POSTGRES_PSWD=postgres
NODE_POSTGRES_USER=postgres
```


To use it, simply prefix the call to your program with `envset` and the name of the environment section. The node `app.js` will be running with the environment variables specified in the **development** section of the **.envset** file.

```console
$ envset development -- node app.js
```

#### Variable substitution

You can execute commands that use environment variables in the command arguments.

Is important to note that you need to scape the variable so that it is not replaced in the shell as you call `envset`. You can do so by using single quotes `'` or the scape char `\$`.

```sh
$ envset development -- say '${MSG}'
$ envset development -- say \${MSG} 
```

#### Inherit environment

You can control environment inheritance using two flags:

- `--isolated`
- `--inherit`

By default `envset` will run commands in a clean environment. Sometimes you want the executed command to access the host's environment. To do so you need to pass the `--isolated=false` flag.

```console
$ envset development --isolated=false -- spd-say '${APP_NAME}' 
```

Some commands might rely on environment variables set on your shell, for instance if you want to `go run`:

```console
$ envset development -- go run cmd/app/server.go
missing $GOPATH
```
You will get an error saying that `$GOPATH` is not available. You should run the command with the `--isolated=false`:

```console
$ envset development --isolated=false -- go run cmd/app/server.go
```

The `-inherit` flag lets you specify a list of environment variable keys that will be inherited from the parent environment.

In the previous example instead of exposing the whole parent environment we could just expose `$GOPATH`:
``console
$ envset development -I=GOPATH -I=HOME -- go run cmd/app/server.go
```

#### Load env file to current shell session

If you want to make the variables defined in a env file to your running shell session use something like the following snippet.


```sh
$ eval $(envset development)
```

#### Required environment variables

You can specify a list of required environment variables for your command using the `--required` flag or its `-R` alias.

Given the following env file:

```ini
APP_MESSAGE="this is a test"
```

If you run the following command:

```console
$ envset development --required=BOOM -R BOOM2 -- node index.js
```

`envset` will exit with an error and a message with the missing variables:

```console
missing required keys: BOOM,BOOM2
```

### Generating an example template

If we run the `envset template` command with the previous **.envset** file we generate a **envset.example** file:

```ini
[development]
NODE_AWS_SECRET_ACCESS_KEY={{NODE_AWS_SECRET_ACCESS_KEY}}
NODE_AWS_ACCESS_KEY_ID={{NODE_AWS_ACCESS_KEY_ID}}
NODE_HONEYBADGER_KEY={{NODE_HONEYBADGER_KEY}}
NODE_POSTGRES_ENDPOINT={{NODE_POSTGRES_ENDPOINT}}
NODE_POSTGRES_DATABASE={{NODE_POSTGRES_DATABASE}}
NODE_POSTGRES_PSWD={{NODE_POSTGRES_PSWD}}
NODE_POSTGRES_USER={{NODE_POSTGRES_USER}}

[production]
NODE_AWS_SECRET_ACCESS_KEY={{NODE_AWS_SECRET_ACCESS_KEY}}
NODE_AWS_ACCESS_KEY_ID={{NODE_AWS_ACCESS_KEY_ID}}
NODE_HONEYBADGER_KEY={{NODE_HONEYBADGER_KEY}}
NODE_POSTGRES_ENDPOINT={{NODE_POSTGRES_ENDPOINT}}
NODE_POSTGRES_DATABASE={{NODE_POSTGRES_DATABASE}}
NODE_POSTGRES_PSWD={{NODE_POSTGRES_PSWD}}
NODE_POSTGRES_USER={{NODE_POSTGRES_USER}}
```


### Support for .env files

You can load other environment files like `.env` files:

```console
$ envset --env-file=.env -- node index.js
```

### Metadata

The `metadata` command will generate a JSON file capturing the values of the provided env file.

### Metadata Compare

Note that `envset metadata compare` will output to stderr in the case that both files do not match.

```console
$ envset metadata compare --section=development .metadata/data.json staging-metadata.json 2>&1 | jq . 
```

## Installation

### macOS
<!-- 
TODO: List how to install in all different platforms
-->

Add tap to brew:

```console
$ brew tap goliatone/homebrew-tap
```

Install `envset`:

```console
$ brew install envset
```


### Ubuntu/Debian

```console
$ export tag=<version>
$ cd /tmp
$ wget https://github.com/goliatone/go-envset/releases/download/v${tag}/envset_${tag}_linux_x86_64.deb
$ sudo dpkg -i envset_${tag}_linux_x86_64.deb
```

### CentOS/Redhat

```console
$ yum localinstall https://github.com/goliatone/go-envset/releases/download/v<version>/envset_<version>_linux_x86_64.rpm
```

### Manual Install

```console
$ wget https://github.com/goliatone/go-envset/releases/download/v<version>/envset_<version>_linux_x86_64.tar.gz
$ tar -C /usr/bin/ -xzvf envset_<version>_linux_x86_64.tar.gz envset
$ chmod +x /usr/bin/envset
```

## Documentation

`envset` will look for a file defining different environments and make them available as commands.

```ini
[development]
APP_BASE_URL=https://localhost:3003

[production]
APP_BASE_URL=https://envset.sh
```

### Commands

The following is a list of the available commands:

* metadata
    * compare
* template

### Variable Expansion

`envset` can interpolate variables using POSIX variable expansion in both the loaded environment file and the running command arguments. 

```ini
[development]
CLIENT_NAME=$(whoami -f)
CLIENT_ID=${CLIENT_NAME}.devices.local
```

```
$ envset development -- node cli.js --user '${USER}'
```

### Commands

If you type `envset` without arguments it will display help and a list of supported environment names.

## .envset file


## .envsetrc
You can create an `.envsetrc` file with configuration options for `envset`.

The default `.envsetrc` looks like this:

```
expand=true
isolated=true
filename=.envset
export_environment=APP_ENV

[metadata]
dir=.meta
file=data.json

[template]
path=.
file=envset.example

[environments]
name=test
name=staging
name=production
name=development
```

### Configuration

Follows `rc` [standards][rcstand].


### Configuration Syntax

The loaded files need to be valid `ini` syntax.

```ini
[development]
APP_BASE_URL=https://localhost:3003

[production]
APP_BASE_URL=https://envset.sh
```


## License
Copyright (c) 2015-2021 goliatone  
Licensed under the MIT license.



[ini]: https://en.wikipedia.org/wiki/INI_file
[dtools]: http://cr.yp.to/daemontools.html
[envdir]: http://cr.yp.to/daemontools/envdir.html
[rcstand]: https://github.com/dominictarr/rc#standards
[12factor]: http://12factor.net/config
[vcn]: https://github.com/goliatone/vcn
[npm-fix-perm]:https://docs.npmjs.com/getting-started/fixing-npm-permissions



<!-- 
Add self-update 
* [go-update](https://github.com/tj/go-update)
* [go-update](https://github.com/inconshreveable/go-update)
* [s3update](https://github.com/heetch/s3update): Related article [here](https://medium.com/inside-heetch/self-updating-tools-in-go-lang-9c07291d6285)

documentation extension for urfav/cli/v2
https://github.com/clok/cdocs
-->