# Configuration Options

The application gets configured by a set of configuration files. `001_default.yaml` contains the main configuration, administration commands will create the other configurations based on the `001_default.yaml` configuration.

## Databases

It is recommended to use a redis database. You need two different databases, i.e. for the application and the metrics.
The `metrics` key contains the database configuration for the metrics and the `database` key contains the database configuration for the application.

### Redis

You can configure the application to use a standalone redis instance. 

```yaml
name: db # or meter
type: redis
settings:
    addresses: [ "localhost:6379" ] # Set of addresses pointing to the SAME redis server.
    database: 1 # or 0
    password: "" # Redis password
    master: "mymaster" # Redis master name
```

### Redis + Sentinel

You can configure the application to use a redis + sentinel instance. 

```yaml
name: db # or meter
type: redis
settings:
    sentinel_addresses: [ "localhost:26379" ] # Set of addresses pointing to the SAME sentinel server.
    database: 1 # or 0
    password: "" # Redis password
    master: "mymaster" # Redis master name
    sentinel_username: "username" # Sentinel username
    sentinel_password: "password" # Sentinel password
```


### Redis Sharded

You can also use the application-based Redis sharding by specifing multiple shards. You can use Redis standalone or Redis + Sentinel connections. The application will automagically shard the Redis keys to the different Redis instances. Make sure, that the configuration order is consistent, otherwise the database gets mixed up.

```yaml
name: db # or meter
type: redis
settings:
    shards:
      - addresses: [ "localhost:6379" ] # Set of addresses pointing to the SAME redis server.
        database: 1 # or 0
        password: "" # Redis password
        master: "mymaster" # Redis master name 
      - sentinel_addresses: [ "localhost:26379" ] # Set of addresses pointing to the SAME sentinel server.
        database: 1 # or 0
        password: "" # Redis password
        master: "mymaster" # Redis master name
        sentinel_username: "username" # Sentinel username
        sentinel_password: "password" # Sentinel password
```