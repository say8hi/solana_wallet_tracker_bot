from dataclasses import dataclass
from typing import Optional

from environs import Env


@dataclass
class TgBot:
    token: str
    admin_ids: list[int]

    @staticmethod
    def from_env(env: Env):
        token = env.str("BOT_TOKEN")
        admin_ids = list(map(int, env.list("ADMINS")))
        return TgBot(token=token, admin_ids=admin_ids)


@dataclass
class Postgres:
    db_name: str
    db_user: str
    db_pass: str
    db_host: str

    @staticmethod
    def from_env(env: Env):
        db_name = env.str("POSTGRES_DB")
        db_user = env.str("POSTGRES_USER")
        db_pass = env.str("POSTGRES_PASSWORD")
        db_host = env.str("POSTGRES_HOST")
        return Postgres(
            db_name=db_name, db_user=db_user, db_pass=db_pass, db_host=db_host
        )


@dataclass
class Redis:
    redis_pass: Optional[str]
    redis_port: Optional[int]
    redis_host: Optional[str]
    redis_tx_channel: Optional[str]
    redis_cmd_channel: Optional[str]

    def dsn(self, database_num: int = 0) -> str:
        if self.redis_pass:
            return f"redis://:{self.redis_pass}@{self.redis_host}:{self.redis_port}/{database_num}"
        else:
            return f"redis://{self.redis_host}:{self.redis_port}/{database_num}"

    @staticmethod
    def from_env(env: Env):
        redis_pass = env.str("REDIS_PASSWORD")
        redis_port = env.int("REDIS_PORT")
        redis_host = env.str("REDIS_HOST")
        redis_tx_channel = env.str("REDIS_TX_CHANNEL")
        redis_cmd_channel = env.str("REDIS_CMD_CHANNEL")

        return Redis(
            redis_pass=redis_pass,
            redis_port=redis_port,
            redis_host=redis_host,
            redis_cmd_channel=redis_cmd_channel,
            redis_tx_channel=redis_tx_channel,
        )


@dataclass
class Misc:
    dev: Optional[bool]

    @staticmethod
    def from_env(env: Env):
        dev = env.str("IN_DEVELOPMENT")

        return Misc(dev=dev)


@dataclass
class Config:
    tg_bot: TgBot
    postgres: Postgres
    redis: Redis
    misc: Misc


def load_config(path: Optional[str] = None) -> Config:
    env = Env()
    env.read_env(path)

    return Config(
        tg_bot=TgBot.from_env(env),
        postgres=Postgres.from_env(env),
        redis=Redis.from_env(env),
        misc=Misc.from_env(env),
    )
