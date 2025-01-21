import logging
from uuid import UUID

from fastapi import APIRouter, FastAPI, HTTPException
from pydantic import BaseModel
from tasks.db_helper import load_user_db_config, load_identification_db_config
from tasks.db_identification_attempt import get_latest_identification_attempt, update_latest_identification_attempt
from tasks.db_user import get_user, update_user, get_vector_dist

from tasks.compare import convert_blob_to_librosa, preprocess_recording, extract_features


router = APIRouter()
logger = logging.getLogger(__name__)


class ProcessReferenceRecordingsRequest(BaseModel):
    rid: UUID


@router.post("/jobs/processReferenceRecordings")
async def process_reference_recordings(request: ProcessReferenceRecordingsRequest):
    """
    Process the reference recordings the user recorded on the website
    """
    try:
        dbConfigUser = load_user_db_config()

        recordings = await get_user(dbConfigUser, request.rid)

        preprocessed_recordings = []
        for (recording, sr) in recordings:
            preprocessed_recordings.append(preprocess_recording(recording, sr))

        mfccs = []
        for (recording, sr) in preprocessed_recordings:
            mfccs.append(extract_features(recording, sr))

        await update_user(dbConfigUser, request.rid, mfccs)
    except Exception as e:
        logger.error(str(e))
        raise HTTPException(status_code=500, detail=str(e))


class IdentifyRequest(BaseModel):
    user_rid: UUID


@router.post("/jobs/identify")
async def identify(request: IdentifyRequest):
    """
    To identify the user at login
    """
    try:
        dbConfigUser = load_user_db_config()
        dbConfigIdentification = load_identification_db_config()
        
        attempt = await get_latest_identification_attempt(dbConfigIdentification, request.user_rid)

        recording, sr = convert_blob_to_librosa(attempt.recording)

        preprocessed_recording, sr = preprocess_recording(recording, sr)

        mfcc = extract_features(preprocessed_recording, sr)

        dist = await get_vector_dist(dbConfigUser, request.user_rid, mfcc)

        # TODO adjust threshold
        logger.info(f"distance of identification: {dist}")
        identified = False
        if dist < 5:
            identified = True

        await update_latest_identification_attempt(dbConfigIdentification, attempt.rid, identified, mfcc)

        attempt = await get_latest_identification_attempt(dbConfigIdentification, request.user_rid)
        print(attempt.toString())
    except Exception as e:
        logger.error(str(e))
        raise HTTPException(status_code=500, detail=str(e))


app = FastAPI()
app.include_router(router)