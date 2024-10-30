import json
from aiogram import F, Router
from aiogram.filters import CommandStart
from aiogram.fsm.context import FSMContext
from aiogram.types import CallbackQuery, InaccessibleMessage, Message
from redis.asyncio.client import Redis
from base58 import b58decode

from tgbot.database.models import User
from tgbot.database.orm import AsyncORM
from tgbot.keyboards.inline import (
    cancel_menu,
    support_menu,
)
from tgbot.keyboards.reply import main_menu
from tgbot.misc.states import AddNewAddress

user_router = Router()


@user_router.callback_query(F.data == "cancel")
async def cancel_current(call: CallbackQuery, state: FSMContext):
    await state.clear()
    if call.message and not isinstance(call.message, InaccessibleMessage):
        await call.message.delete()
    await call.answer("Canceled")


@user_router.message(CommandStart())
async def user_start(message: Message):
    if not message.from_user:
        return

    await message.answer(
        "Welcome to the Solana Wallet Tracker Bot!",
        reply_markup=main_menu,
    )


@user_router.message(F.text == "❗️Info")
async def support_handler(message: Message, user: User):
    await message.answer(
        "Contact support⤵️",
        reply_markup=support_menu(user.lang),
    )


@user_router.message(F.text == "➕New address")
async def add_new_address(message: Message, state: FSMContext):
    msg = await message.answer(
        "Send me a new addresses",
        reply_markup=cancel_menu(),
    )
    await state.set_state(AddNewAddress.receive_value)
    await state.update_data(edit_msg_id=msg.message_id)


@user_router.message(AddNewAddress.receive_value)
async def add_new_address_receive_value(
    message: Message, state: FSMContext, redis: Redis
):
    if not message.text or not message.bot:
        return

    data = await state.get_data()
    edit_msg_id = data["edit_msg_id"]

    addresses = message.text.split("\n")
    addresses_dict = {}
    for address in addresses:
        sol_address, name = address.split()
        try:
            # b58decode(sol_address)
            if len(sol_address) != 44:
                raise ValueError
        except Exception:
            await message.answer(f"❗️{sol_address}\n└─Wrong address format")
            continue

        addresses_dict[name] = sol_address

        command = {"action": "add", "address": sol_address, "chat_id": message.chat.id}

        await redis.publish("wallet_commands", json.dumps(command))

    await message.delete()
    await message.bot.edit_message_text(
        chat_id=message.chat.id,
        message_id=edit_msg_id,
        text="Successfully added",
        reply_markup=cancel_menu("close"),
    )


@user_router.callback_query(F.data == "personal_acc")
@user_router.message(F.text == "👤Profile")
async def personal_acc_handler(event: Message | CallbackQuery, user: User):
    method_dict = {Message: event.answer}
    if isinstance(event, CallbackQuery):
        method_dict[CallbackQuery] = event.message.edit_text

    await method_dict[type(event)](
        "profile",
    )
