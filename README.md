# sockroom

This is demo service for exploring some infrastructure technologies. It accepts
websocket connections to any path and use it as room. Then it resends all
messages to all subscribers.

## Architecture
Sockroom behind the scene use Local dispatcher for sending messages to
subscribers. It works on `map[string]map[*Subscriber]struct{}` dictionary.

Also sockroom can use [NATS](nats.io) for sync messages between many instances.
