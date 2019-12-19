# nombda

> Nombda is no lambda.

## What is nombda?

Nombda is a Go webservice which allows you to trigger with a POST an interactive script.
Use cases are:
 - trigger a container upgrade from your CI and rollback if it fails
 - or any remote shell operations

 A tiny DSL is used to describe actions triggered.

## Installation

```
go get -v github.com/bjorand/nombda
```

## Running nombda
```
NOMBDA_TOKEN=xxx CONFIG_DIR=/nombda/conf.d nombda -listen-addr 0.0.0.0:8080
```
Check that nombda is running:
```
curl localhost:8080/ping
```

## Create your first hook

You can create your first hook file in `CONFIG_DIR`.

Let's say we have a static website served by Nginx from a local Git repository.

We want to pull a repository when we trigger a call to nombda, then reload nginx and rollback if something fails.

Create the hook directoy where you will create your first action file:
```
cd /nombda/conf.d
mkdir mywebsite
```

Create an action file `git_update.yml`:

```
handlers:
  rollback:
    - name: is rollback possible
      command: git rev-parse --short $(cat /var/www/.website_rollback_sha1)
      cd: /var/www/website
    - handler: reload_nginx
    - handler: healthcheck

  reload_nginx:
    - name: reload nginx
      command: nginx -s reload

  healthcheck:
    - name: healthcheck container
      command: curl -fvs localhost/ping
      retry: 3
      interval: 3
steps:
  - name: pulling git repository
    command: git pull
    cd: /var/www/website

  - handler: reload_nginx
    on_failure: rollback

  - handler: healthcheck
    on_failure: rollback

  - name: save commit for potential rollback
    command: git rev-parse --short HEAD > /var/www/.website_rollback_sha1
```

Finally you can trigger your action `git_update` in hook `mywebsite` with curl:

```
curl -XPOST -H"Auth-token=xxx" localhost:8080/mywebsite/git_update
```

## Integrations

- Nombda can be triggered by a Github Action: https://github.com/marketplace/actions/nombda-hook

## DSL documentation

Hooks are described with a simple yaml syntax which aims to help writing complex operations.

### Grammar

At root level, you can define:
- `tasks`
- `vars`
- `handlers`

Here is a complete example:

```
vars:
  dir: /var/www/myapp

handlers:
  update:
    - name: pull code
      command: git pull
      cd: ${var.dir}

tasks:
  - handler: update

  - name: install deps
    command: bundle install
    cd: ${var.dir}
```

Same hook can be rewritten as:
```
vars:
  dir: /var/www/myapp

handlers:
  update:
    - name: pull code
      command: git pull
      cd: ${var.dir}

    - handler: bundle_install

  bundle_install:
    - name: install dep
      command: bundle install
      cd: ${var.dir}

tasks:
  - handler: update
```


#### `command` module attributes

- `command` string: run this command
- `only_if` string: run command specified in string and run task `command` attribute only if return code is 0.
- `register` string: save `command` output in specified variable name
- `cd` string: change directory for running `command`
- `on_failure` string: if `command` fails, run the specified handler listed in root `handlers`
- `continue_after_failure` bool: continue to next task if `command` fails.


#### `handler` module attributes

- `handler` string: run the specified handler listed in root `handlers`
- `vars` map of string: call `handlers` with defined variables
- `on_failure` string: call specified handler if `handler` fails

### Using `vars` and `register`

`vars` at root level define global variables:

```
vars:
  dir: /tmp
  tmpfile: foo

tasks:
  - name: create tmp file
    command: touch ${var.dir}/foo
  - name: delete tmp file
    command: rm ${var.tmpfile}
    cd: ${var.dir}
```

`vars` attribute can also be used to call handlers:

```
handlers:
  create:
    - name: create tmp file
      command: touch ${var.tmpfile}

tasks:
  - handler: create
    vars:
      tmpfile: /tmp/foo
```
It overrides global variable of the same name if it exists.

`register` attribute save value in a global variable:

```
handlers:
  notify_deployment:
    - name: post event to Ops dashboard
      command: curl -XPOST -d'{"revision": "${var.git_sha}"}' http://...

tasks:
  - name: save git sha
    command: git rev-parse --short HEAD
    register: git_sha
  - handler: notify_deployment
```

### Interpolation

```
tasks:
  - name: set var
    command: echo 1
    register: a
  - name: use var
    command: echo ${var.a}
```
Command raw `echo ${var.a}` string will be evaluated and replace by:
`echo 1` for execution.

`{var.NAME}` can interpolate variables created with keyword `vars` or `register`.

## License

[MIT License](https://github.com/bjorand/nombda/blob/master/LICENSE)
