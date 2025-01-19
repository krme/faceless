from uuid import UUID
from typing import Dict

from pydantic import BaseModel


class Attempt(BaseModel):
    """
    Chunk data structure
    """
    id: int = 0
    rid: UUID = UUID(int=0)
    user_rid: UUID = UUID(int=0)
    recording: bytes

    def from_json_map(json: Dict[str, any]):
        return Attempt(
            id=json['id'],
            rid=json['rid'],
            user_rid=json['user_rid'],
            recording=json['recording']
        )
    