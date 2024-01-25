# Telarr

Telarr is a simple and easy to use bot for telegram designed to interact with *Radarr* and *Sonarr*.

## How to use

1. Create a bot with the [BotFather](https://t.me/botfather)

2. Clone the repository

```bash
git clone https://github.com/AlexT59/Telarr.git
```

3. Edit the configuration file and rename it as `config.yaml`

4. Edit the docker-compose file to fit your needs

5. Run the container

```bash
docker-compose up -d
```

## Authorizations files

The authorizations files are used to allow or deny users to use the bot.
Two files are used, one for whitelisting (`autorized.json`) and one for blacklisting (`blacklist.json`).
Both files are formatted as follow:

```json
[
    {
        "id": 123456789,
        "name": "username"
    }
]
```
