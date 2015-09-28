---
layout: "intro"
page_title: "Running Nomad"
sidebar_current: "getting-started-running"
description: |-
  Learn about the Nomad agent, and the lifecycle of running and stopping.
---

# Running Nomad

Nomad relies on a long running agent on every machine in the cluster.
The agent can run either in server or client mode. Each region must
have at least one server, though a cluster of 3 or 5 servers is recommended.
A single server deployment is _**highly**_ discouraged as data loss is inevitable
in a failure scenario.

All other agents run in client mode. A client is a very lightweight
process that registers the host machine, performs heartbeating, and runs any tasks
that are assigned to it by the servers. The agent must be run on every node that
is part of the cluster so that the servers can assign work to those machines.

## Starting the Agent

For simplicity, we will run a single Nomad agent in development mode. This mode
is used to quickly start an agent that is acting as a client and server to test
job configurations or prototype interactions. It should _**not**_ be used in
production as it does not persist state.

```
vagrant@nomad:~$ sudo nomad agent -dev

==> Starting Nomad agent...
==> Nomad agent configuration:

                 Atlas: <disabled>
                Client: true
             Log Level: DEBUG
                Region: global (DC: dc1)
                Server: true

==> Nomad agent started! Log data will stream in below:

    2015/09/27 23:51:27 [INFO] serf: EventMemberJoin: nomad.global 127.0.0.1
    2015/09/27 23:51:27 [INFO] nomad: starting 4 scheduling worker(s) for [service batch _core]
    2015/09/27 23:51:27 [INFO] client: using alloc directory /tmp/NomadClient599911093
    2015/09/27 23:51:27 [INFO] raft: Node at 127.0.0.1:4647 [Follower] entering Follower state
    2015/09/27 23:51:27 [INFO] nomad: adding server nomad.global (Addr: 127.0.0.1:4647) (DC: dc1)
    2015/09/27 23:51:27 [WARN] fingerprint.network: Ethtool not found, checking /sys/net speed file
    2015/09/27 23:51:28 [WARN] raft: Heartbeat timeout reached, starting election
    2015/09/27 23:51:28 [INFO] raft: Node at 127.0.0.1:4647 [Candidate] entering Candidate state
    2015/09/27 23:51:28 [DEBUG] raft: Votes needed: 1
    2015/09/27 23:51:28 [DEBUG] raft: Vote granted. Tally: 1
    2015/09/27 23:51:28 [INFO] raft: Election won. Tally: 1
    2015/09/27 23:51:28 [INFO] raft: Node at 127.0.0.1:4647 [Leader] entering Leader state
    2015/09/27 23:51:28 [INFO] raft: Disabling EnableSingleNode (bootstrap)
    2015/09/27 23:51:28 [DEBUG] raft: Node 127.0.0.1:4647 updated peer set (2): [127.0.0.1:4647]
    2015/09/27 23:51:28 [INFO] nomad: cluster leadership acquired
    2015/09/27 23:51:29 [DEBUG] client: applied fingerprints [arch cpu host memory storage network]
    2015/09/27 23:51:29 [DEBUG] client: available drivers [docker exec java]
    2015/09/27 23:51:29 [DEBUG] client: node registration complete
    2015/09/27 23:51:29 [DEBUG] client: updated allocations at index 1 (0 allocs)
    2015/09/27 23:51:29 [DEBUG] client: allocs: (added 0) (removed 0) (updated 0) (ignore 0)
    2015/09/27 23:51:29 [DEBUG] client: state updated to ready
```

As you can see, the Nomad agent has started and has output some log
data. From the log data, you can see that our agent is running in both
client and server mode, and has claimed leadership of the cluster.
Additionally, the local client has been registered and marked as ready.

-> **Note:** Typically any agent running in client mode must be run with root level
privilege. Nomad makes use of operating system primitives for resource isolation
which require elevated permissions. The agent will function as non-root, but
certain task drivers will not be available.

## Cluster Nodes

If you run [`nomad node-status`](/docs/commands/node-status.html) in another terminal, you
can see the registered nodes of the Nomad cluster:

```text
$ vagrant ssh
...

$ nomad node-status
ID                                    DC   Name   Class   Drain  Status
72d3af97-144f-1e5f-94e5-df1516fe4add  dc1  nomad  <none>  false  ready
```

The output shows our Node ID, which is randomly generated UUID,
it's datacenter, node name, node class, drain mode and current status.
We can see that our node is in the ready state, and task draining is
currently off.

The agent is also running in server mode, which means it is part of
the [gossip protocol](/docs/internals/gossip.html) used to connect all
the server instances together. We can view the members of the gossip
ring using the [`server-members`](/docs/commands/server-members.html) command:

```text
$ nomad server-members
Name          Addr       Port  Status  Proto  Build     DC   Region
nomad.global  127.0.0.1  4648  alive   2      0.1.0dev  dc1  global
```

The output shows our own agent, the address it is running on, its
health state, some version information, and the datacenter and region.
Additional metadata can be viewed by providing the `-detailed` flag.

## <a name="stopping"></a>Stopping the Agent

You can use `Ctrl-C` (the interrupt signal) to halt the agent.
By default, all signals will cause the agent to forcefully shutdown.
The agent [can be configured](/docs/agent/config.html) to gracefully
leave on either the interrupt or terminate signals.

After interrupting the agent, you should see it leave the cluster
and shut down:

```
^C==> Caught signal: interrupt
    [DEBUG] http: Shutting down http server
    [INFO] agent: requesting shutdown
    [INFO] client: shutting down
    [INFO] nomad: shutting down server
    [WARN] serf: Shutdown without a Leave
    [INFO] agent: shutdown complete
```

By gracefully leaving, Nomad clients update their status to prevent
futher tasks from being scheduled and to start migrating any tasks that are
already assigned. Nomad servers notifies other their peers they intend to leave.
When a server leaves, replication to that server stops. If a server fails,
replication continues to be attempted until the node recovers. Nomad will
automatically try to reconnect to _failed_ nodes, allowing it to recover from
certain network conditions, while _left_ nodes are no longer contacted.

If an agent is operating as a server, a graceful leave is important to avoid
causing a potential availability outage affecting the
[consensus protocol](/docs/internals/consensus.html). If a server does
forcefully exit and will not be returning into service, the
[`server-force-leave` command](/docs/commands/server-force-leave.html) should
be used to force the server from a _failed_ to a _left_ state.

## Next Steps

The development Nomad agent is up and running. Let's try to [run a job](jobs.html)!