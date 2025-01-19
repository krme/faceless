from uuid import UUID

from fastapi import APIRouter, FastAPI, HTTPException
from pydantic import BaseModel
from tasks.db_functions import get_user, load_db_config, update_user, get_latest_identification_attempt, get_vector_dist, update_latest_identification_attempt
from tasks.compare import preprocess_recording, extract_features


router = APIRouter()


class ProcessReferenceRecordingsRequest(BaseModel):
    rid: UUID


@router.post("/processReferenceRecordings")
async def process_reference_recordings(request: ProcessReferenceRecordingsRequest):
    """
    To process the recordings the user recorded on the website
    """
    try:
        dbConfig = load_db_config()
        recordings = get_user(dbConfig, request.rid)

        preprocessed_recordings = []
        for recording in recordings:
            preprocessed_recordings.append(preprocess_recording(recording))

        mfccs = []
        for recording in preprocessed_recordings:
            mfccs.append(extract_features(recording))
        
        update_user(dbConfig, request.rid, preprocessed_recordings, mfccs)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


class IdentifyRequest(BaseModel):
    rid: UUID


@router.post("/identify")
async def identify(request: IdentifyRequest):
    """
    To identify the user at login
    """
    try:
        dbConfig = load_db_config()
        recording = get_latest_identification_attempt(dbConfig, request.rid)

        preprocessed_recording = preprocess_recording(recording)

        mfcc = extract_features(preprocessed_recording)

        dist = get_vector_dist(mfcc)

        # adjust threshold
        identified = False
        if dist < 50:
            identified = True

        update_latest_identification_attempt(dbConfig, request.rid, identified, mfcc)
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
