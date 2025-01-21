import asyncpg
from pydantic import BaseModel
import os


class DBConfig(BaseModel):
    """
    DB crendentials
    """
    user: str
    password: str
    database: str
    host: str
    port: str


def load_user_db_config() -> DBConfig:
    return DBConfig(
        user=os.environ.get("DB_USER_USERNAME"),
        password=os.environ.get("DB_USER_PASSWORD"),
        database=os.environ.get("DB_USER_DATABASE"),
        host=os.environ.get("DB_USER_HOST"),
        port=os.environ.get("DB_USER_PORT", '5432') 
    )


def load_identification_db_config() -> DBConfig:
    return DBConfig(
        user=os.environ.get("DB_IDENTIFICATION_USERNAME"),
        password=os.environ.get("DB_IDENTIFICATION_PASSWORD"),
        database=os.environ.get("DB_IDENTIFICATION_DATABASE"),
        host=os.environ.get("DB_IDENTIFICATION_HOST"),
        port=os.environ.get("DB_IDENTIFICATION_PORT", '5432') 
    )


async def init_db(config: DBConfig) -> asyncpg.Pool:
    try:
        conn = await asyncpg.connect(
            user=config.user,
            password=config.password,
            database=config.database,
            host=config.host,
            port=config.port,
        )

        if not conn:
            raise Exception("No connection.")

        print("Database connection created successfully")
        return conn
    except Exception as e:
        print(f"Failed to connect to the database: {e}")
        raise


async def close_db(conn):
    await conn.close()
    print("Database connection closed")