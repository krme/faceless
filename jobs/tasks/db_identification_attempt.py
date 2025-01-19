from typing import List
from uuid import UUID

from custom_types.attempt import Attempt
from jobs.tasks.compare import convert_blob_to_librosa
from jobs.tasks.db_helper import DBConfig, close_db, init_db


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


async def get_latest_identification_attempt(db_config: DBConfig, rid: UUID):
    conn = None
    try:
        conn = await init_db(db_config)

        query = """
            SELECT
                id,
                rid,
                user_rid,
                recording
            FROM identification_attempt
            WHERE user_rid = $1
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
        print(f"Error while getting latest_identification data: {str(e)}")
        raise Exception(f"Error while getting latest_identification data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
    return recording
