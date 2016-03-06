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
go run src/mcbot/mcbot.go [--port port] [--config configFile] [--debug cmd]
```

This will start an http server listening for slack command requests on `[port]`.
Using `configFile` as the configuration file (explained below).
To integrate this with slack, add a new Slash command custom integration, and
set the URL to point to a server running this mcbot server.

<a name="server-configuration"></a>
### Server configuration file

The server configuration file is a file containing JSON specifying several
default parameters of the server. Here is an example configuration file. If a
config file is not specified, an *unspecified* default one will be used.

```json
{
    "HiddenByDefault": true,
    "DefaultPort": 25565,
    "DefaultIp": "127.0.0.1",
    "SlackToken": "mySecretSlackToken"
}
```

If `HiddenByDefault` is set, then by default only the user that sent the slack
command will be able to see the output. `SlackToken` will specify the slack
token that needs to be sent for the server to operate. If it's the empty string,
then the token is not checked. **This isn't recommended because it will permit
any team to use your mcbot server.**

### Server debug flag

If you pass the `--debug` flag to the server with a command, instead of starting
an HTTP server listening for Slack requests, it will (once) perform the given
command as though it were a request from a user, dumping the output that would
be sent to slack to standard out. For example, you can run

```
go run src/mcbot/mcbot.go --config myconfig.cfg --debug "--ip localhost --full"
```

to test out your configuration and server.

## Using slash command

Usage of the slash command is simple. In a slack channel on a team with the
integration set up, use the following command

```
/mcbot [--ip ip] [--port port] [--hidden=bool] [--full=bool]
```

Where `ip` and `port` are, respectively, the IP and port of the minecraft
server, and `hidden` controls whether every user in the channel sees the output
of the command. Setting `full` causes a full stat to be performed, which
includes information about exactly which players are currently on. Upon success,
the server will respond with something looking like the following (in the case
of a basic stat)

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
