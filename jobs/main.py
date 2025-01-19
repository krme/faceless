import os
import logging
from dotenv import load_dotenv

import uvicorn
from fastapi import FastAPI, Request

import handler.compareAudio as compareAudio
import handler.createSentence as createSentence


load_dotenv()


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)
app = FastAPI()
app.include_router(compareAudio.router)
app.include_router(createSentence.router)


@app.get("/")
async def health():
    return {"status": "healthy"}


@app.middleware("https")
async def log_requests(request: Request, call_next):
    response = await call_next(request)
    logger.info(f"Received {request.method} request to {request.url}")
    logger.info(f"Returning {response.status_code} response")
    return response


if __name__ == "__main__":
    port = int(os.environ.get("JOBS_PORT", 3000))
    print('Server is running')
    uvicorn.run(app, host="0.0.0.0", port=port)
