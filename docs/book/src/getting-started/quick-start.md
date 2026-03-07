# Quick Start

1. Install Tapestry (see [Installation](installation.md))
2. Configure your Dolt connection in `~/.config/tapestry/config.toml`:

```toml
[dolt]
host = "dolt.lan"
port = 3306
user = "root"
database = "beads_aegis"
```

3. Start the server:

```bash
tapestry serve --addr :8080
```

4. Open `http://localhost:8080` to see the command center.
