# mcquery

mcquery is a library implementing part of the query protocol for minecraft
servers paired with a command line utility for performing queries and a simple
http server suitable to serve as the target of a custom slack command
integration.

## Running command line utility

After cloning the repository and setting `GOPATH` to be the root of this repo,
you can run

```
go run src/main/main.go --ip [ip] --port [port]
```

which will stat the server at the given ip and port

## Running slack command server

To run the mcbot slackbot server, run

```
go run src/mcbot/mcbot.go --port [port]
```

This will start an http server listening for slack command requests on `[port]`.
To integrate this with slack, add a new Slash command custom integration, and
set the URL to point to a server running this mcbot server.

## Using slash command

Usage of the slash command is simple. In a slack channel on a team with the
integration set up, simply type

```
/mcbot [ip] [port]
```

Where `[ip]` and `[port]` are, respectively, the IP and port of the minecraft
server. Upon success, the server will respond with something looking like the
following

```
MOTD: A Minecraft Server
Gametype: SMP
Map: world
NumPlayers: 0
MaxPlayers: 20
HostPort: 25565
HostIp: 127.0.0.1
```

**NOTE** The slash command syntax is likely to change in the future as I add
features / fix things
