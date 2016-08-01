# Whaleprint

![whaleprint] (https://github.com/mantika/whaleprint/blob/master/blue2.jpg)

Whaleprint allows to manage DAB (distributed application bundle) as service blueprints for docker swarm mode


## Rationale

After playing around with docker experimental [DAB's](https://github.com/docker/docker/blob/master/experimental/docker-stacks-and-bundles.md) we realized
that even though the concept looks promising, the tooling around it was somehow poor and pretty much useless. The only thing that you can do with this as today is 
generate a DAB from a `docker-compose` yml file and then run `docker stack deploy` or `docker deploy` in order to deploy it to your swarm mode cluster and that's pretty much it. 

We immediately started thinking of different ways to enhance the dev & ops experience with this new feature and we came up with some nice ideas that makes this possible.
The main concept behind this project is that we believe service stack deployments (specially in production) should be __transparent__, __reliable__ and above all [__declarative__ and not imperative](https://en.wikipedia.org/wiki/Declarative_programming#Definition).  

Whaleprint makes possible to use your current DAB files as swarm mode blueprints and will show you with __extreme detail__  __exactly__ which and how your services will be deployed/removed. 
At the same time it will also handle service update diffs describing precisely what things will change and what will be their new updated value.

Check it out!:

[![asciicast](https://asciinema.org/a/9a7oq68a9eoilwqwq459xdr3t.png)](https://asciinema.org/a/9a7oq68a9eoilwqwq459xdr3t)


## What other things does whaleprint do?


- Preview and apply your DAB changesets (duh!)
- Extend the current DAB format to support MOAR features. ([#7](https://github.com/mantika/whaleprint/issues/7))
- Manage multiple stacks simultaneously
- Fetch  DAB's from an URL
- Remove and deploy service stacks entirely
- Allow to apply specific service update through the `--target` option
- Outputs relevant computed stack information like Published ports
- Alternatively print complete plan detail instead of changesets only


## How do I use it?

Check this YouTube video to see a demo: https://www.youtube.com/watch?v=nwtJflxY560


## Installing whaleprint

Just download the binary for your platform from the [Releases](https://github.com/mantika/whaleprint/releases) section, put it anywher in your PATH and enjoy!

## Extending DAB

Whaleprint not only supports current DAB format but it also extends it in a backward-compatible way and allows to specify some other properties like 
Replicas and Constraints (more features to come).

Here's an example:

```javascript
{
  "services": {
    "vote": {
      "Image": "docker/example-voting-app-vote@sha256:20faa449b42b5f0797b1b1a3028a2dd7ac0ece00b0d100b19e6dff4d1a0af2e3",
      "EndpointMode": "dnsrr", // Here we set the endpoint mode
      "Constraints": [
        "engine.labels.disk == ssd" // We can also add custom constraints
      ],
      "Replicas": 5, // And set the number of replicas
      "Networks": [
        "fruta"
      ]
    }
  },
  "version": "0.1"
}
```

As you can see **Replicas**, **Constraints** and **EndpoitMode** are extended features that are not currently supported in the current [DAB specification](https://github.com/docker/docker/blob/master/experimental/docker-stacks-and-bundles.md). Some other features like setting service **PublishedPorts** is also possible.

## FAQ

#### Do I need some custom docker configuration or version for this?

No, it just works out of the box with your current docker installation

#### This is cool, in which docker version/platform does it work?

Whaleprint works in __any__ OS that's currently running docker 1.12 RC 

#### I'm getting connection errors when trying to use whaleprint. 

~~In OSX this might happen because of an [issue](https://github.com/docker/engine-api/pull/320) in engine-api.
In the meantime just set your `DOCKER_HOST` env variable to your unix socket or TCP connection and you should be ok~~. This issue has been merged, it shouldn't be necessary to do this workaround anymore. 

#### What about performance?.

It's designed to show results instantly even with a large amount of services. 


## Side notes

While working on whaleprint we learnt a lot from docker internals as well as the new `swarm mode` and `swarmkit` core principles. We also found some issues
([#1171](https://github.com/docker/swarmkit/issues/1171)) and sent some PR's fixing small bugs ([#320](https://github.com/docker/engine-api/pull/320), [#1207](https://github.com/docker/swarmkit/pull/1207)).

Some of the concepts and ideas behind Whaleprint are inspired by some other products/companies like HashiCorp terraform who we admire for their excellence and ability to build amazing stuff.
