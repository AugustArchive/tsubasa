# 🐇 Tsubasa
> *Tiny, and simple Elasticsearch microservice to abstract searching objects!*

## Why did you build this?
**Tsubasa** was built to be a simple abstraction to not use the official SDKs to search objects within an
Elasticsearch index, so this is just a simple way to retrieve data from it.

## Installation
> :warning: **Tsubasa is alpha software, it is NOT production ready! Beware~ >!<**

Sweet, you want to use **Tsubasa** for your own use cases! You can install the **Tsubasa** server:

- using the [**Noelware Helm Charts**](#helm-chart);
- using the official [**Tsubasa Docker Image**](#docker-image);
- locally under the main repository [you see](#locally-with-git)

### System Requirements
This is the minimum system requirements to bootstrap **Tsubasa**. You don't need to worry about this if
you're running on the Helm Chart since that handles it for you.

- **2GB** or higher of system RAM
- **2 CPU Cores** or higher
- **Go** 1.17 or higher
- An instance of **Elasticsearch** running. This supports single-node and multi-node instances and multiple
  authentication methods.

### Helm Chart
You can install **Tsubasa** on your Kubernetes cluster with a single command! You will need to index the
**Noelware** Helm Charts under the "noel" user (which is me! :D)

> :warning: **You are required to be using Kubernetes >=1.22 and Helm 3!**

```sh
$ helm repo add noel https://charts.noelware.org/~/noel
```

You should be able to search the **tsubasa** repository when using the **helm search** command.

Now, you should be able to just run **tsubasa** on a single command:

```sh
$ helm install <my-release> noel/tsubasa
```

...and the server should be running now. The helm chart assumes you installed an Elasticsearch cluster installed.

### Docker Image
You can use the official Docker images on [ghcr.io](https://github.com/auguwu/tsubasa/pkgs/containers/tsubasa) or on [Docker Hub](https://hub.docker.com/r/auguwu/tsubasa)!

~ ; ... coming soon >o< ... ; ~

### Locally with Git
~ ; ... coming soon >o< ... ; ~

## Configuraton
**Tsubasa** is configured using a TOML file which must be in the following locations:

- `/app/noel/tsubasa/config.toml` if using the **Docker Image**
- `$ROOT/config.toml` if running locally or wanting to contribute to **Tsubasa**
- `TSUBASA_CONFIG_FILE` environment variable, which will override both clauses above.

You can find an example in the [documentation](https://docs.floofy.dev/services/tsubasa/configuration)!

## Contributing
~ ; ... coming soon >o< ... ; ~

## License
**Tsubasa** is released under the **Apache 2.0** License by [Noel](https://floofy.dev)!
