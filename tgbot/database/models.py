import datetime
from typing import Annotated
from sqlalchemy import BigInteger, ForeignKey, text
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .database import Base

intpk = Annotated[int, mapped_column(primary_key=True)]
created_at = Annotated[
    datetime.datetime, mapped_column(server_default=text("TIMEZONE('utc', now())"))
]


class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True)
    username: Mapped[str] = mapped_column(unique=True)
    balance: Mapped[float] = mapped_column(default=0)
    registered_at: Mapped[created_at]

    addresses: Mapped[list["Address"]] = relationship(
        back_populates="user",
        primaryjoin="User.id == Address.user_id",
        order_by="Address.id.desc()",
    )


class Address(Base):
    __tablename__ = "addresses"

    id: Mapped[intpk]
    user_id: Mapped[int] = mapped_column(
        BigInteger, ForeignKey("users.id", ondelete="CASCADE")
    )
    sol_address: Mapped[str]
    name: Mapped[str]
    active: Mapped[bool]
    created_at: Mapped[created_at]

    user: Mapped["User"] = relationship(
        back_populates="addresses",
    )
