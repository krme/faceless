import json
from uuid import UUID

import numpy as np

from custom_types.attempt import Attempt
from tasks.db_helper import DBConfig, close_db, init_db


async def update_latest_identification_attempt(db_config: DBConfig, rid: UUID, identified: bool, mfcc: np.ndarray):
    conn = None
    try:
        conn = await init_db(db_config)

        # SQL-Query zur Einfügung der Embeddings in die Datenbank
        insert_query = """
            UPDATE
                identification_attempt
            SET
                recording_mfcc = $1,
                identified = $2,
                updated_at = NOW()
            WHERE
                rid=$3;
        """

        await conn.fetch(insert_query, json.dumps(mfcc.tolist()), identified, rid)
    except Exception as e:
        raise Exception(f"Error while updating latest_identification data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)


async def get_latest_identification_attempt(db_config: DBConfig, user_rid: UUID) -> Attempt:
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
        query_result = await conn.fetch(query, user_rid)

        if len(query_result) == 0:
            raise Exception("No identification attempt found")

        attempt = Attempt.from_json_map(query_result[0])
    except Exception as e:
        print(f"Error while getting latest_identification data: {str(e)}")
        raise Exception(f"Error while getting latest_identification data: {str(e)}")
    finally:
        if conn:
            await close_db(conn)
    return attempt
