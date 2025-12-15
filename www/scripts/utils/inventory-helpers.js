/**
 * Inventory Helper Utilities
 * Shared functions for creating and managing character inventories
 */

/**
 * Add items with proper stacking logic
 * @param {Array} allItems - Array of items
 * @param {string} itemId - Item ID to add
 * @param {number} quantity - Quantity to add
 */
async function addItemWithStacking(allItems, itemId, quantity) {
  // Get item data to check stack limit
  const itemData = await getItemById(itemId);
  const stackLimit = itemData ? parseInt(itemData.stack) || 1 : 1;

  let remainingQuantity = quantity;

  // Try to add to existing stacks first
  for (let existingItem of allItems) {
    if (existingItem.item === itemId && existingItem.quantity < stackLimit) {
      const canAdd = Math.min(
        remainingQuantity,
        stackLimit - existingItem.quantity
      );
      existingItem.quantity += canAdd;
      remainingQuantity -= canAdd;

      if (remainingQuantity <= 0) break;
    }
  }

  // Create new stacks for remaining quantity
  while (remainingQuantity > 0) {
    const stackSize = Math.min(remainingQuantity, stackLimit);
    allItems.push({ item: itemId, quantity: stackSize });
    remainingQuantity -= stackSize;
  }
}

/**
 * Unpack packs and return contents
 * @param {string} packId - Pack item ID
 * @returns {Promise<Array|null>} Array of slot objects or null
 */
async function unpackItem(packId) {
  try {
    console.log(`🎒 Attempting to unpack: "${packId}"`);
    const packData = await getItemById(packId);
    if (packData) {
      console.log(`🎒 Loaded pack data for ${packId}:`, packData);
      if (packData.contents) {
        // Parse contents string if it's a string, or use directly if array
        const contents =
          typeof packData.contents === "string"
            ? JSON.parse(packData.contents)
            : packData.contents;

        console.log(`🎒 Pack contents:`, contents);
        // Convert to proper slot format
        const slots = [];
        let slotIndex = 0;
        contents.forEach((item) => {
          const itemId = item[0];
          const itemName = itemId.toLowerCase();
          // Don't include the backpack itself or any pack items
          if (
            itemName !== "backpack" &&
            !itemName.includes("-pack") &&
            !itemName.includes("pack-")
          ) {
            slots.push({
              slot: slotIndex,
              item: itemId,
              quantity: item[1],
            });
            slotIndex++;
          } else {
            console.log(`🎒 Excluding from pack contents: ${itemId}`);
          }
        });

        // Fill remaining slots with null
        const totalSlots = 20; // Backpack has 20 slots
        while (slots.length < totalSlots) {
          slots.push({
            slot: slots.length,
            item: null,
            quantity: 0,
          });
        }

        console.log(
          `🎒 Successfully unpacked ${packId} into ${
            slots.filter((s) => s.item).length
          } items`
        );
        return slots;
      } else {
        console.warn(`🎒 Pack ${packId} has no contents field`);
      }
    } else {
      console.warn(`🎒 Pack data not found for: ${packId}`);
    }
  } catch (error) {
    console.warn(`🎒 Could not unpack: ${packId}`, error);
  }
  return null;
}

/**
 * Add items to general slots or bag
 * @param {Object} inventory - Inventory object
 * @param {Object} item - Item to add
 * @param {number} slotIndex - General slot index
 * @param {Object} itemData - Item data (optional)
 */
function addToGeneralSlotOrBag(inventory, item, slotIndex, itemData = null) {
  // Check if item is a container (has 'container' tag)
  const isContainer =
    itemData?.tags?.includes("container") ||
    item.item.toLowerCase().includes("pouch") ||
    item.item.toLowerCase().includes("pack");

  // Containers should NOT go in the bag (no containers inside containers)
  if (isContainer) {
    console.log(`📦 ${item.item} is a container, placing in general slot`);
    if (slotIndex < 4 && inventory.general_slots[slotIndex].item === null) {
      inventory.general_slots[slotIndex] = {
        slot: slotIndex,
        item: item.item,
        quantity: item.quantity,
      };
      return;
    }
  } else {
    // Try to add to backpack first if it exists (only non-containers)
    if (
      inventory.gear_slots.bag.item !== null &&
      inventory.gear_slots.bag.contents
    ) {
      const emptyBagSlot = inventory.gear_slots.bag.contents.find(
        (slot) => slot.item === null
      );
      if (emptyBagSlot) {
        console.log(
          `📦 Adding ${item.item} to backpack slot ${emptyBagSlot.slot}`
        );
        emptyBagSlot.item = item.item;
        emptyBagSlot.quantity = item.quantity;
        return;
      }
    }
  }

  // If backpack is full or item is a container, use general slots
  if (slotIndex < 4 && inventory.general_slots[slotIndex].item === null) {
    console.log(`📦 Adding ${item.item} to general slot ${slotIndex}`);
    inventory.general_slots[slotIndex] = {
      slot: slotIndex,
      item: item.item,
      quantity: item.quantity,
    };
    return;
  }

  console.warn(`⚠️ Could not place item ${item.item} - inventory full`);
}

