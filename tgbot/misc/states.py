from aiogram.fsm.state import StatesGroup, State


class BroadcastState(StatesGroup):
    BS1 = State()
    BS2 = State()


class AddNewAddress(StatesGroup):
    receive_value = State()
