# Daily Reward API
The API behind my DailyReward checker architecture. This simply pings Hypixel checks my profile and returns when the last reward was claimed. If the timestamp doesn't represent the current day it will also send a Discord webhook notification to a private server of mine. <br>
You can use this API for you own profile easily. Clone this project on [vercel](https://vercel.com) and set up the following ENVIRONMENT VARS:
- DISCORD_MESSAGE: The discord message to send (use "<t:%s:D>" to format the last claim date correctly)
- DISCORD_TTS: true or false (send as tts)
- DISCORD_USERNAME: Give any name to the webhook
- DISCORD_AVATAR: Link to an avatar for the webhook
- DISCORD_WEBHOOK: Discord webhook link
- API_KEY: [Hypixel API key](https://developer.hypixel.net/)
- PLAYER_UUID: Your UUID in the "-" format, example: 8f802f1b-b19d-40b5-b36c-8ae614b20fb3
