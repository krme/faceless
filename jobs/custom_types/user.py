from uuid import UUID
from typing import Dict

from pydantic import BaseModel


class User(BaseModel):
    """
    Chunk data structure
    """
    id: int = 0
    rid: UUID = UUID(int=0)
    recording_1: bytes | None
    recording_2: bytes | None
    recording_3: bytes | None

    def from_json_map(json: Dict[str, any]):
        print(json)
        return User(
            id=json['id'],
            rid=json['rid'],
            recording_1=json['recording_1'],
            recording_2=json['recording_2'],
            recording_3=json['recording_3']
        )

    def get_recordings(self):
        return [self.recording_1, self.recording_2, self.recording_3]