from typing import Callable, Dict, Any, Awaitable

from aiogram import BaseMiddleware
from aiogram.types import TelegramObject


class ConfigMiddleware(BaseMiddleware):
    def __init__(self, config, redis) -> None:
        self.config = config
        self.redis = redis

    async def __call__(
        self,
        handler: Callable[[TelegramObject, Dict[str, Any]], Awaitable[Any]],
        event: TelegramObject,
        data: Dict[str, Any],
    ) -> Any:
        data["config"] = self.config
        data["redis"] = self.redis

        result = await handler(event, data)
        return result
