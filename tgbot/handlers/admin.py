import asyncio

from aiogram import F, Router
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.types import CallbackQuery, InaccessibleMessage, Message

from tgbot.database.orm import AsyncORM
from tgbot.filters.admin import AdminFilter
from tgbot.keyboards.inline import (
    admin_menu,
    back_admin,
    choose_menu,
)
from tgbot.misc.states import (
    BroadcastState,
)
from tgbot.services.broadcaster import broadcast

admin_router = Router()
admin_router.message.filter(AdminFilter())


@admin_router.callback_query(F.data == "back_admin")
@admin_router.message(Command("admin"))
async def admin_start(update: Message | CallbackQuery, state: FSMContext):
    if isinstance(update, Message):
        await update.answer("Админ-меню", reply_markup=admin_menu)
        return

    if update.message and not isinstance(update.message, InaccessibleMessage):
        await state.clear()
        await update.message.edit_text("Админ-меню", reply_markup=admin_menu)


@admin_router.message(Command("wipe_lines"))
async def wipe_all_lines_from_db(message: Message):
    await AsyncORM.lines.del_all()
    await message.answer("Все строки были успешно удалены.")


# ======================================================================================================================
# Broadcast
@admin_router.callback_query(F.data == "broadcast")
async def broadcast_main(call: CallbackQuery, state: FSMContext):
    if not call.message or isinstance(call.message, InaccessibleMessage):
        await call.answer("Произошла ошибка!")
        return

    msg_to_edit = await call.message.edit_text(
        "<b>Отправьте фото с текстом для рассылки\n└─❕Можно просто текст</b>",
        reply_markup=back_admin,
    )
    await state.set_state(BroadcastState.BS1)
    await state.update_data(msg_to_edit=msg_to_edit.message_id)


@admin_router.message(BroadcastState.BS1, F.content_type.in_({"text", "photo"}))
async def receive_broadcast_data(message: Message, state: FSMContext):
    data = await state.get_data()
    msg_to_edit: int = data["msg_to_edit"]
    await message.delete()
    if message.photo:
        file_id = message.photo[-1].file_id
        await state.update_data(photo=file_id, text=message.caption)
        await asyncio.sleep(2)
        await message.bot.delete_message(message.chat.id, msg_to_edit)
        await message.answer_photo(
            photo=file_id,
            caption=f"{message.caption}\n\n" f"<b>Все правильно? Отправляем?</b>",
            reply_markup=choose_menu,
        )
    else:
        if not message.text:
            await message.bot.edit_message_text(
                chat_id=message.chat.id,
                message_id=msg_to_edit,
                text="Ни текст, ни фотография не были обнаружены в вашем сообщении, попробуйте еще раз.",
                reply_markup=choose_menu,
            )
            return

        await state.update_data(text=message.text)
        await message.bot.edit_message_text(
            chat_id=message.chat.id,
            message_id=msg_to_edit,
            text=message.text + "\n\n<b>Все правильно? Отправляем?</b>",
            reply_markup=choose_menu,
        )
    await state.set_state(BroadcastState.BS2)


@admin_router.callback_query(BroadcastState.BS2, F.data == "yes")
async def agree_and_start(call: CallbackQuery, state: FSMContext, bot):
    if not call.message or isinstance(call.message, InaccessibleMessage):
        await call.answer("Произошла ошибка!")
        return

    data = await state.get_data()
    text, photo_name, silent_mode = (
        data.get("text"),
        data.get("photo"),
        data.get("silent_mode"),
    )
    await state.clear()
    await call.message.delete()
    users = await AsyncORM.users.get_all()
    to_delete: Message = await call.message.answer("<b>Рассылка начата</b>")
    done = await broadcast(
        bot, users, text, photo_name, disable_notification=silent_mode
    )
    await to_delete.delete()
    await call.message.answer(
        f"<b>Рассылка закончена</b>\n" f"Получили сообщение: <code>{done}</>\n",
        reply_markup=back_admin,
    )
