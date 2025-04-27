This directory contains an example [Docker Compose](https://docs.docker.com/compose/) file to deploy saturn-bot.

## Usage

1.  Copy the [docker-compose.yml](./docker-compose.yml) file to a directory.
1.  Create the configuration file `config.yaml` in the same directory.
    Refer to the [Configuration documentation](https://saturn-bot.readthedocs.io/en/stable/reference/configuration)
    for all available configuration options.
1.  Create the directory `tasks`.
1.  Add a [Task](https://saturn-bot.readthedocs.io/en/stable/reference/task/),
    for example `tasks/hello-world.yaml`.
1.  The directory tree looks like this:

    ```text
    .
    ├── config.yaml
    ├── docker-compose.yml
    └── tasks
        └── hello-world.yaml
    ```

1.  Start the containers: `docker compose up`
1.  Access the UI: `open http://localhost:3035`
