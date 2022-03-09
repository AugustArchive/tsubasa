# 🐇 Tsubasa
> *Tiny microservice to define a schema and then be executed by any search engine you wish to use, like Elasticsearch, Meilisearch, or OpenSearch!*

## Why did you build this?
I built this as a way to define multiple schemas from many projects I create that should be reliable
on indexing and searching without doing it myself. This is a HTTP **service** that is like GraphQL but for
search engines like **Elasticsearch**!

A schema is defined using the **TQL** (tsubasa query language):

```tql
schema(name: "some_name") {
  define {
    field("user_id", INT);
  }
}
```

This will generate the following in JSON:
```json
{
  "version": 1,
  "schema": {
    "name": "some_name",
    "engine": "elastic",
    "definitions": [
      {
        "key": "field",
        "name": "user_id",
        "data_type": "INT",
        "extra_metadata": null
      }
    ]
  }
}
```

You can also create custom data types that live on the server, i.e:

```tql
createDataType(name: "SOME_DATA_TYPE", raw_type: STRING)
```

Which will generate the following in JSON:
```json
{
  "version": 1,
  "executor": "createDataType",
  "metadata": {
    "name": "SOME_DATA_TYPE",
    "raw_type": "STRING"
  }
}
```

Now we created a custom data type (optional) and created a schema, now we can test it!

**Tsubasa** will do daily checks on the connection of Elasticsearch, Meilisearch, and OpenSearch
and will report any errors that might have occured when retrieving the status of the engine itself.

If any Elasticsearch or OpenSearch node fails or if Meilisearch cannot be reached, you will always get the following data payload with the **503 Service Unavailable** status code:

```js
{
  "success": false,
  "errors": [
    {
      "code": "ENGINE_UNAVAILABLE",
      "message": "Elasticsearch node(s) [...node list] has failed."
      
      // If we are in development, we can retrieve the stack that
      // Tsubasa has retrieved:
      "response": {
        "status_code": 400,
        "data": {
          [...]
        }
      }
    }
  ]
}
```

Now, to test the schema, we just need to point to `<tsubasa-server>/execute`:

```json
{
  "query": "search(match_type: FUZZINESS, item: \"...\") { use what field is available! }"
}
```

And you should get a result back, if succeeded:

```js
{
  "success": true,
  "data": {
    "took": 50, // in milliseconds
    "data": [
      {
        "user_id": "..."
      }
    ]
  }
}
```

## Installation
Sweet, you want to use **Tsubasa** for your own use cases! You can install the **Tsubasa** server:

- using the [**Noelware Helm Charts**](#helm-chart);
- using the official [Tsubasa Docker Image](#docker-image);
- locally under the main repository [you see](#locally-with-git)

### System Requirements
This is the minimum system requirements to bootstrap **Tsubasa**. You don't need to worry about this if
you're running on the Helm Chart since that handles it for you.

- **2GB** or higher of system RAM
- **2 CPU Cores** or higher

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

...and the server should be running now. The helm chart assumes you installed **etcd** and the search engine
of your choice, we do not wanna bring unnecessary **etcd** instances.

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
