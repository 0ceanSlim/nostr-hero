/**
 * Item Helper Utilities
 * Shared functions for loading and accessing item data from the database
 */

// Shared cache for item database
let itemsDatabaseCache = null;

/**
 * Load all items from the database API
 * Uses caching to avoid repeated API calls
 * @returns {Promise<Array>} Array of item objects
 */
async function loadItemsFromDatabase() {
  if (itemsDatabaseCache) {
    return itemsDatabaseCache;
  }

  try {
    const response = await fetch("/api/items");
    if (response.ok) {
      itemsDatabaseCache = await response.json();
      console.log(`📦 Loaded ${itemsDatabaseCache.length} items from database`);
      return itemsDatabaseCache;
    }
  } catch (error) {
    console.warn("Could not load items from database:", error);
  }
  return [];
}

/**
 * Get item data from database cache by ID
 * @param {string} itemId - The item ID to look up
 * @returns {Promise<Object|null>} Item object or null if not found
 */
async function getItemById(itemId) {
  try {
    const items = await loadItemsFromDatabase();

    // Find item by ID (exact match)
    const item = items.find((i) => i.id === itemId);

    if (item) {
      // Convert database format to expected frontend format
      return {
        id: item.id,
        name: item.name,
        description: item.description,
        type: item.item_type,
        tags: item.tags || [],
        rarity: item.rarity,
        gear_slot: item.properties?.gear_slot,
        slots: item.properties?.slots,
        contents: item.properties?.contents,
        stack: item.properties?.stack,
        ...item.properties, // Spread all other properties
      };
    } else {
      console.warn(`❌ Item ID "${itemId}" not found in database`);
    }
  } catch (error) {
    console.warn(`Could not load item data for ID: ${itemId}`, error);
  }
  return null;
}

/**
 * Clear the items cache (useful for forcing a reload)
 */
function clearItemsCache() {
  itemsDatabaseCache = null;
}
