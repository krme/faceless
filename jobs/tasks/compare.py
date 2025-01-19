from typing import Tuple
import numpy as np
import librosa


def extract_features(y: np.ndarray, sr: float) -> np.ndarray:
    if y is None or sr is None or sr == 0:
        return np.array([])
    mfccs = librosa.feature.mfcc(y=y, sr=sr, n_mfcc=40) 
    mfccs_mean = np.mean(mfccs.T, axis=0) 
    return mfccs_mean 


def preprocess_recording(y, sr) -> Tuple[np.ndarray, float]:
    if y is None or sr is None:
        return np.array([]), 0
    recording, _ = librosa.effects.trim(y)
    return recording, sr
