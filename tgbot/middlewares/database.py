from typing import Callable, Dict, Any, Awaitable

from aiogram import BaseMiddleware
from aiogram.types import CallbackQuery, Message, TelegramObject

from tgbot.database.orm import AsyncORM


class DatabaseMiddleware(BaseMiddleware):
    async def __call__(
        self,
        handler: Callable[[TelegramObject, Dict[str, Any]], Awaitable[Any]],
        event: Message | CallbackQuery,
        data: Dict[str, Any],
    ) -> Any:
        if not event.from_user:
            return

        user = await AsyncORM.users.get(event.from_user.id)
        if not user:
            user = await AsyncORM.users.create(
                id=event.from_user.id,
                username=event.from_user.username,
            )

        if user and event.from_user.username != user.username:
            user = await AsyncORM.users.update(
                user.id, username=event.from_user.username
            )

        data["user"] = user
        result = await handler(event, data)
        return result
