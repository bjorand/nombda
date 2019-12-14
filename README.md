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

## Usage
```
NOMBDA_TOKEN=xxx CONFIG_DIR=/nombda/conf.d nombda -listen-addr 0.0.0.0:8080
```
Check that nombda is running:
```
curl localhost:8080/ping
```

Then you can create your first hook file in `CONFIG_DIR`.

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

## License

[MIT License](https://github.com/bjorand/nombda-github-action/blob/master/LICENSE)
