# Vantage Events Server

Vantage Events Server distributes events from Vantage to clients via websockets. It is designed for maintaining connections with a high number of clients, in both server-to-server and server-to-browser scenarios.

The live state is sent to each new subscriber when the websocket starts, before real-time events are broadcast. This allows clients to restore the live state by opening a websocket connection.

## Concept

Vantage embraces the concept of event sourcing: state is communicated in a series of (change) events that are published. Subscribers are responsible for restoring and maintaining state by observing events. Vantage supports a recovering mechanism for clients by sending the relevant events when they subscribe, so that they can restore the current state at any point in time.

The key events are sent in the following order:

- `CompetitionActivatedEvent`: activation of a competition
   - `DistanceActivatedEvent`: activation of a distance
      - `HeatActivatedEvent`: activation of a heat (there can be multiple concurrent)
         - `HeatClearedEvent`: clear of a heat (i.e. clock reset)
         - `HeatStartedEvent`: start of a heat
         - `RaceLapPassingAddedEvent`: new passing (transponder loop or lap)
         - `LastRaceSpeedChangedEvent`: speed changed
         - `RaceLapAddedEvent`: new lap
         - `LastPresentedRaceLapChangedEvent`: presented lap changed
         - `HeatDeactivatedEvent`: deactivation of a heat
      - `HeatCommittedEvent`: commit of a heat with result and time (heat doesn't have to be active)
      - `DistanceDeactivatedEvent`: deactivation of a distance

All events are JSON encoded. The type of the event can be found in `typeName`.

### Heats and Races

In long track speed skating, each pair is a heat. All heats are in round 1.

For example, distance 9 pair 5 has heat round 1 and number 5. In pair 5, there are two races with lane inner (0) and outer (1).

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
$ go get github.com/emando/vantage-events/cmd/eventrecorder
```

### Record Events

To record events of competition activations:

```bash
$ eventrecorder record --file activations.json
```

>Pass `--help` for additional options.

This stores activations in `activations.json`.

To record events of a competition:

```bash
$ eventrecorder record --file competition.json --competition 52d432dc-d6b8-4045-8a4c-e5e5bdfc8b1e
```

This stores the competition events in `competition.json`.

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

### Example Events

You can find example events in the `examples` folder. These are recordings from actual events that can be used during development.

## Legal

Copyright © 2020 Emando B.V.
