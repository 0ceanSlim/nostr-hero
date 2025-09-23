#!/usr/bin/env python3
"""
World Map Visualizer for Nostr Hero
Generates a visual overview of the game world matching the exact grid layout from GAME_DESIGN.md
"""

import json
import os
from collections import defaultdict
import matplotlib.pyplot as plt
import matplotlib.patches as patches
from matplotlib.patches import FancyBboxPatch, Rectangle
import numpy as np

class WorldMapVisualizer:
    def __init__(self, data_path="..\\www\\data\\locations"):
        self.data_path = data_path
        self.cities = {}
        self.environments = {}
        self.all_districts = {}
        self.connections = defaultdict(set)
        self.racial_starting_cities = {}

    def load_locations(self):
        """Load all location JSON files"""
        cities_path = os.path.join(self.data_path, "cities")
        environments_path = os.path.join(self.data_path, "environments")

        # Load cities
        if os.path.exists(cities_path):
            for filename in os.listdir(cities_path):
                if filename.endswith('.json'):
                    filepath = os.path.join(cities_path, filename)
                    with open(filepath, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                        self.cities[data['id']] = data

                        # Extract all districts
                        for district_name, district_data in data.get('districts', {}).items():
                            district_id = district_data['id']
                            district_data['parent_city'] = data['id']
                            district_data['district_name'] = district_name
                            self.all_districts[district_id] = district_data

        # Load environments
        if os.path.exists(environments_path):
            for filename in os.listdir(environments_path):
                if filename.endswith('.json'):
                    filepath = os.path.join(environments_path, filename)
                    with open(filepath, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                        self.environments[data['id']] = data

        # Load racial starting cities
        racial_file = os.path.join(os.path.dirname(self.data_path), "racial-starting-cities.json")
        if os.path.exists(racial_file):
            with open(racial_file, 'r', encoding='utf-8') as f:
                racial_data = json.load(f)
                self.racial_starting_cities = racial_data.get("racial_starting_cities", {})

    def get_starting_races(self, city_id):
        """Get races that start in this city"""
        races = []
        for race, starting_city in self.racial_starting_cities.items():
            if starting_city == city_id:
                races.append(race)
        return races

    def create_grid_world_map(self):
        """Create the exact world map layout from GAME_DESIGN.md"""
        fig, ax = plt.subplots(figsize=(20, 16))

        # Define the exact grid positions based on GAME_DESIGN.md
        # Grid coordinates (x, y) - 7x7 grid
        grid_positions = {
            # Row 6 (top)
            'town-north': (1, 6),
            'arctic-north-town': (3, 6),
            'town-northeast': (5, 6),

            # Row 5
            'swamp-kingdom': (1, 4),
            'mountain-northeast': (5, 4),

            # Row 4 (middle)
            'village-west': (0, 2),
            'hill-kingdom': (1, 2),
            'kingdom': (3, 2),  # Center
            'urban-kingdom': (5, 2),
            'city-east': (6, 2),
            'desert-city': (7, 2),
            'village-southeast': (8, 2),

            # Row 3
            'forest-kingdom': (3, 1),

            # Row 2
            'city-south': (3, 0),
            'swamp-south': (4, 0),
            'village-southwest': (5, 0),

            # Row 1
            'coastal-south': (3, -1),

            # Row 0 (bottom)
            'village-south': (3, -2)
        }

        # Draw grid
        for location_id, (x, y) in grid_positions.items():
            self.draw_location_at_grid(ax, location_id, x, y)

        # Draw connections
        self.draw_grid_connections(ax, grid_positions)

        # Set up the plot
        ax.set_xlim(-1, 9)
        ax.set_ylim(-3, 7)
        ax.set_aspect('equal')
        ax.grid(True, alpha=0.3)
        ax.set_title("Nostr Hero World Map - Exact Layout from Game Design", fontsize=16, fontweight='bold', pad=20)

        # Add legend
        self.add_comprehensive_legend(ax)

        plt.tight_layout()
        output_path = "exact_world_map.png"
        plt.savefig(output_path, dpi=300, bbox_inches='tight')
        print(f"Exact world map saved as: {output_path}")
        plt.show()

    def draw_location_at_grid(self, ax, location_id, grid_x, grid_y):
        """Draw a location (city or environment) at specific grid coordinates"""
        # Convert grid coordinates to plot coordinates
        x, y = grid_x, grid_y

        if location_id in self.cities:
            self.draw_city_detailed(ax, location_id, x, y)
        elif location_id in self.environments:
            self.draw_environment(ax, location_id, x, y)

    def draw_city_detailed(self, ax, city_id, x, y):
        """Draw a city with all its districts"""
        city = self.cities[city_id]
        city_type = city.get('type', 'city')

        # City type colors and sizes
        colors = {
            'kingdom': '#FFD700',  # Gold
            'city': '#87CEEB',     # Sky blue
            'town': '#98FB98',     # Pale green
            'village': '#F0E68C'   # Khaki
        }

        sizes = {
            'kingdom': 0.4,
            'city': 0.35,
            'town': 0.3,
            'village': 0.25
        }

        color = colors.get(city_type, '#FFFFFF')
        size = sizes.get(city_type, 0.3)

        # Draw main city area
        city_rect = FancyBboxPatch((x-size/2, y-size/2), size, size,
                                  boxstyle="round,pad=0.02", facecolor=color,
                                  edgecolor='black', linewidth=2, alpha=0.9)
        ax.add_patch(city_rect)

        # Add city name
        ax.text(x, y+0.1, city['name'], ha='center', va='center',
               fontsize=10, fontweight='bold')

        # Add city type
        ax.text(x, y, city_type.upper(), ha='center', va='center',
               fontsize=8, fontweight='bold', alpha=0.7)

        # Add starting races
        starting_races = self.get_starting_races(city_id)
        if starting_races:
            race_text = ', '.join(starting_races)
            ax.text(x, y-0.15, f"Starts: {race_text}", ha='center', va='center',
                   fontsize=7, style='italic')

        # Add entry fee
        entry_fee = city.get('entry_fee', 0)
        ax.text(x, y-0.25, f"Fee: {entry_fee}g", ha='center', va='center',
               fontsize=7, color='darkgreen')

        # Draw districts around the city
        self.draw_city_districts(ax, city, x, y, size)

    def draw_city_districts(self, ax, city, center_x, center_y, city_size):
        """Draw districts around a city"""
        districts = city.get('districts', {})
        district_size = 0.08
        district_offset = city_size/2 + district_size + 0.05

        # District positions relative to center
        district_positions = {
            'north': (center_x, center_y + district_offset),
            'south': (center_x, center_y - district_offset),
            'east': (center_x + district_offset, center_y),
            'west': (center_x - district_offset, center_y)
        }

        for district_name, district_data in districts.items():
            if district_name in district_positions:
                dx, dy = district_positions[district_name]

                # Draw district
                district_rect = FancyBboxPatch((dx-district_size/2, dy-district_size/2),
                                             district_size, district_size,
                                             boxstyle="round,pad=0.01",
                                             facecolor='lightgray', alpha=0.6,
                                             edgecolor='gray', linewidth=1)
                ax.add_patch(district_rect)

                # Label district
                ax.text(dx, dy, district_name[0].upper(), ha='center', va='center',
                       fontsize=6, fontweight='bold')

                # Draw connection to city center
                ax.plot([center_x, dx], [center_y, dy], 'k-', alpha=0.3, linewidth=1)

    def draw_environment(self, ax, env_id, x, y):
        """Draw an environment"""
        env = self.environments[env_id]
        env_type = env.get('environment_type', 'unknown')

        # Environment colors
        env_colors = {
            'forest': '#228B22',    # Forest green
            'hill': '#9ACD32',      # Yellow green
            'swamp': '#556B2F',     # Dark olive green
            'mountain': '#696969',  # Dim gray
            'arctic': '#E0FFFF',    # Light cyan
            'desert': '#F4A460',    # Sandy brown
            'coastal': '#20B2AA',   # Light sea green
            'urban': '#D3D3D3',     # Light gray
            'underwater': '#4682B4' # Steel blue
        }

        color = env_colors.get(env_type, '#DCDCDC')

        # Draw environment
        env_rect = FancyBboxPatch((x-0.3, y-0.15), 0.6, 0.3,
                                 boxstyle="round,pad=0.02", facecolor=color,
                                 edgecolor='darkgreen', linewidth=2, alpha=0.8)
        ax.add_patch(env_rect)

        # Add environment name
        ax.text(x, y+0.05, env['name'], ha='center', va='center',
               fontsize=9, fontweight='bold')

        # Add environment type
        ax.text(x, y-0.05, env_type.title(), ha='center', va='center',
               fontsize=7, style='italic')

        # Add travel info
        travel_time = env.get('travel_time', '?')
        difficulty = env.get('travel_difficulty', 'unknown')
        rations = env.get('required_rations', '?')
        ax.text(x, y-0.15, f"{travel_time}d, {difficulty}, {rations}r",
               ha='center', va='center', fontsize=6, color='darkblue')

    def draw_grid_connections(self, ax, grid_positions):
        """Draw connections between locations based on the grid layout"""
        # Define the connections from the design document
        connections = [
            # Top row
            ('town-north', 'arctic-north-town'),
            ('arctic-north-town', 'town-northeast'),

            # Vertical connections
            ('town-north', 'swamp-kingdom'),
            ('town-northeast', 'mountain-northeast'),
            ('swamp-kingdom', 'kingdom'),
            ('mountain-northeast', 'city-east'),

            # Middle row
            ('village-west', 'hill-kingdom'),
            ('hill-kingdom', 'kingdom'),
            ('kingdom', 'urban-kingdom'),
            ('urban-kingdom', 'city-east'),
            ('city-east', 'desert-city'),
            ('desert-city', 'village-southeast'),

            # Vertical from kingdom
            ('kingdom', 'forest-kingdom'),
            ('forest-kingdom', 'city-south'),

            # Bottom connections
            ('city-south', 'swamp-south'),
            ('swamp-south', 'village-southwest'),
            ('city-south', 'coastal-south'),
            ('coastal-south', 'village-south')
        ]

        for loc1, loc2 in connections:
            if loc1 in grid_positions and loc2 in grid_positions:
                x1, y1 = grid_positions[loc1]
                x2, y2 = grid_positions[loc2]

                # Draw connection line
                ax.plot([x1, x2], [y1, y2], 'gray', linewidth=3, alpha=0.7)

                # Add arrow to show direction
                dx, dy = x2 - x1, y2 - y1
                ax.annotate('', xy=(x2, y2), xytext=(x1, y1),
                           arrowprops=dict(arrowstyle='->', color='gray', alpha=0.5))

    def add_comprehensive_legend(self, ax):
        """Add a comprehensive legend"""
        # City type legend
        city_legend = [
            patches.Patch(color='#FFD700', label='Kingdom (4 connections)'),
            patches.Patch(color='#87CEEB', label='City (3 connections)'),
            patches.Patch(color='#98FB98', label='Town (2 connections)'),
            patches.Patch(color='#F0E68C', label='Village (1 connection)')
        ]

        # Environment legend
        env_legend = [
            patches.Patch(color='#228B22', label='Forest'),
            patches.Patch(color='#9ACD32', label='Hills'),
            patches.Patch(color='#556B2F', label='Swamp'),
            patches.Patch(color='#696969', label='Mountain'),
            patches.Patch(color='#E0FFFF', label='Arctic'),
            patches.Patch(color='#F4A460', label='Desert'),
            patches.Patch(color='#20B2AA', label='Coastal'),
            patches.Patch(color='#D3D3D3', label='Urban')
        ]

        # Create two legends
        legend1 = ax.legend(handles=city_legend, title="Cities & Settlements",
                           loc='upper left', bbox_to_anchor=(0.02, 0.98))
        legend2 = ax.legend(handles=env_legend, title="Environments",
                           loc='upper left', bbox_to_anchor=(0.02, 0.65))

        # Add the first legend back
        ax.add_artist(legend1)

    def create_text_overview(self):
        """Create a detailed text overview"""
        print("\n" + "="*80)
        print("NOSTR HERO WORLD MAP - EXACT LAYOUT")
        print("="*80)

        print("\nWORLD GRID LAYOUT:")
        print("""
                    Town----------Arctic---------Town
                      |                           |
                      |                           |
                    Swamp                      Mountain
                      |                           |
                      |                           |
Village---Hill-----KINGDOM---------Urban---------City----Desert---Village
                      |
                      |
                    Forest
                      |
                      |
                    City---------Swamp---------Village
                      |
                      |
                   Coastal
                      |
                      |
                   Village
        """)

        print("\nLOCATION DETAILS:")
        print("-" * 50)

        # Kingdom (center)
        kingdom = self.cities.get('kingdom')
        if kingdom:
            races = self.get_starting_races('kingdom')
            race_str = f" | Starts: {', '.join(races)}" if races else ""
            print(f"🏰 KINGDOM: {kingdom['name']} | Fee: {kingdom.get('entry_fee', 0)}g{race_str}")

        # Cities
        for city_id in ['city-east', 'city-south']:
            if city_id in self.cities:
                city = self.cities[city_id]
                races = self.get_starting_races(city_id)
                race_str = f" | Starts: {', '.join(races)}" if races else ""
                print(f"🏛️ CITY: {city['name']} | Fee: {city.get('entry_fee', 0)}g{race_str}")

        # Towns
        for town_id in ['town-north', 'town-northeast']:
            if town_id in self.cities:
                town = self.cities[town_id]
                races = self.get_starting_races(town_id)
                race_str = f" | Starts: {', '.join(races)}" if races else ""
                print(f"🏘️ TOWN: {town['name']} | Fee: {town.get('entry_fee', 0)}g{race_str}")

        # Villages
        for village_id in ['village-west', 'village-southeast', 'village-southwest', 'village-south']:
            if village_id in self.cities:
                village = self.cities[village_id]
                races = self.get_starting_races(village_id)
                race_str = f" | Starts: {', '.join(races)}" if races else ""
                print(f"🏡 VILLAGE: {village['name']} | Fee: {village.get('entry_fee', 0)}g{race_str}")

        print("\nENVIRONMENTS:")
        print("-" * 50)
        for env_id, env in self.environments.items():
            env_type = env.get('environment_type', 'unknown').title()
            travel_time = env.get('travel_time', '?')
            difficulty = env.get('travel_difficulty', 'unknown')
            rations = env.get('required_rations', '?')
            print(f"🌍 {env_type.upper()}: {env['name']} | {travel_time} days, {difficulty}, {rations} rations")

    def generate_image_prompt(self):
        """Generate a prompt for AI image generation"""
        print("\n" + "="*80)
        print("AI IMAGE GENERATION PROMPT")
        print("="*80)

        # Create the prompt
        prompt = """Create a medieval fantasy world map in pixel art style showing:

MAIN PROMPT:
A top-down pixel art fantasy world map in 16-bit retro style with medieval aesthetics. The map should show a grid-based layout with distinct biomes and settlements connected by paths.

LAYOUT (from top to bottom, left to right):
- TOP ROW: Frosthold Town (snowy northern town) connected by icy path to Frozen Wastes (arctic wasteland) connected to Ironpeak Town (mountain mining town)
- SECOND ROW: Mistmarsh Swamplands (dark swamp) below Frosthold, Cragspire Mountains (tall peaks) below Ironpeak
- MIDDLE ROW: Millhaven Village (farming village) - Windswept Hills (rolling green hills) - The Royal Kingdom (grand golden castle city, center) - Merchant's Highway (stone road) - Goldenhaven City (wealthy trade city) - Sunscorch Desert (sandy dunes) - Dusthaven Village (oasis village)
- FOURTH ROW: Darkwood Forest (dense green forest) directly below the Kingdom
- FIFTH ROW: Verdant City (nature city with gardens) - Shadowmere Wetlands (dark swamp) - Marshlight Village (swamp village on stilts)
- SIXTH ROW: Suncrest Coastlands (rocky coastline) below Verdant City
- BOTTOM ROW: Saltwind Village (fishing village by the sea)

VISUAL STYLE:
- 16-bit pixel art aesthetic like classic SNES RPGs
- Medieval fantasy theme with castles, villages, forests, mountains
- Each location should be visually distinct with appropriate colors and terrain
- The Kingdom should be the largest and most impressive (golden castle)
- Cities should have multiple buildings, towns should be medium-sized, villages should be small
- Environments should show their terrain type clearly (green forests, brown mountains, blue water, yellow desert, etc.)
- Stone paths/roads connecting the locations
- Grid-based layout that's easy to follow
- Rich, saturated colors typical of retro fantasy games

SPECIFIC DETAILS:
- The Royal Kingdom: Large golden castle with multiple towers at the center
- Goldenhaven City: Wealthy port city with ships and golden buildings
- Verdant City: City surrounded by gardens and green spaces
- Frosthold Town: Snow-covered buildings with thick walls
- Ironpeak Town: Built into mountainside with mining equipment
- Villages: Small clusters of houses appropriate to their environment
- Environments: Clearly defined biomes with appropriate pixel art textures"""

        print(prompt)

        print("\n" + "="*80)
        print("COPY THE ABOVE PROMPT TO YOUR IMAGE AI MODEL")
        print("="*80)

    def generate_world_overview(self):
        """Generate complete world overview"""
        print("Loading world data...")
        self.load_locations()

        print("Creating detailed text overview...")
        self.create_text_overview()

        print("\nGenerating AI image prompt...")
        self.generate_image_prompt()

        print("\nGenerating exact grid world map...")
        self.create_grid_world_map()

def main():
    """Main function"""
    try:
        visualizer = WorldMapVisualizer()
        visualizer.generate_world_overview()
    except ImportError as e:
        if "matplotlib" in str(e):
            print("Error: matplotlib is required for visual map generation.")
            print("Install it with: pip install matplotlib")
            print("\nGenerating text overview only...")
            visualizer = WorldMapVisualizer()
            visualizer.load_locations()
            visualizer.create_text_overview()
        else:
            raise e
    except Exception as e:
        print(f"Error: {e}")
        print("Make sure you're running this from the tools directory and the data files exist.")

if __name__ == "__main__":
    main()