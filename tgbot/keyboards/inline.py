from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton
from aiogram.utils.keyboard import InlineKeyboardBuilder

admin_menu = InlineKeyboardMarkup(
    inline_keyboard=[
        [
            InlineKeyboardButton(text="📬Рассылка", callback_data="broadcast"),
            InlineKeyboardButton(text="🗃Категории", callback_data="categories"),
        ],
        [
            InlineKeyboardButton(
                text="📦Запросы на отработку", callback_data="handle_req:main:0"
            ),
            InlineKeyboardButton(text="📑История", callback_data="history_reqs:main:0"),
        ],
        [
            InlineKeyboardButton(
                text="💳Запросы на вывод", callback_data="withdraw_req:main:0"
            )
        ],
        [InlineKeyboardButton(text="🔎Поиск запроса", callback_data="find_strings")],
        [InlineKeyboardButton(text="✖️Закрыть", callback_data="close")],
    ]
)


def support_menu(lang):
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text="🆘Поддержка" if lang == "ru" else "🆘Support",
                    url="https://t.me/cookierowsupport",
                )
            ],
        ]
    )


def cancel_menu(arg="cancel"):
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text="✖️Cancel" if arg == "cancel" else "✖️Close",
                    callback_data="cancel",
                )
            ]
        ]
    )


back_admin = InlineKeyboardMarkup(
    inline_keyboard=[[InlineKeyboardButton(text="🔙Назад", callback_data="back_admin")]]
)


choose_menu = InlineKeyboardMarkup(
    inline_keyboard=[
        [
            InlineKeyboardButton(text="✔️Да", callback_data="yes"),
        ],
        [InlineKeyboardButton(text="🔙Назад", callback_data="back_admin")],
    ]
)
