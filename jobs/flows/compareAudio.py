from jobs.tasks.db_functions import get_user, init_db, update_user, close_db, get_latest_identification_attempt, get_vector_dist, update_latest_identification
from jobs.tasks.compare import preprocess_recording, extract_features
import datetime
# from sklearn.metrics import pairwise 


def register_user(uid):
    conn = init_db()
    recordings = get_user(conn, uid)

    preprocessed_recordings = []
    for recording in recordings:
        preprocessed_recordings.append(preprocess_recording(recording))

    mfccs = []
    for recording in preprocessed_recordings:
        mfccs.append(extract_features(recording))
    
    update_user(conn, uid, preprocessed_recordings, mfcc_mean)
    close_db(conn)


def identify_attempt(uid):
    conn = init_db()
    recording = get_latest_identification_attempt()

    preprocessed_recording = preprocess_recording(recording)

    mfcc = extract_features(preprocessed_recording)

    dist = get_vector_dist(mfcc)

    # adjust threshold
    if dist < 50:
        identified = True
    else:
        identified = False

    timestamp = datetime.datetime.timestamp()
    
    update_latest_identification()

 
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
