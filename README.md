# mcquery

mcquery is a library implementing part of the query protocol for minecraft
servers paired with a command line utility for performing queries and a simple
http server suitable to serve as the target of a custom slack command
integration.

## Running command line utility

After cloning the repository and setting `GOPATH` to be the root of this repo,
you can run

```
go run src/main/main.go --ip ip --port port
```

which will stat the server at the given ip and port

## Running slack command server

To run the mcbot slackbot server, run

```
go run src/mcbot/mcbot.go [--port port] [--config configFile]
```

This will start an http server listening for slack command requests on `[port]`.
Using `configFile` as the configuration file (explained below).
To integrate this with slack, add a new Slash command custom integration, and
set the URL to point to a server running this mcbot server.

<a name="server-configuration"></a>
### Server configuration file

The server configuration file is a file containing JSON specifying several
default parameters of the server. Here is an example configuration file.

```json
{
    "HiddenByDefault": true,
    "DefaultPort": 25565,
    "DefaultIp": "127.0.0.1"
}
```

If `HiddenByDefault` is set, then by default only the user that sent the slack
command will be able to see the output.

## Using slash command

Usage of the slash command is simple. In a slack channel on a team with the
integration set up, use the following command

```
/mcbot [--ip ip] [--port port] [--hidden=bool]
```

Where `ip` and `port` are, respectively, the IP and port of the minecraft
server, and `hidden` controls whether every user in the channel sees the output
of the command. Upon success, the server will respond with something looking like the
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

If a particular parameter is not given, the effects of the command will depend
on the server default configuration as outlined [above](#server-configuration).


**NOTE** The slash command syntax is likely to change in the future as I add
features / fix things
