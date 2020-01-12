# Vantage Events Server

Vantage Events Server distributes events from Vantage to clients via websockets. It is designed for maintaining connections with a high number of clients, in both server-to-server and server-to-browser scenarios.

The live state is sent to each new subscriber when the websocket starts, before real-time events are broadcast. This allows clients to restore the live state by opening a websocket connection.

## Event Aggregator

The Event Aggregator component runs on the server. For each websocket client, this component subscribes to NATS Streaming Server and follows competitions and live events within a competition.

### Connect

Clients can connect to the Event Aggregator using the following endpoints:

- `/v1/competitions`: stream with `CompetitionActivatedEvent` of the last 24 hours. This allows clients to present a competition selector screen.
- `/v1/competitions/{id}`: stream with competition events from the specified competition ID.

You can use [wscat](https://github.com/websockets/wscat) to connect to the Event Aggregator. For example:

```bash
$ wscat -c wss://events.emandovantage.com/v1/competitions
$ wscat -c wss://events.emandovantage.com/v1/competitions/52d432dc-d6b8-4045-8a4c-e5e5bdfc8b1e
```

## Event Recorder

The Event Recorder is a utility that allows recording and replaying events for development purposes. The Event Recorder connects to the Event Aggregator and stores events in a file with a timestamp. The Event Recorder can then replay the file and send the stored events in real time to subscribers. Optionally, you can specify a speed value to reduce the wait time between events.

### Installation

Make sure you have [Go](https://golang.org/doc/install) installed in your environment.

```bash
$ go install github.com/emando/vantage-events/cmd/eventrecorder
```

### Record Events

To record events:

```bash
$ eventrecorder record --file test.json
```

>Pass `--help` for additional options.

This stores the files in `test.json`.

### Replay Events

To replay events:

```bash
$ eventrecorder replay --file test.json
```

>Pass `--help` for additional options.

This starts a webserver on port 3000 and replays events to websocket subscribers.

You can subscribe to the replayed events:

```bash
$ wscat -c ws://localhost:3000
```

This stream has the same format as the stream produced by the Event Aggregator.

>Note: You can increase the speed by passing `--speed` to the `replay` command, i.e.:
```bash
$ eventrecorder replay --file test.json --speed 4
```

## Legal

Copyright © 2020 Emando B.V.
