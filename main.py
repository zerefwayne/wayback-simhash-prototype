import os
import re
import glob
import time
import string
from simhash import Simhash

import concurrent.futures

def get_file_paths():
    files = glob.glob("files/*.html")
    files.sort()
    return files

def read_file(file):
    with open(file, 'r', encoding='utf-8') as f:
        return f.read()

def read_files(files):
    html_files = []
    with concurrent.futures.ThreadPoolExecutor() as executor:
        html_files = list(executor.map(read_file, files))
    return html_files

def process_html(html):
    # Remove script and style elements
    html = re.sub(r'(?s)<(script|style).*?>.*?</\1>', '', html)
    
    # Get lowercase text
    html = html.lower()
    
    # Remove all punctuation
    html = html.translate(str.maketrans('', '', string.punctuation))
    
    # Break into lines and remove leading and trailing space on each
    lines = [line.strip() for line in html.split('\n')]
    
    # Break multi-headlines into a line each
    processed_lines = []
    for line in lines:
        processed_lines.extend(line.split('.'))
    
    # Drop blank lines and update features map
    file_feature = {}
    for line in processed_lines:
        if line:
            file_feature[line] = file_feature.get(line, 0) + 1
    
    return file_feature

def process_html_files(html_files):
    file_features = []
    with concurrent.futures.ThreadPoolExecutor() as executor:
        file_features = list(executor.map(process_html, html_files))
    return file_features

def compute_simhash(features):
    max_value = max(features.values())
    feature_list = [(k, v * 255 // max_value) for k, v in features.items()]
    return Simhash(feature_list).value

def compute_simhashes(file_features):
    simhashes = []
    with concurrent.futures.ThreadPoolExecutor() as executor:
        simhashes = list(executor.map(compute_simhash, file_features))
    return simhashes

def main():
    start = time.time()

    step_start = time.time()
    file_paths = get_file_paths()
    print(f"getFilePaths execution time: {time.time() - step_start:.4f}s")

    step_start = time.time()
    html_files = read_files(file_paths)
    print(f"readFiles execution time: {time.time() - step_start:.4f}s")

    step_start = time.time()
    file_features = process_html_files(html_files)
    print(f"processHTMLFiles execution time: {time.time() - step_start:.4f}s")

    step_start = time.time()
    simhashes = compute_simhashes(file_features)
    print(f"computeSimhashes execution time: {time.time() - step_start:.4f}s")

    print(f"Total execution time: {time.time() - start:.4f}s\n")

    for i, hash in enumerate(simhashes):
        print(f"File {i}: {hash:x}")

if __name__ == "__main__":
    main()