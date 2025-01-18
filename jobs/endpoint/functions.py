from fastapi import APIRouter, HTTPException, FastAPI
from jobs.flows.compareAudio import register_user, identify_attempt
from uuid import UUID


router = APIRouter()

@router.post("/identify")
async def identify(rid: UUID):
    """
    To identify the user at login
    """
    try:
        result = identify_attempt(rid)
        return {"identified": result}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/ProcessReferenceRecordings")
async def process_reference_recordings(rid: UUID):
    """
    To process the recordings the user recorded on the website
    """
    try:
        result = register_user(rid)
        return {"registered": result}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

app = FastAPI()
app.include_router(router)