import os
import json
import re
import glob
import subprocess
import matplotlib.pyplot as plt
import numpy as np

def main():
    # Create a local directory for results
    local_dir = "local_results"
    os.makedirs(local_dir, exist_ok=True)
    
    # Download files from GCS
    print("Downloading files from GCS...")
    cmd = f"gsutil cp gs://skill-activation-test/tests/results/*.json {local_dir}/"
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    
    if result.returncode != 0:
        print(f"Warning or Error downloading files: {result.stderr}")
        
    files = glob.glob(os.path.join(local_dir, "*.json"))
    print(f"Found {len(files)} JSON files.")
    
    # Data structure to hold pass rates
    # length is 1-10 (rows), specificity is 1-6 (columns)
    grid = np.full((10, 6), np.nan)
    
    # Regex to extract length and specificity
    pattern = re.compile(r'length-(\d+)_specificity-(\d+)')
    
    for file_path in files:
        filename = os.path.basename(file_path)
        match = pattern.search(filename)
        if match:
            length = int(match.group(1))
            specificity = int(match.group(2))
            
            try:
                with open(file_path, 'r') as f:
                    data = json.load(f)
                    pass_rate = data.get('pass_rate')
                    
                    if pass_rate is not None:
                        if 1 <= length <= 10 and 1 <= specificity <= 6:
                            grid[length-1, specificity-1] = pass_rate
            except Exception as e:
                print(f"Error reading {filename}: {e}")
                
    # Plotting the heatmap
    fig, ax = plt.subplots(figsize=(8, 10))
    
    # Use imshow to create the heatmap
    im = ax.imshow(grid, cmap='RdYlGn', vmin=0.0, vmax=1.0)
    
    # Show all ticks and label them
    ax.set_xticks(np.arange(6))
    ax.set_yticks(np.arange(10))
    ax.set_xticklabels(np.arange(1, 7))
    ax.set_yticklabels(np.arange(1, 11))
    
    # Axis labels
    ax.set_xlabel('Specificity')
    ax.set_ylabel('Length')
    ax.set_title('Skillgrade Pass Rate Heatmap')
    
    # Create colorbar
    cbar = ax.figure.colorbar(im, ax=ax)
    cbar.ax.set_ylabel("Pass Rate", rotation=-90, va="bottom")
    
    # Annotate cells with values
    for i in range(10):
        for j in range(6):
            val = grid[i, j]
            if not np.isnan(val):
                text = ax.text(j, i, f"{val:.2f}",
                               ha="center", va="center", color="black")
            else:
                text = ax.text(j, i, "N/A",
                               ha="center", va="center", color="gray")
                
    fig.tight_layout()
    plt.savefig('heatmap.png')
    print("Heatmap successfully saved to heatmap.png")

if __name__ == "__main__":
    main()

