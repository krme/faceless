import io
import json
from typing import List
import numpy as np
from pydantic import BaseModel
import os
import asyncpg
from uuid import UUID
import librosa
import soundfile as sf
from pydub import AudioSegment

from custom_types.user import User
from custom_types.attempt import Attempt


def convert_blob_to_librosa(blob):
    if not blob:
        return (None, None)
    with open('temp.webm', 'ab') as f:
        f.write(blob)
        webm: AudioSegment = AudioSegment.from_file('temp.webm', format='webm')
        webm.export('temp.wav', format='wav')
        y, sr = librosa.load('temp.wav', sr=None)
        # bytes = io.BytesIO(blob)
        # y, sr = librosa.load(bytes, sr=None) 
    return y, sr


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
        user=os.environ.get("PG_USERNAME"),
        password=os.environ.get("PG_PASSWORD"),
        database=os.environ.get("PG_DATABASE"),
        host=os.environ.get("PG_HOST"),
        port=os.environ.get("PG_PORT", '5432') 
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
            raise Exception("Failed to create database connection (NONE).")

        print("Database connection created successfully")
        return conn
    except Exception as e:
        print(f"Failed to connect to the database: {e}")
        raise

async def close_db(conn):
    await conn.close()
    print("Database connection closed")


async def get_user(db_config: DBConfig, rid: UUID):
    conn = None
    recordings = []
    try:
        conn = await init_db(db_config)

        query = """
            SELECT
                id,
                rid,
                recording_1,
                recording_2,
                recording_3
            FROM "user"
            WHERE rid = $1;"""
        query_result = await conn.fetch(query, rid)
        # add convert_blob_to_librosa
        user: User = User.from_json_map(query_result[0])

        recordings = [convert_blob_to_librosa(recording) for recording in user.get_recordings()]

        # recordings.append(recordings.from_json_map(query_result))

    except Exception as e:
        raise Exception(f"Error while getting user data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
            print("Database connection closed successfully")
    return recordings


async def get_latest_identification_attempt(db_config: DBConfig, rid: UUID):
    conn = None
    try:
        conn = await init_db(db_config)

        query = """
            SELECT
                id,
                rid,
                recording
            FROM identification_attempt
            WHERE rid = $1
            ORDER BY created_at DESC
            LIMIT 1;
        """
        query_result = await conn.fetch(query, rid)
        # extract binary file from json
        # reconstruct adio file from binary
        # add convert_blob_to_librosa
        attempt = Attempt.from_json_map(query_result)

        recording = convert_blob_to_librosa(attempt.recording)

    except Exception as e:
        raise Exception(f"Error while getting latest_identification data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
    return recording


async def update_user(db_config: DBConfig, rid: UUID, mfcc: List[float]):
    conn = None
    try:
        conn = await init_db(db_config)

        mfcc_1 = np.zeros(40).tolist()
        mfcc_2 = np.zeros(40).tolist()
        mfcc_3 = np.zeros(40).tolist()
        if len(mfcc[0].tolist()) == 40:
            mfcc_1 = mfcc[0].tolist()
        if len(mfcc[1].tolist()) == 40:
            mfcc_2 = mfcc[1].tolist()
        if len(mfcc[2].tolist()) == 40:
            mfcc_3 = mfcc[2].tolist()

        insert_query = """
            UPDATE
                "user"
            SET
                recording_1_mfcc = $1,
                recording_2_mfcc = $2,
                recording_3_mfcc = $3,
                updated_at = NOW()
            WHERE 
                rid=$4;
        """

        # Einfügen in die Datenbank
        await conn.execute(insert_query, json.dumps(mfcc_1), json.dumps(mfcc_2), json.dumps(mfcc_3), rid)
    except Exception as e:
        raise Exception(f"Error while updating user data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)


async def update_latest_identification_attempt(db_config: DBConfig, rid: UUID, identified: bool, mfcc: List[float]):
    conn = None
    try:
        conn = await init_db(db_config)

        # SQL-Query zur Einfügung der Embeddings in die Datenbank
        insert_query = """
            UPDATE
                "user"
            SET
                recording_mfcc = $1,
                identified = $2,
                updated_at = NOW()
            WHERE
                rid=$3;
        """

        # Einfügen in die Datenbank
        await conn.executemany(insert_query, mfcc, identified, rid)

    except Exception as e:
        raise Exception(f"Error while updating latest_identification data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)


async def get_vector_dist(db_config: DBConfig, rid: UUID, recording_mfcc: List[float]):
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
        raise Exception(f"Error while getting dist of vector: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
    return query_result