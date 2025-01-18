import numpy as np 
import librosa 
# from sklearn.metrics import pairwise 
 
def extract_features(file_path): 
    # Load the audio file 
    y, sr = librosa.load(file_path, sr=None) 
     
    # Extract MFCCs 
    mfccs = librosa.feature.mfcc(y=y, sr=sr, n_mfcc=40) 
     
    # Take the mean of the MFCCs across time 
    mfccs_mean = np.mean(mfccs.T, axis=0) 
     
    return mfccs_mean 
 
def compare_audio(file1, file2): 
    # Extract features from both audio files 
    features1 = extract_features(file1) 
    features2 = extract_features(file2) 
     
    # Compute the Euclidean distance between the feature vectors 
    distance = np.linalg.norm(features1 - features2) 
     
    return distance 
 
# Example usage 
file1 = './cat.wav' 
file2 = './elephant.wav' 


# file1 = './700-122866-0000.flac' 
# file2 = './700-122866-0001.flac' 

# file1 = './116-288045-0001.flac' 
# file2 = './116-288045-0000.flac' 
 
distance = compare_audio(file1, file2) 
print(f"Distance between the two audio files: {distance}") 
 
# Set a threshold for comparison 
threshold = 35  # You may need to adjust this based on your data was 30
if distance < threshold: 
    print("The same person is likely speaking in both audio files.") 
else: 
    print("The speakers are likely different.") 
