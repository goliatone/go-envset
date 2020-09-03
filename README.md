# envset

`envset` runs another program with a custom environment according to values defined in a configuration file with [ini][ini] file format.

---

## Environment level configuration

Application configuration usually is environment specific and will change between build distributions.

If you follow the [12 factor app][12factor] guidelines, then you store part of your application configuration in the environment. 

Environment variables enable us to manage application configuration outside of our application code.

By application configuration we mean small and sensitive data such as API keys, database credentials. Not all environment configuration are secrets, instead there might be build distribution specific values such as the application's base URL to build OAuth callbacks, logging verbosity or anything that is changes between development and production.

`envset` helps you manage and set environment variables for multiple build distributions.

Is as simple as calling:

```
envset development -- node server.js
```

## Similar Tools

Inspired by [daemontools][dtools]' tool [envdir][envdir] and tools such as [dotenv](https://github.com/bkeepers/dotenv).

* Distributed as a single binary
* No dependencies in your codebase
    * e.g. `dotenv-rails` and `dotenv`[^node-dotenv] for Node.js require you to use a library
* Support multiple environments in a single file
* Generates an example file with your current env vars to keep documentation updated.
* Interpolation of variable using POSIX variable expansion.
* Command expansion

Instead of having an `.env` file per environment you can have one single `.envset` file with one section per environment. 


[^node-dotenv]: You an actually require the library outside of your project with the `node -r` flag.

## Examples

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


To use it, simply prefix the call to your program with `envset` and the name of the environment:

```
$ envset development -- node app.js
```

You can:

```ini
[local]
MSG=Hello World
```

```
envset local -- env | grep MSG | say
```

## Installation

TODO: List how to install in all different platforms


```
$ brew install envset
```

## Documentation

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
;Default environment names
filename=./.envset
exportEnvironment=NODE_ENV

[environments]
names[]=test
names[]=staging
names[]=production
names[]=development
```

### Configuration

Follows `rc` [standards][rcstand].


### Configuration Syntax

The loaded files need to be valid `ini` syntax.

```
```


## License
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
-->