from aiogram.types import ReplyKeyboardMarkup, KeyboardButton


main_menu = ReplyKeyboardMarkup(
    resize_keyboard=True,
    keyboard=[
        [
            KeyboardButton(text="➕New address"),
            KeyboardButton(text="⚙️Manage addresses"),
        ],
        [
            KeyboardButton(text="👤Profile"),
            KeyboardButton(text="❗️Info"),
        ],
    ],
)
