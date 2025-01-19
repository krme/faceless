import logging
from uuid import UUID

from fastapi import APIRouter, FastAPI, HTTPException
from pydantic import BaseModel
from tasks.db_helper import load_db_config
from tasks.db_identification_attempt import get_latest_identification_attempt, update_latest_identification_attempt
from tasks.db_user import get_user, load_db_config, update_user, get_vector_dist

from tasks.compare import preprocess_recording, extract_features


router = APIRouter()
logger = logging.getLogger(__name__)


class ProcessReferenceRecordingsRequest(BaseModel):
    rid: UUID


@router.post("/jobs/processReferenceRecordings")
async def process_reference_recordings(request: ProcessReferenceRecordingsRequest):
    """
    Process the reference recordings the user recorded on the website
    """
    logger.info("______ Hallo ______")
    try:
        dbConfig = load_db_config()
        recordings = await get_user(dbConfig, request.rid)

        preprocessed_recordings = []
        for (recording, sr) in recordings:
            preprocessed_recordings.append(preprocess_recording(recording, sr))

        logger.info(preprocessed_recordings)

        mfccs = []
        for (recording, sr) in preprocessed_recordings:
            mfccs.append(extract_features(recording, sr))

        logger.info("mfccs")
        
        await update_user(dbConfig, request.rid, mfccs)
    except Exception as e:
        logger.error(str(e))
        raise HTTPException(status_code=500, detail=str(e))


class IdentifyRequest(BaseModel):
    rid: UUID


@router.post("/jobs/identify")
async def identify(request: IdentifyRequest):
    """
    To identify the user at login
    """
    try:
        dbConfig = load_db_config()
        recording, sr = await get_latest_identification_attempt(dbConfig, request.rid)

        preprocessed_recording = preprocess_recording(recording, sr)

        mfcc = extract_features(preprocessed_recording)

        dist = get_vector_dist(mfcc)

        # adjust threshold
        identified = False
        if dist < 50:
            identified = True

        await update_latest_identification_attempt(dbConfig, request.rid, identified, mfcc)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


app = FastAPI()
app.include_router(router)
