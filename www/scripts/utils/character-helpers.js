/**
 * Character & Vault Helper Utilities
 * Shared functions for character creation and vault management
 */

/**
 * Generate empty starting vault for a location
 * @param {string} location - Location ID
 * @returns {Object} Vault object with 40 empty slots
 */
function generateStartingVault(location) {
  const vaultSlots = [];
  for (let i = 0; i < 40; i++) {
    vaultSlots.push({
      slot: i,
      item: null,
      quantity: 0
    });
  }

  return {
    location: location,
    building: getVaultBuildingForLocation(location),
    slots: vaultSlots
  };
}

/**
 * Get vault building ID for a given location
 * @param {string} location - Location ID
 * @returns {string} Building ID for the vault
 */
function getVaultBuildingForLocation(location) {
  const vaultBuildings = {
    'kingdom': 'vault_of_crowns',
    'village-west': 'burrowlock',
    'village-south': 'halfling_burrows',
    'village-southeast': 'secure_cellars',
    'village-southwest': 'stone_vaults',
    'town-north': 'northwatch_vault',
    'town-northeast': 'stormhold_storage',
    'city-east': 'shadowhaven_vaults',
    'city-south': 'coastal_storage',
    'forest-kingdom': 'silverwood_treasury',
    'hill-kingdom': 'ironforge_vaults',
    'mountain-northeast': 'draconis_hoard',
    'swamp-kingdom': 'mire_keep_storage'
  };

  return vaultBuildings[location] || 'vault_of_crowns';
}

/**
 * Convert location/district/building IDs to display names
 * @param {string} locationId - Location ID
 * @param {string} districtKey - District key
 * @param {string} buildingId - Building ID
 * @returns {Promise<Object>} Object with display names
 */
async function getDisplayNamesForLocation(locationId, districtKey, buildingId) {
  try {
    const response = await fetch('/api/locations');
    if (!response.ok) {
      console.warn('Failed to fetch locations from API');
      return { location: locationId, district: districtKey, building: buildingId };
    }

    const allLocations = await response.json();

    // Find the location
    const location = allLocations.find(loc => loc.id === locationId);
    if (!location) {
      console.warn(`Location not found: ${locationId}`);
      return { location: locationId, district: districtKey, building: buildingId };
    }

    // Find the district
    const district = location.properties?.districts?.[districtKey];
    if (!district) {
      console.warn(`District not found: ${districtKey} in ${locationId}`);
      return {
        location: location.name || locationId,
        district: districtKey,
        building: buildingId
      };
    }

    // Find the building
    const building = district.buildings?.find(b => b.id === buildingId);
    if (!building) {
      console.warn(`Building not found: ${buildingId} in ${districtKey}`);
      return {
        location: location.name || locationId,
        district: district.name || districtKey,
        building: buildingId
      };
    }

    return {
      location: location.name || locationId,
      district: district.name || districtKey,
      building: building.name || buildingId
    };

  } catch (error) {
    console.error('Error fetching location names:', error);
    return { location: locationId, district: districtKey, building: buildingId };
  }
}
