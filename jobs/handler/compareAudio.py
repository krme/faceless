import logging
from uuid import UUID

from fastapi import APIRouter, FastAPI, HTTPException
from pydantic import BaseModel
from tasks.db_functions import get_user, load_db_config, update_user, get_latest_identification_attempt, update_latest_identification_attempt, get_vector_dist
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

 
# def compare_audio(file1, file2): 
#     # Extract features from both audio files 
#     features1 = extract_features(file1) 
#     features2 = extract_features(file2) 
     
#     # Compute the Euclidean distance between the feature vectors 
#     distance = np.linalg.norm(features1 - features2) 
     
#     return distance 
 
# # Example usage 
# file1 = './cat.wav' 
# file2 = './elephant.wav' 


# # file1 = './700-122866-0000.flac' 
# # file2 = './700-122866-0001.flac' 

# # file1 = './116-288045-0001.flac' 
# # file2 = './116-288045-0000.flac' 
 
# distance = compare_audio(file1, file2) 
# print(f"Distance between the two audio files: {distance}") 
 
# # Set a threshold for comparison 
# threshold = 35  # You may need to adjust this based on your data was 30
# if distance < threshold: 
#     print("The same person is likely speaking in both audio files.") 
# else: 
#     print("The speakers are likely different.") 
