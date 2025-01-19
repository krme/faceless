from fastapi import APIRouter, FastAPI
from mistralai import Mistral


router = APIRouter()


@router.post("/createSentence")
def createSentence() -> str:
    api_key = "rNQf5SkjXzuEbKHMjRGdsmgWlBLODXhz"
    model = "mistral-small-latest"

    client = Mistral(api_key=api_key)

    chat_response = client.chat.complete(
        model = model,
        messages = [
            {
                "role": "user",
                "top_p": 0.9,
                "content": "Create 1 absurdly creative, very, very, very short sentence with no constraints on logic or topic in simple English and nothing else, no nothing. without sure or anything.",
            },
        ]
    )

    print(chat_response.choices[0].message.content)
    return chat_response.choices[0].message.content


app = FastAPI()
app.include_router(router)
