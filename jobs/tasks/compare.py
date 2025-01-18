import numpy as np
import librosa

def extract_features(file_path): 
    # Load the audio file 
    y, sr = librosa.load(file_path, sr=None) 
     
    # Extract MFCCs s
    mfccs = librosa.feature.mfcc(y=y, sr=sr, n_mfcc=40) 
     
    # Take the mean of the MFCCs across time 
    mfccs_mean = np.mean(mfccs.T, axis=0) 
     
    return mfccs_mean 


def preprocess_recording(recording):
    recording = librosa.effects.trim()

    return recording
