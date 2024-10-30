from typing import Callable, Dict, Any, Awaitable

from aiogram import BaseMiddleware
from aiogram.types import CallbackQuery, Message, TelegramObject

from tgbot.config import Config


class DeveloperMiddleware(BaseMiddleware):
    async def __call__(
        self,
        handler: Callable[[TelegramObject, Dict[str, Any]], Awaitable[Any]],
        event: Message | CallbackQuery,
        data: Dict[str, Any],
    ) -> Any:
        if not event.from_user:
            return

        config: Config = data["config"]
        if config.misc.dev and event.from_user.id not in config.tg_bot.admin_ids:
            return

        result = await handler(event, data)
        return result
