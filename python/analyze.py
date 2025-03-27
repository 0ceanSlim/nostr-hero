import json
from collections import Counter
from typing import List, Dict, Tuple

def load_json_data(file_path: str) -> List[Dict]:
    with open(file_path, 'r') as file:
        return json.load(file)

def count_races(data: List[Dict]) -> Dict[str, int]:
    return Counter(entry['character']['race'] for entry in data)

def calculate_total_stats(stats: Dict[str, int]) -> int:
    return sum(stats.values())

def find_impressive_npubs(data: List[Dict]) -> Tuple[str, str]:
    npub_stats = [(entry['npub'], calculate_total_stats(entry['character']['stats'])) for entry in data]
    most_impressive = max(npub_stats, key=lambda x: x[1])
    least_impressive = min(npub_stats, key=lambda x: x[1])
    return most_impressive[0], least_impressive[0]

def calculate_uniqueness_score(entry: Dict, class_weights: List[Dict], background_weights: List[Dict]) -> float:
    race = entry['character']['race']
    character_class = entry['character']['class']
    background = entry['character']['background']
    
    class_weight = next((item['weight'] for item in class_weights if item['race'] == race and item['class'] == character_class), 0)
    background_weight = next((item['weight'] for item in background_weights if item['race'] == race and item['background'] == background), 0)
    
    return 100 - (class_weight + background_weight) / 2

def find_most_unique_character(data: List[Dict], class_weights: List[Dict], background_weights: List[Dict]) -> str:
    uniqueness_scores = [(entry['npub'], calculate_uniqueness_score(entry, class_weights, background_weights)) for entry in data]
    return max(uniqueness_scores, key=lambda x: x[1])[0]

