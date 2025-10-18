Telegram Bot & Login Widget - Step by step

1. Create Bot (BotFather):
   - Open Telegram app or web (https://web.telegram.org/)
   - Search for @BotFather and start chat
   - Send /newbot and follow prompts (name, username)
   - BotFather will return a token: `123456:ABC-DEF...`. Save it to `.env` as TELEGRAM_BOT_TOKEN

2. Get your Telegram user ID (for whitelist):
   - Add @userinfobot or @get_id_bot in Telegram
   - Send /start and it will return your numeric Telegram ID
   - Put this ID into TELEGRAM_WHITELIST in `.env`, e.g. TELEGRAM_WHITELIST=123456789,987654321

3. Telegram Login Widget for web login:
   - Register your site with Telegram: follow https://core.telegram.org/widgets/login
   - On frontend, include the widget script and configure `data-telegram-login` with your bot's username.
   - The widget returns fields: id, first_name, username, photo_url, auth_date, hash
   - Send these fields to the API `/login` endpoint to exchange for a JWT

Security: verify the `hash` on server side using bot token (see Telegram docs). This skeleton has a placeholder check; implement full verification before production.
