import os
from alembic import command
from alembic.config import Config
from alembic.runtime.migration import MigrationContext
from betterlogging import logging
from sqlalchemy.ext.asyncio import AsyncEngine, create_async_engine
from alembic.script import ScriptDirectory
from sqlalchemy import text
from typing import Optional, Tuple
from urllib.parse import urlparse

from sqlalchemy.ext.asyncio.session import async_sessionmaker

from tgbot.database.orm import AsyncORM


class DatabaseConfig:
    def __init__(self, database_url: str):
        """
        Initialize database configuration from URL
        Example URL: postgresql+asyncpg://user:pass@localhost:5432/dbname
        """
        self.async_url = database_url

        # Parse the URL
        parsed = urlparse(database_url)

        # Store components
        self.user = parsed.username
        self.password = parsed.password
        self.host = parsed.hostname
        self.port = parsed.port or 5432
        self.database = parsed.path.lstrip("/")

        # Generate sync URL for migrations (using psycopg)
        self.sync_url = (
            f"postgresql+psycopg://{self.user}:{self.password}"
            f"@{self.host}:{self.port}/{self.database}"
        )


async def test_connection(url: str) -> bool:
    """Test database connection"""
    from sqlalchemy.ext.asyncio import create_async_engine

    engine = create_async_engine(url)
    try:
        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))
            return True
    except Exception as e:
        logging.error(f"Connection test failed: {e}")
        return False
    finally:
        await engine.dispose()


async def wait_for_database(url: str, timeout: int = 30, interval: int = 1) -> bool:
    """Wait for database to become available"""
    import asyncio

    logging.info(f"Waiting for database connection... (timeout: {timeout}s)")

    start_time = asyncio.get_event_loop().time()
    while (asyncio.get_event_loop().time() - start_time) < timeout:
        if await test_connection(url):
            logging.info("Database connection established!")
            return True

        await asyncio.sleep(interval)

    logging.error("Database connection timeout")
    return False


class MigrationManager:
    def __init__(self, database_url: str, alembic_cfg_path: str):
        """
        Initialize migration manager

        :param database_url: Database connection URL
        :param alembic_cfg_path: Path to alembic.ini file
        """

        self.db_config = DatabaseConfig(database_url)
        self.alembic_cfg = Config(alembic_cfg_path)

        self.alembic_cfg.set_main_option("sqlalchemy.url", self.db_config.sync_url)

    async def is_database_empty(self, engine: AsyncEngine) -> bool:
        """Check if database has any tables"""
        async with engine.connect() as connection:
            result = await connection.execute(
                text(
                    """
                SELECT COUNT(*) 
                FROM information_schema.tables 
                WHERE table_schema = 'public'
                AND table_type = 'BASE TABLE';
            """
                )
            )
            count = result.scalar()
            return count == 0

    def init_alembic_config(self) -> None:
        """Initialize Alembic configuration and create initial migration if needed"""
        try:
            migrations_dir = os.path.join(
                os.path.dirname(self.alembic_cfg.config_file_name), "versions"
            )
            if not os.path.exists(migrations_dir):
                os.makedirs(migrations_dir)
                logging.info("Created migrations directory")

            if not os.listdir(migrations_dir):
                logging.info("No migrations found, creating initial migration...")
                command.revision(
                    self.alembic_cfg, autogenerate=True, message="Initial migration"
                )
                logging.info("Initial migration created")
            elif self.has_model_changes():
                logging.info("Detected model changes, creating new migration...")
                command.revision(
                    self.alembic_cfg, autogenerate=True, message="Auto migration"
                )
                logging.info("New migration created for model changes")
            else:
                logging.info("No model changes detected, skipping migration creation")

        except Exception as e:
            logging.error(f"Error initializing Alembic: {e}")
            raise

    def has_model_changes(self) -> bool:
        """Check if there are any model changes that need migration"""
        from alembic.migration import MigrationContext
        from alembic.autogenerate import compare_metadata
        from sqlalchemy import create_engine
        from tgbot.database.database import Base

        engine = create_engine(self.db_config.sync_url)

        try:
            with engine.connect() as connection:
                context = MigrationContext.configure(connection)
                diff = compare_metadata(context, Base.metadata)

                return len(diff) > 0
        finally:
            engine.dispose()

    async def check_migrations(
        self, engine: AsyncEngine
    ) -> Tuple[bool, Optional[str], Optional[str]]:
        """
        Check if there are pending migrations
        """
        try:
            async with engine.connect() as connection:

                def get_current_revision(connection):
                    context = MigrationContext.configure(connection)
                    return context.get_current_revision()

                current_rev = await connection.run_sync(get_current_revision)

                # Get latest available revision
                script = ScriptDirectory.from_config(self.alembic_cfg)
                head_rev = script.get_current_head()

                needs_upgrade = current_rev != head_rev

                return needs_upgrade, current_rev, head_rev

        except Exception as e:
            logging.error(f"Error checking migrations: {e}")
            raise

    async def ensure_version_table(self, engine: AsyncEngine) -> None:
        """Create alembic_version table if it doesn't exist"""
        async with engine.connect() as connection:
            # Check if alembic_version table exists
            result = await connection.execute(
                text(
                    """
                    SELECT EXISTS (
                        SELECT FROM information_schema.tables 
                        WHERE table_name = 'alembic_version'
                    );
                """
                )
            )
            exists = result.scalar()

            if not exists:
                # Initialize alembic version table
                await connection.execute(
                    text(
                        "CREATE TABLE alembic_version (version_num VARCHAR(32) PRIMARY KEY);"
                    )
                )
                await connection.execute(
                    text("INSERT INTO alembic_version (version_num) VALUES ('head');")
                )
                await connection.commit()

    async def apply_migrations(self, engine: AsyncEngine) -> None:
        """Apply migrations - create initial migration if database is empty"""
        try:
            is_empty = await self.is_database_empty(engine)

            if is_empty or self.has_model_changes():
                logging.info(
                    "Database changes detected, initializing migrations..."
                    if not is_empty
                    else "Empty database detected, initializing migrations..."
                )
                self.init_alembic_config()

            logging.info("Applying migrations...")
            command.upgrade(self.alembic_cfg, "head")
            logging.info("Migrations applied successfully")

        except Exception as e:
            logging.error(f"Error applying migrations: {e}")
            raise

    async def init_migrations(self, engine: AsyncEngine) -> None:
        """Initialize migrations with connection check"""
        if not await wait_for_database(self.db_config.async_url):
            raise ConnectionError("Could not connect to database")

        await self.apply_migrations(engine)


async def init_db_and_migrations(database_url: str, alembic_cfg_path: str) -> None:
    """
    Initialize database and apply migrations if needed

    Usage example:
        await init_db_and_migrations(
            database_url="postgresql+asyncpg://user:pass@localhost/dbname",
            alembic_cfg_path="./alembic.ini"
        )
    """
    # Create engine
    engine = create_async_engine(
        database_url,
        # echo=True,
    )

    # Initialize migration manager
    migration_manager = MigrationManager(database_url, alembic_cfg_path)

    try:
        # Apply migrations if needed
        await migration_manager.init_migrations(engine)

    except Exception as e:
        logging.error(f"Failed to initialize database and migrations: {e}")
        raise

    finally:
        await engine.dispose()

    async_session_factory = async_sessionmaker(engine)
    AsyncORM.set_session_factory(async_session_factory)
    AsyncORM.init_models()