def main():
    data = load_json_data('registry.json')
    
    # Count races
    race_counts = count_races(data)
    print("Race counts:")
    for race, count in race_counts.items():
        print(f"{race}: {count}")
    
    # Find impressive NPUBs
    most_impressive, least_impressive = find_impressive_npubs(data)
    print(f"\nMost impressive NPUB: {most_impressive}")
    print(f"Least impressive NPUB: {least_impressive}")
    
    # Define class and background weights
    class_weights = [
    {"race": "Human", "class": "Fighter", "weight": 20},
    {"race": "Human", "class": "Wizard", "weight": 15},
    {"race": "Human", "class": "Rogue", "weight": 15},
    {"race": "Human", "class": "Cleric", "weight": 12},
    {"race": "Human", "class": "Paladin", "weight": 10},
    {"race": "Human", "class": "Ranger", "weight": 10},
    {"race": "Human", "class": "Bard", "weight": 10},
    {"race": "Human", "class": "Warlock", "weight": 8},
    {"race": "Elf", "class": "Ranger", "weight": 25},
    {"race": "Elf", "class": "Wizard", "weight": 20},
    {"race": "Elf", "class": "Rogue", "weight": 15},
    {"race": "Elf", "class": "Bard", "weight": 12},
    {"race": "Elf", "class": "Fighter", "weight": 10},
    {"race": "Elf", "class": "Cleric", "weight": 8},
    {"race": "Elf", "class": "Warlock", "weight": 5},
    {"race": "Elf", "class": "Paladin", "weight": 5},
    {"race": "Dwarf", "class": "Fighter", "weight": 25},
    {"race": "Dwarf", "class": "Cleric", "weight": 20},
    {"race": "Dwarf", "class": "Paladin", "weight": 15},
    {"race": "Dwarf", "class": "Ranger", "weight": 12},
    {"race": "Dwarf", "class": "Rogue", "weight": 10},
    {"race": "Dwarf", "class": "Wizard", "weight": 8},
    {"race": "Dwarf", "class": "Warlock", "weight": 5},
    {"race": "Dwarf", "class": "Bard", "weight": 5},
    {"race": "Halfling", "class": "Rogue", "weight": 30},
    {"race": "Halfling", "class": "Bard", "weight": 20},
    {"race": "Halfling", "class": "Ranger", "weight": 15},
    {"race": "Halfling", "class": "Fighter", "weight": 10},
    {"race": "Halfling", "class": "Wizard", "weight": 8},
    {"race": "Halfling", "class": "Cleric", "weight": 7},
    {"race": "Halfling", "class": "Warlock", "weight": 5},
    {"race": "Halfling", "class": "Paladin", "weight": 5},
    {"race": "Orc", "class": "Fighter", "weight": 30},
    {"race": "Orc", "class": "Barbarian", "weight": 25},
    {"race": "Orc", "class": "Ranger", "weight": 15},
    {"race": "Orc", "class": "Cleric", "weight": 10},
    {"race": "Orc", "class": "Warlock", "weight": 8},
    {"race": "Orc", "class": "Rogue", "weight": 5},
    {"race": "Orc", "class": "Wizard", "weight": 4},
    {"race": "Orc", "class": "Bard", "weight": 3},
    {"race": "Tiefling", "class": "Warlock", "weight": 30},
    {"race": "Tiefling", "class": "Sorcerer", "weight": 25},
    {"race": "Tiefling", "class": "Rogue", "weight": 15},
    {"race": "Tiefling", "class": "Bard", "weight": 10},
    {"race": "Tiefling", "class": "Fighter", "weight": 8},
    {"race": "Tiefling", "class": "Wizard", "weight": 5},
    {"race": "Tiefling", "class": "Cleric", "weight": 4},
    {"race": "Tiefling", "class": "Paladin", "weight": 3},
    {"race": "Gnome", "class": "Wizard", "weight": 25},
    {"race": "Gnome", "class": "Bard", "weight": 20},
    {"race": "Gnome", "class": "Rogue", "weight": 15},
    {"race": "Gnome", "class": "Warlock", "weight": 12},
    {"race": "Gnome", "class": "Fighter", "weight": 10},
    {"race": "Gnome", "class": "Cleric", "weight": 8},
    {"race": "Gnome", "class": "Ranger", "weight": 5},
    {"race": "Gnome", "class": "Paladin", "weight": 5},
    {"race": "Dragonborn", "class": "Paladin", "weight": 30},
    {"race": "Dragonborn", "class": "Fighter", "weight": 20},
    {"race": "Dragonborn", "class": "Sorcerer", "weight": 15},
    {"race": "Dragonborn", "class": "Cleric", "weight": 12},
    {"race": "Dragonborn", "class": "Warlock", "weight": 8},
    {"race": "Dragonborn", "class": "Bard", "weight": 6},
    {"race": "Dragonborn", "class": "Ranger", "weight": 5},
    {"race": "Dragonborn", "class": "Rogue", "weight": 4},
]
    
    background_weights = [
    {"race": "Human", "background": "Soldier", "weight": 20},
    {"race": "Human", "background": "Noble", "weight": 20},
    {"race": "Human", "background": "Charlatan", "weight": 15},
    {"race": "Human", "background": "Sage", "weight": 15},
    {"race": "Human", "background": "Outlander", "weight": 10},
    {"race": "Human", "background": "Acolyte", "weight": 10},
    {"race": "Human", "background": "Entertainer", "weight": 5},
    {"race": "Human", "background": "Hermit", "weight": 5},
    {"race": "Elf", "background": "Sage", "weight": 25},
    {"race": "Elf", "background": "Noble", "weight": 20},
    {"race": "Elf", "background": "Outlander", "weight": 15},
    {"race": "Elf", "background": "Hermit", "weight": 10},
    {"race": "Elf", "background": "Acolyte", "weight": 10},
    {"race": "Elf", "background": "Entertainer", "weight": 8},
    {"race": "Elf", "background": "Soldier", "weight": 7},
    {"race": "Elf", "background": "Charlatan", "weight": 5},
    {"race": "Dwarf", "background": "Soldier", "weight": 25},
    {"race": "Dwarf", "background": "Acolyte", "weight": 20},
    {"race": "Dwarf", "background": "Outlander", "weight": 15},
    {"race": "Dwarf", "background": "Noble", "weight": 12},
    {"race": "Dwarf", "background": "Sage", "weight": 10},
    {"race": "Dwarf", "background": "Hermit", "weight": 8},
    {"race": "Dwarf", "background": "Entertainer", "weight": 5},
    {"race": "Dwarf", "background": "Charlatan", "weight": 5},
    {"race": "Halfling", "background": "Charlatan", "weight": 30},
    {"race": "Halfling", "background": "Entertainer", "weight": 25},
    {"race": "Halfling", "background": "Hermit", "weight": 15},
    {"race": "Halfling", "background": "Outlander", "weight": 10},
    {"race": "Halfling", "background": "Acolyte", "weight": 8},
    {"race": "Halfling", "background": "Sage", "weight": 5},
    {"race": "Halfling", "background": "Soldier", "weight": 4},
    {"race": "Halfling", "background": "Noble", "weight": 3},
    {"race": "Orc", "background": "Soldier", "weight": 30},
    {"race": "Orc", "background": "Outlander", "weight": 25},
    {"race": "Orc", "background": "Hermit", "weight": 15},
    {"race": "Orc", "background": "Charlatan", "weight": 10},
    {"race": "Orc", "background": "Entertainer", "weight": 8},
    {"race": "Orc", "background": "Acolyte", "weight": 5},
    {"race": "Orc", "background": "Noble", "weight": 4},
    {"race": "Orc", "background": "Sage", "weight": 3},
    {"race": "Tiefling", "background": "Charlatan", "weight": 25},
    {"race": "Tiefling", "background": "Noble", "weight": 20},
    {"race": "Tiefling", "background": "Hermit", "weight": 15},
    {"race": "Tiefling", "background": "Entertainer", "weight": 12},
    {"race": "Tiefling", "background": "Sage", "weight": 10},
    {"race": "Tiefling", "background": "Outlander", "weight": 8},
    {"race": "Tiefling", "background": "Soldier", "weight": 5},
    {"race": "Tiefling", "background": "Acolyte", "weight": 5},
    {"race": "Gnome", "background": "Sage", "weight": 25},
    {"race": "Gnome", "background": "Entertainer", "weight": 20},
    {"race": "Gnome", "background": "Charlatan", "weight": 15},
    {"race": "Gnome", "background": "Hermit", "weight": 12},
    {"race": "Gnome", "background": "Acolyte", "weight": 10},
    {"race": "Gnome", "background": "Noble", "weight": 8},
    {"race": "Gnome", "background": "Outlander", "weight": 5},
    {"race": "Gnome", "background": "Soldier", "weight": 5},
    {"race": "Dragonborn", "background": "Soldier", "weight": 30},
    {"race": "Dragonborn", "background": "Noble", "weight": 25},
    {"race": "Dragonborn", "background": "Acolyte", "weight": 15},
    {"race": "Dragonborn", "background": "Sage", "weight": 10},
    {"race": "Dragonborn", "background": "Outlander", "weight": 8},
    {"race": "Dragonborn", "background": "Entertainer", "weight": 5},
    {"race": "Dragonborn", "background": "Charlatan", "weight": 4},
    {"race": "Dragonborn", "background": "Hermit", "weight": 3},
]
    
    # Find most unique character
    most_unique = find_most_unique_character(data, class_weights, background_weights)
    print(f"\nMost unique character NPUB: {most_unique}")

if __name__ == "__main__":
    main()