/**
 * Dynamic inventory creation with proper equipment placement
 * @param {Array} allItems - Array of item objects {item, quantity}
 * @returns {Promise<Object>} Inventory object with gear_slots and general_slots
 */
async function createInventoryFromItems(allItems) {
  // Initialize empty inventory structure
  const inventory = {
    general_slots: [
      { slot: 0, item: null, quantity: 0 },
      { slot: 1, item: null, quantity: 0 },
      { slot: 2, item: null, quantity: 0 },
      { slot: 3, item: null, quantity: 0 },
    ],
    gear_slots: {
      bag: { item: null, quantity: 0 },
      left_arm: { item: null, quantity: 0 },
      right_arm: { item: null, quantity: 0 },
      armor: { item: null, quantity: 0 },
      necklace: { item: null, quantity: 0 },
      ring: { item: null, quantity: 0 },
      ammunition: { item: null, quantity: 0 },
      clothes: { item: null, quantity: 0 },
    },
  };

  let remainingItems = [...allItems];
  let currentGeneralSlot = 0;
  let twoHandedEquipped = false;

  // 1. First pass - Handle packs (automatically unpack to bag slot)
  for (let i = remainingItems.length - 1; i >= 0; i--) {
    const item = remainingItems[i];
    const itemName = item.item.toLowerCase();

    if (itemName.includes("pack")) {
      console.log(`🎒 Found pack: ${item.item}`);
      // This is a pack - unpack it to bag slot
      const packContents = await unpackItem(item.item);
      if (packContents) {
        console.log(`🎒 Successfully unpacked ${item.item}:`, packContents);
        // Equip the pack itself as a backpack to bag slot (the pack becomes the backpack)
        inventory.gear_slots.bag = {
          item: "backpack", // All packs become backpacks when equipped
          quantity: 1,
          contents: packContents,
        };
        remainingItems.splice(i, 1);
      }
    }
  }

  // 2. Second pass - Handle all equipment items based on gear_slot
  for (let i = remainingItems.length - 1; i >= 0; i--) {
    const item = remainingItems[i];
    const itemData = await getItemById(item.item);

    console.log(`🔍 Checking item: ${item.item}`, {
      hasData: !!itemData,
      tags: itemData?.tags,
      gearSlot: itemData?.gear_slot,
    });

    // Check if item has equipment tag and gear_slot
    if (
      itemData &&
      itemData.tags &&
      itemData.tags.includes("equipment") &&
      itemData.gear_slot
    ) {
      const gearSlot = itemData.gear_slot;
      console.log(`⚔️ Found equipment: ${item.item} → ${gearSlot}`);

      // Handle different gear slots
      if (gearSlot === "armor" && inventory.gear_slots.armor.item === null) {
        console.log(`Equipping armor: ${item.item}`);
        inventory.gear_slots.armor = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (
        gearSlot === "necklace" &&
        inventory.gear_slots.necklace.item === null
      ) {
        console.log(`Equipping necklace: ${item.item}`);
        inventory.gear_slots.necklace = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (
        gearSlot === "ring" &&
        inventory.gear_slots.ring.item === null
      ) {
        console.log(`Equipping ring: ${item.item}`);
        inventory.gear_slots.ring = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (
        gearSlot === "ammunition" &&
        inventory.gear_slots.ammunition.item === null
      ) {
        console.log(`Equipping ammunition: ${item.item}`);
        inventory.gear_slots.ammunition = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (
        gearSlot === "clothes" &&
        inventory.gear_slots.clothes.item === null
      ) {
        console.log(`Equipping clothes: ${item.item}`);
        inventory.gear_slots.clothes = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (gearSlot === "hands") {
        // Handle weapons - check if two-handed
        const isTwoHanded =
          itemData.tags && itemData.tags.includes("two-handed");

        if (isTwoHanded) {
          if (inventory.gear_slots.right_arm.item === null) {
            console.log(`Equipping two-handed weapon: ${item.item}`);
            inventory.gear_slots.right_arm = {
              item: item.item,
              quantity: item.quantity,
            };
            twoHandedEquipped = true;
            remainingItems.splice(i, 1);
          }
        } else {
          // One-handed weapon
          if (inventory.gear_slots.right_arm.item === null) {
            console.log(`Equipping weapon in right hand: ${item.item}`);
            inventory.gear_slots.right_arm = {
              item: item.item,
              quantity: item.quantity,
            };
            remainingItems.splice(i, 1);
          } else if (
            inventory.gear_slots.left_arm.item === null &&
            !twoHandedEquipped
          ) {
            console.log(`Equipping weapon in left hand: ${item.item}`);
            inventory.gear_slots.left_arm = {
              item: item.item,
              quantity: item.quantity,
            };
            remainingItems.splice(i, 1);
          }
        }
      }
    }
  }

  // 3. Put remaining items in general slots or bag
  for (let item of remainingItems) {
    const itemData = await getItemById(item.item);
    addToGeneralSlotOrBag(inventory, item, currentGeneralSlot++, itemData);
  }

  return inventory;
}
