# PubSub

## API Endpoints

**Both endpoints gracefully degrade to http. Subscribe will just return
all historic messages in that case.**

*/<topic>/publish*
- If the topic doesn't exist pubsub will create the topic
- Pubsub will create the message in the desired topic

*/<topic>/subscribe*
- Gets all messages in the data store
- If upgraded to websocket connection realtime messages will be streamed
  as well


## Build and Run

```
GO111MODULES=on go build -o pubsub *.go && ./pubsub
```


## Automated tests

```
go test ...
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
./scripts/websocat -t test_topic -m publish

# In another window start 1 or more subscribers
./scripts/websocat -t test_topic -m subscribe
```

