# CraftOps

A command-line tool for running a Minecraft server and keeping its mods up to date.

CraftOps takes care of the repetitive parts of running a modded Minecraft server: starting and stopping it, updating mods from Modrinth, making backups, and sending Discord alerts. Everything is driven by one small config file.

## Features

- Start, stop, restart, and check a server running in GNU screen
- Update all configured Modrinth mods in one command, with retries and a dry-run mode
- Create compressed backups with retention limits and exclude patterns
- Send Discord warnings before restarts and alerts when something goes wrong
- Run health checks for paths, dependencies, and API access
- Install with a one-line script, from source, or as a Docker image

## Usage

You'll need Linux or macOS, GNU screen, and Java 17+ for the server.

```bash
curl -fsSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash

craftops init -o ~/.config/craftops/config.toml
craftops health
craftops server start
```

Edit the generated config to point at your server, mods, and backup folder, then run `craftops --help` to see every command.

To build from source instead, you'll need Go 1.27+:

```bash
make install
```

## Good to know

- Settings live in a single TOML file; `craftops init` creates a sensible starting point.
- Mods are resolved from the Modrinth project pages you list in the config.
- Discord alerts stay off until you add a webhook URL.
- Backups are plain `.tar.gz` archives — store them somewhere safe.

## Built with

Go, Cobra, and the Modrinth API.

## License

MIT
