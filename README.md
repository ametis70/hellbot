# hellbot
hellbot is a [HELLDIVERS™](https://www.arrowheadgamestudios.com/aboutarrowhead/games/helldivers/) [Discord](https://discord.com/) bot

> Scientists Super Earth has developed a new stratagem. The BOT-21 'Hellbot' will broadcast interdimensional messages to gather all the Helldivers willing to participate in the Galactic Campaign.

## What does this bot do

- Send alerts when a capital city defend event starts/ends
- Send alerts when an attack on a faction homeworld starts/end 
- Give information about ongoing events when requested
- Give information about weapons/stratagems when requested

## How to add this bot to X server?

This bot does not have a public instance (yet?), so the only way to add it to a server is through self-hosting. Roughly, these are the steps to do so:

1. Create an app in the [Discord Developer Portal](https://discord.com/developers/applications)  
2. Enable the bot for that app and save the bot token
3. Use the OAuth2 URL generator to create an authorization link with `bot` and `application.commands` permissions
4. Use the generated url to add the app to your server 
5. [Copy the ID for the channel](https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID-) where hellbot will send attack/defend event alerts
6. Run the bot

## How to run the bot

Once you added the bot to your server, you can use [Docker](https://docs.docker.com/get-started/overview/) to run it:

```sh
docker run -e\
  -e TOKEN=<Replace with bot token>\
  -e CHANNEL_ID=<Replace with channel ID>\
  ghcr.io/ametis70/hellbot:latest
```

The image is built for `linux/amd64`, `linux/arm/v7` and `linux/arm64`

### With Docker Compose

Create a `docker-compose.yml` file with the following content:

```yml
version: "2" 
services: 
  hellbot: 
    container_name: hellbot 
    image: ghcr.io/ametis70/hellbot:latest 
    environment: 
      - TOKEN=${TOKEN} 
      - CHANNEL_ID=${CHANNEL_ID} 
    restart: unless-stopped
```

and a `.env` file in the same directory:

```sh
TOKEN=<TOKEN>
CHANNEL_ID=<Replace with channel ID>
```

Then run with:

```sh
docker-compose up -d
```

# Contributing

Bug reports and feature requests are welcome, feel free to use one of the templates to fill an issue.

Pull requests are also welcome. There is no unit testing set up (yet), so just be sure to use [`gofmt`](https://pkg.go.dev/cmd/gofmt) to properly format your code and it will be reviewed.

# Changelog

See [releases](https://github.com/ametis70/hellbot/releases).

(Automatically generated using  [release-it](https://github.com/release-it/release-it) and [auto-changelog](https://github.com/cookpete/auto-changelog))
