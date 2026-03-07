# Contributing

Tapestry lives in the `tapestry` rig within Gas Town.

## Development

```bash
go test ./...
go build ./cmd/tapestry
```

## Deployment

Build the binary, serve via HTTP, then on the target host:

```bash
wget -O /usr/local/bin/tapestry http://<build-host>:8000/tapestry
chmod +x /usr/local/bin/tapestry
systemctl restart tapestry
```
