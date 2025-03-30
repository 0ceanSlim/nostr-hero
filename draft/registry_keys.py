import json
import csv

# Read the JSON file
with open('registry.json', 'r') as json_file:
    data = json.load(json_file)

# Extract pubkeys
pubkeys = [entry['pubkey'] for entry in data]

# Write pubkeys to CSV
with open('pubkeys.csv', 'w', newline='') as csv_file:
    writer = csv.writer(csv_file)
    writer.writerow(['pubkey'])  # Header
    for pubkey in pubkeys:
        writer.writerow([pubkey])

print(f"Extracted {len(pubkeys)} pubkeys and saved to pubkeys.csv")
