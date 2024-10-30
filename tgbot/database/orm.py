from typing import Generic, List, Type, TypeVar
from sqlalchemy import and_, desc, select
from sqlalchemy.orm import joinedload, sessionmaker
from sqlalchemy.exc import NoResultFound
from tgbot.database.database import Base

from .models import Address, User

T = TypeVar("T", bound=Base)


class CRUDBase(Generic[T]):
    def __init__(self, model: Type[T], session_factory: sessionmaker):
        self.model = model
        self.session_factory = session_factory

    async def create(self, **kwargs) -> T:
        async with self.session_factory() as session:
            obj = self.model(**kwargs)
            session.add(obj)
            await session.commit()
            await session.refresh(obj)
            return obj

    async def get(self, id: int) -> T:
        async with self.session_factory() as session:
            query = select(self.model).filter_by(id=id)
            result = await session.execute(query)
            try:
                return result.scalars().one()
            except NoResultFound:
                return None

    async def get_all(self, **kwargs) -> List[T]:
        async with self.session_factory() as session:
            filters = []
            for key, value in kwargs.items():
                if isinstance(value, (tuple, list)):
                    filters.append(getattr(self.model, key).in_(value))
                else:
                    filters.append(getattr(self.model, key) == value)

            query = (
                select(self.model).filter(and_(*filters)).order_by(desc(self.model.id))
            )
            result = await session.execute(query)
            return result.scalars().all()

    async def update(self, id: int, **kwargs) -> T | None:
        async with self.session_factory() as session:
            query = select(self.model).filter_by(id=id)
            result = await session.execute(query)
            try:
                obj = result.scalars().one()
                for key, value in kwargs.items():
                    setattr(obj, key, value)

                await session.commit()
                await session.refresh(obj)
                return obj
            except NoResultFound:
                return None

    async def delete(self, id: int) -> bool:
        async with self.session_factory() as session:
            query = select(self.model).filter_by(id=id)
            result = await session.execute(query)
            try:
                obj = result.scalars().one()
                await session.delete(obj)
                await session.commit()
                return True
            except NoResultFound:
                return False

    async def count(self, **kwargs) -> int:
        async with self.session_factory() as session:
            query = select(self.model).filter_by(**kwargs)
            result = await session.execute(query)
            return len(result.scalars().all())


class UsersRepo(CRUDBase[User]):
    def __init__(self, session):
        super().__init__(User, session)

    async def get(self, id: int) -> User:
        async with self.session_factory() as session:
            query = (
                select(self.model).filter_by(id=id).options(joinedload(User.addresses))
            )
            result = await session.execute(query)
            unique_result = result.unique()

            try:
                return unique_result.scalars().one()
            except NoResultFound:
                return None


class AsyncORM:
    session_factory: sessionmaker

    # models
    users: UsersRepo
    addresses: CRUDBase[Address]

    @classmethod
    def set_session_factory(cls, session_factory):
        cls.session_factory = session_factory

    @classmethod
    def init_models(cls):
        cls.users = UsersRepo(cls.session_factory)
        cls.addresses = CRUDBase(Address, cls.session_factory)
