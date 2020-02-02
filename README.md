# PubSub

### Notes for reviewer:

- While in mid-implementation I added topics without remembering it was
  specifically asked to not be implemented. It can be taken out if
  necessary 
- The service can be used with either http or websockets depending if
  the client chooses to upgrade
- `handlers.go` needs refactoring but that was skipped in the interest
  of time
- `./scripts/curl-ws` is mostly to document how to use websockets with
  curl and was never fleshed out to be a full script
- The datastore uses a mutex to maintain consistency right now but for
  performance I would probably move to eventual-consistency in the
  future to increase throughput depending on use-case

## API Endpoints

**Both endpoints gracefully degrade to http. Subscribe will just return
all historic messages in that case.**

*/topic/{topic}/publish*
- If the topic doesn't exist pubsub will create the topic
- Pubsub will create the message in the desired topic

*/topic/{topic}/subscribe*
- Gets all messages in the data store
- If upgraded to websocket connection realtime messages will be streamed
  as well


## Build and Run

```
GO111MODULES=on go build -o pubsub *.go && ./pubsub
```


## Automated tests

```
go test -v *.go
```


# Use

Interact will this application by using `curl` or a websocket client. I
used `websocat` during development

https://github.com/vi/websocat


Install with:

```
# To install make sure the rust ecosystem is installed on your computer.
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Install websocat
cargo install websocat
```

Test Endpoints:
```
# Start publisher
./scripts/websocat -m publish

# In another window start 1 or more subscribers
./scripts/websocat -m subscribe
```

