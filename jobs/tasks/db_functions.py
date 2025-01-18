from pydantic import BaseModel, Field
import os
import asyncpg
import ssl
from prefect import task, get_run_logger
from uuid import UUID


class DBConfig(BaseModel):
    """
    DB crendentials
    """
    user: str
    password: str
    database: str
    host: str
    port: str

def load_db_config() -> DBConfig:
    return DBConfig(
        user=os.environ.get("PG_USER"),
        password=os.environ.get("PG_PASSWORD"),
        database=os.environ.get("PG_DATABASE"),
        host=os.environ.get("PG_HOST"),
        port=os.environ.get("PG_PORT", '5432') 
    )

async def init_db(config: DBConfig) -> asyncpg.Pool:
    # Pfad zum Client-Zertifikat
    # SSL-Verbindung konfigurieren
    ssl_context = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
    
    # Disable hostname checking and certificate verification
    ssl_context.check_hostname = False  # Do not verify hostname
    ssl_context.verify_mode = ssl.CERT_NONE  # Do not verify certificates

    try:
        conn = await asyncpg.connect(
            user=config.user,
            password=config.password,
            database=config.database,
            host=config.host,
            port=config.port,
            ssl=ssl_context
        )

        # Create tables if not existent
        #await create_extension_and_tables_if_not_exists(conn) TODO check what to do if no table exists

        if not conn:
            raise Exception("Failed to create database connection (NONE).")

        print("Database connection created successfully")
        return conn
    except Exception as e:
        print(f"Failed to connect to the database: {e}")
        raise

async def close_db(conn):
    await conn.close()
    print("Database connection closed")


@task
async def get_user(db_config: DBConfig, rid: UUID):
    # logger = get_run_logger()
    conn = None
    recordings = []
    try:
        logger = get_run_logger()
        conn = await init_db(db_config)

        query = "SELECT id, rid, recording_1, recording_2, recording_3 FROM DB.user WHERE rid = $1;"
        query_result = await conn.fetch(query, rid)
        recordings.append(recordings.from_json_map(query_result))

    except Exception as e:
        logger.error(f"Error while getting user data: {str(e)}")
        raise Exception(f"Error while getting user data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
            print("Database connection closed successfully")
    return recordings


@task
async def get_latest_identification_attempt(db_config: DBConfig, rid: UUID):
    # logger = get_run_logger()
    conn = None
    recordings = []
    try:
        logger = get_run_logger()
        conn = await init_db(db_config)

        query = "SELECT id, rid, recording FROM DB.identification_attempt WHERE rid = $1 ORDER BY created_at DESC LIMIT 1;"
        query_result = await conn.fetch(query, rid)
        recordings.append(recordings.from_json_map(query_result))
        # extract binary file from json
        # reconstruct adio file from binary


    except Exception as e:
        logger.error(f"Error while getting user data: {str(e)}")
        raise Exception(f"Error while getting user data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
            print("Database connection closed successfully")
    return recordings




@task
async def update_user(db_config: DBConfig, rid: UUID, recordings_normalised, mfcc):
    conn = None
    try:
        logger = get_run_logger()
        conn = await init_db(db_config)

        # SQL-Query zur Einfügung der Embeddings in die Datenbank
        insert_query = """
            UPDATE DB.user SET 
            recording_1_normalised = $1, recording_2_normalised = $2, recording_3_normalised = $3 
            recording_1_mfcc = $4, recording_2_mfcc = $5, recording_3_mfcc = $6
            WHERE rid=$7 ;
        """

        #insert_values = []
        #for recording in recordings_normalised:
        #    insert_values.append((chunk_diff.chunk_id, chunk_diff.chunk_rid, chunk_diff.flow_id, chunk_diff.diff_type.value, chunk_diff.summary, chunk_diff.recommendation))
        
        # Einfügen in die Datenbank
        await conn.executemany(insert_query, recordings_normalised[0],recordings_normalised[1],recordings_normalised[2], mfcc[0],mfcc[1],mfcc[2],rid)
        logger.info(f"3 recordings_normalised and mfccs inserted successfully")

    except Exception as e:
        logger.error(f"Error while updating user data: {str(e)}")
        raise Exception(f"Error while updating user data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
            print("Database connection closed successfully")


@task
async def update_latest_identification_attempt(db_config: DBConfig, rid: UUID, recording_normalised, mfcc):
    conn = None
    try:
        logger = get_run_logger()
        conn = await init_db(db_config)

        # SQL-Query zur Einfügung der Embeddings in die Datenbank
        insert_query = """
            UPDATE DB.user SET recording_normalised = $1, recording_mfcc = $2 WHERE rid=$3 ;
        """

        #insert_values = []
        #for recording in recordings:
        #    insert_values.append((chunk_diff.chunk_id, chunk_diff.chunk_rid, chunk_diff.flow_id, chunk_diff.diff_type.value, chunk_diff.summary, chunk_diff.recommendation))
        
        # Einfügen in die Datenbank
        await conn.executemany(insert_query, recording_normalised, mfcc, rid)
        logger.info(f"1 recording_normalised and mfcc inserted successfully")

    except Exception as e:
        logger.error(f"Error while updating user data: {str(e)}")
        raise Exception(f"Error while updating user data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
            print("Database connection closed successfully")



@task
async def get_vector_dist(db_config: DBConfig, rid UUID, recording_mfcc ):
    '''
    Gets chunks by ordered by vector distance and with a distance threshhold.
    The limit of chunks is used per doc_hash.
    '''
    logger = get_run_logger()

    conn = None
    try:
        conn = await init_db(db_config)
        query = """
        SELECT
            ((d.distance_1 + d.distance_2 + d.distance_3) / 3)
        FROM
            (SELECT 
                (recording_1_mfcc_mean <-> $2::vector(40)) AS distance_1
                (recording_2_mfcc_mean <-> $2::vector(40)) AS distance_2
                (recording_3_mfcc_mean <-> $2::vector(40)) AS distance_3
            FROM 
                "user"
            WHERE 
                u.rid = $1) AS d;"""

        query_result = await conn.fetch(query, rid, recording_mfcc)

    except Exception as e:
        logger.error(f"Error while getting chunks by vector: {str(e)}")
        raise Exception(f"Error while getting chunks by vector: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
            logger.info("Database connection closed successfully")
    return query_result