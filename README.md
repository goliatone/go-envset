# envset

`envset` run commands in an environment defined using a [ini][ini] configuration file.

---

<!-- vscode-markdown-toc -->
* [Environment level configuration](#environment-level-configuration)
* [Similar Tools](#similar-tools)
* [Examples](#examples)
	* [Executing A Command](#executing-a-command)
		* [Restart Command](#restart-command)
		* [Variable Substitution](#variable-substitution)
		* [Inherit Environment](#inherit-environment)
		* [Overwriting Variables At Runtime](#overwriting-variables-at-runtime)
		* [Load Env File To Current Shell Session](#load-env-file-to-current-shell-session)
		* [Required Environment Variables](#required-environment-variables)
	* [Generating An Example Template](#generating-an-example-template)
	* [Support For .env Files](#support-for-envfiles)
	* [Metadata](#metadata)
	* [Metadata Compare](#metadata-compara)
		* [Ignore Variables](#ignore-variables)
* [Installation](#installation)
	* [macOS](#macos)
	* [Ubuntu/Debian x86_64 - amd64](#ubuntu-debianx86-64-amd64)
	* [CentOS/Redhat x86_64 - amd64](#centos-redhat-x86-64-amd64)
	* [Manual Install x86_64 - amd64](#manual-install-x86-64-amd64)
* [Documentation](#documentation)
	* [Commands](#commands)
	* [Variable Expansion](#variable-expansion)
	* [Commands](#commands-1)
* [.envset File](#envset-file)
* [.envsetrc](#envsetrc)
	* [Configuration](#configuration)
	* [Configuration Syntax](#configuration-syntax)
	* [Ignored And Required Sections](#ignored-and-required-sections)
* [License](#license)

<!-- vscode-markdown-toc-config
	numbering=false
	autoSave=true
	/vscode-markdown-toc-config -->
<!-- /vscode-markdown-toc -->


---

## <a name='environment-level-configuration'></a>Environment level configuration

Application configuration is (usually) specific to an environment and will change between different build environments- e.g. app secrets for a staging environment are different than your production secrets.

The [12 factor app][12factor] guidelines suggest you store your application's configuration in the environment.

Environment variables enable us to manage application configuration outside of our application code.

Application configuration usually are small and sensitive data such as API keys or tokens, database credentials, etc. However not all environment configuration have to be secrets, there might be build distribution specific values such as the application's base URL to build OAuth callbacks, a dependent service endpoint or anything that changes between development and production environments.

`envset` helps you manage environment variables for multiple build environments.

The following command will run a Node.js application with a `development` environment:

```console
$ envset development -- node server.js
```

`envset` will load the variables defined in the `[development]` section of a local `.envset` file and execute the command after the `--`.

See the [examples](#examples) section for more details.

## <a name='similar-tools'></a>Similar Tools

Inspired by [daemontools][dtools]' tool [envdir][envdir] and tools such as [dotenv](https://github.com/bkeepers/dotenv).

* Distributed as a single binary
* No dependencies in your codebase
    * e.g. `dotenv-rails` and `dotenv`<sup>[1](#node-dotenv)</sup> for Node.js require you to use a library
* Support multiple environments in a single file
* Generates an example file with your current env vars to keep documentation updated.
* Interpolation of variable using POSIX variable expansion.
* Command expansion
* Define required variables and exit with error if not set
* By default the shell environment is not loaded in the context

Instead of having an `.env` file per environment you can have one single `.envset` file with one section per environment.

<a name="node-dotenv">1</a>: You an actually require the library outside of your project with the `node -r` flag.

## <a name='examples'></a>Examples

### <a name='executing-a-command'></a>Executing A Command

An **.envset** file could look like this:

```ini
[development]
APP_SECRET_TOKEN=4c038a0b-4ed9-44e6-8682-9c82d5b831fd
APP_REMOTE_SERVICE_KEY=07ed716a-078a-4327-8847-b86394f14575

[production]
APP_SECRET_TOKEN=796bca0f-2692-495b-a994-e8ce82cad55f
APP_REMOTE_SERVICE_KEY=5655969e-af9f-4ac5-b186-184f43468443
```

To use it, simply prefix the call to your program with `envset` and the name of the environment section. The node `app.js` will be running with the environment variables specified in the **development** section of the **.envset** file.

```console
$ envset development -- node app.js
```

#### <a name='restart-command'></a>Restart Command

`envset` will optionally restart your command if it exits with an error code.
There are three flags you can use to manage the restart behavior:
* `--restart`: Restart command on exit error, default to `true`.
* `--max-restarts [int]`: Max times to restart command, defaults to `3`.
* `--forever`: If present restart the command for as long as `envset` is running.

All these can be configured using an `.envsetrc` [file](#envsetrc).

This will restart the node app for a maximum of `--max-restarts`.

```console
$ envset development --restart --max-restarts 10 -- node app.js
```

#### <a name='variable-substitution'></a>Variable Substitution

You can execute commands that use environment variables in the command arguments.

Is important to note that you need to scape the variable so that it is not replaced in the shell as you call `envset`. You can do so by using single quotes `'` or the scape char `\$`.

```console
$ envset development -- say '${APP_ENV}'
$ envset development -- say \${APP_ENV}
```

#### <a name='inherit-environment'></a>Inherit Environment

Sometimes the command you want to run will assume that has access to predefined environment variables:

```console
$ envset development -- spd-say '${APP_ENV}'
Failed to connect to Speech Dispatcher:
Error: Can't connect to unix socket ~/.cache/speech-dispatcher/speechd.sock: No such file or directory. Autospawn: Autospawn failed. Speech Dispatcher refused to start with error code, stating this as a reason:
exit status 1
```

By default the process in which `envset` runs `spd-say` in an isolated mode has no access to your shell.

You can control environment inheritance using two flags:

- `--isolated`: Inherit all parent environment variables
- `--inherit`: Inherit specific parent environment variables

If you need the executed command to inherit the host's environment wholesale use the `--isolated=false` flag.

```console
$ envset development --isolated=false -- spd-say '${APP_ENV}'
```

Some commands might rely on a known environment variable set on your shell, for instance if you want to `go run`:

```console
$ envset development -- go run cmd/app/server.go
missing $GOPATH
```

You will get an error saying that `$GOPATH` is not available. The `--inherit` flag lets you specify a list of environment variable keys that will be inherited from the parent environment:

```console
$ envset development -I=GOPATH -I=HOME -- go run cmd/app/server.go
```

#### <a name='overwriting-variables-at-runtime'></a>Overwriting Variables At Runtime

You can overwrite environment variables without editing your `.envset` file.

```console
APP_NAME="New Name" envset development --isolated=false -- spd-say '${APP_NAME}'
```

#### <a name='load-env-file-to-current-shell-session'></a>Load Env File To Current Shell Session

If you want to make the variables defined in a env file to your running shell session use something like the following snippet.


```sh
$ eval $(envset development)
```

#### <a name='required-environment-variables'></a>Required Environment Variables

You can specify a list of required environment variables for your command using the `--required` flag or its `-R` alias.

Given the following env file:

```ini
[development]
APP_SECRET_TOKEN={{APP_SECRET_TOKEN}}
APP_REMOTE_SERVICE_KEY={{APP_REMOTE_SERVICE_KEY}}
```

If you run the following command:

```console
$ envset development --required=BOOM -R BOOM2 -- node index.js
```

`envset` will exit with an error and a message with the missing variables:

```console
missing required keys: BOOM,BOOM2
```

### <a name='generating-an-example-template'></a>Generating An Example Template

If we run the `envset template` command with the previous **.envset** file we generate a **envset.example** file:

```ini
[development]
APP_SECRET_TOKEN={{APP_SECRET_TOKEN}}
APP_REMOTE_SERVICE_KEY={{APP_REMOTE_SERVICE_KEY}}

[production]
APP_SECRET_TOKEN={{APP_SECRET_TOKEN}}
APP_REMOTE_SERVICE_KEY={{APP_REMOTE_SERVICE_KEY}}
```


### <a name='support-for-envfiles'></a>Support For .env Files

You can load other environment files like `.env` files:

```console
$ envset --env-file=.env -- node index.js
```

### <a name='metadata'></a>Metadata

The `metadata` command will generate a JSON file capturing the values of the provided env file.

### <a name='metadata-compara'></a>Metadata Compare

Note that `envset metadata compare` will output to **stderr** in the case that both files do not match.

```console
$ envset metadata compare --section=development .meta/data.json .meta/prod.data.json
```

You can omit the path to the local source metadata file and only pass the remote file you want to compare against, it will use the configured path:

```
$ envset metadata compare --section=development .meta/prod.data.json
```

Pretty output

```console
•  source: .meta/data.json
   STATUS       ENV KEY         HASH
👻 Missing      MY_APP_NAME     6d22b97ab7dd...


• target: .meta/env.staging.json
👍 target has no extra environment variables

•  different values
   STATUS       ENV KEY         HASH
❓ Different    APP_ENV         2e9975854897...
❓ Different    NEW_THING       8896f09440c1...


👻 Missing in source (1) | 🌱 Missing in target (1)

❓ Different values (2)  | 🤷 Ignored Keys (0)
```

To have JSON output you can pass the `--json` flag:

```console
$ envset metadata compare --json -s development .meta/data.json .meta/prod.json
{
  "name": "development",
  "values": [
    {
      "key": "MY_APP_SECRET",
      "hash": "aca50d5cf2f285a5a5c5469c8fe9df2540b9bea6905a23461b",
      "comment": "different hash value"
    },
    {
      "key": "MY_APP_NAME",
      "hash": "6d22b97ab7dd929f1b30099dcacd3a8f883373cefbe4a59a05",
      "comment": "missing in source"
    }
  ]
}
```

#### <a name='ignore-variables'></a>Ignore Variables

When comparing metadata files you can optionally ignore some variables that you know will be different or will be missing. You can do pass `--ignore` or `-I` flag with the variable name:

```console
$ envset metadata compare --section=development -I IGNORED_VAR .meta/prod.data.json
```

## <a name='installation'></a>Installation

### <a name='macos'></a>macOS
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


### <a name='ubuntu-debianx86-64-amd64'></a>Ubuntu/Debian x86_64 - amd64

```console
$ export tag=<version>
$ cd /tmp
$ wget https://github.com/goliatone/go-envset/releases/download/v${tag}/envset_${tag}_linux_x86_64.deb
$ sudo dpkg -i envset_${tag}_linux_x86_64.deb
```

### <a name='centos-redhat-x86-64-amd64'></a>CentOS/Redhat x86_64 - amd64

```console
$ yum localinstall https://github.com/goliatone/go-envset/releases/download/v<version>/envset_<version>_linux_x86_64.rpm
```

### <a name='manual-install-x86-64-amd64'></a>Manual Install x86_64 - amd64

```console
$ wget https://github.com/goliatone/go-envset/releases/download/v<version>/envset_<version>_linux_x86_64.tar.gz
$ tar -C /usr/bin/ -xzvf envset_<version>_linux_x86_64.tar.gz envset
$ chmod +x /usr/bin/envset
```

## <a name='documentation'></a>Documentation

`envset` will look for a file defining different environments and make them available as commands.

```ini
[development]
APP_BASE_URL=https://localhost:3003

[production]
APP_BASE_URL=https://envset.sh
```

Your `.envset` files can have global variables that you are not part of any section but can be used to do string interpolation in section's variable values.

The syntax to interpolate is `%(KEY)s`.

```ini
VERSION=v0.6.2

[development]
APP_BASE_URL=https://localhost:3003
DOWNLOAD_URL=%(APP_BASE_URL)s/%(VERSION)s

[production]
APP_BASE_URL=https://envset.sh
DOWNLOAD_URL=%(APP_BASE_URL)s/%(VERSION)s
```

You can have a special section with comments and the content will not generate syntax errors.


### <a name='Commands'></a>Commands

The following is a list of the available commands:

* metadata
    * compare
* template

### <a name='VariableExpansion'></a>Variable Expansion

`envset` can interpolate variables using POSIX variable expansion in both the loaded environment file and the running command arguments.

```ini
[development]
CLIENT_NAME=$(whoami -f)
CLIENT_ID=${CLIENT_NAME}.devices.local
```

```
$ envset development -- node cli.js --user '${USER}'
```

### <a name='Commands-1'></a>Commands

If you type `envset` without arguments it will display help and a list of supported environment names.

## <a name='envset-file'></a>.envset File


## <a name='envsetrc'></a>.envsetrc
You can create an `.envsetrc` file with configuration options for `envset`.

The default `.envsetrc` looks like this:

```ini
# Default configuration
filename=.envset
expand=true
isolated=true
export_environment=APP_ENV
restart=true
max_restarts=3
restart_forever=false

[metadata]
dir=.meta
file=data.json
print=true
json=false

[template]
dir=.
file=envset.example

[environments]
name=test
name=staging
name=production
name=development

[comments]
key=COMMENTS
key=DOCUMENTATION
```

### <a name='Configuration'></a>Configuration

Follows `rc` [standards][rcstand].


### <a name='ConfigurationSyntax'></a>Configuration Syntax

The loaded files need to be valid `ini` syntax.

```ini
[development]
APP_BASE_URL=https://localhost:3003

[production]
APP_BASE_URL=https://envset.sh
```

### <a name='IgnoredAndRequiredSections'></a>Ignored And Required Sections

You can add a `[required]` or `[ignored]` section in your `.envsetrc`:

```ini
[ignored]
development=MY_APP_NAME
development=MY_APP_SECRET
staging=MY_IGNORED_STAGING

[required]
development=MY_REQUIRED_APP_NAME
staging=MY_REQUIRED_VAR_STAGING
```

## <a name='license'></a>License
Copyright (c) 2015 goliatone
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
