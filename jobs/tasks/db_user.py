import json
from typing import List
import numpy as np
from uuid import UUID

from custom_types.user import User
from custom_types.attempt import Attempt
from jobs.tasks.compare import convert_blob_to_librosa
from jobs.tasks.db_helper import DBConfig, close_db, init_db


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