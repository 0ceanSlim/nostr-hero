/**
 * Location Display UI Module
 *
 * Handles location display, navigation, music, buildings, NPCs, and vault UI.
 * Provides the main location scene interface with district navigation.
 *
 * @module ui/locationDisplay
 */

import { logger } from '../lib/logger.js';
import { gameAPI } from '../lib/api.js';
import { getGameStateSync, refreshGameState } from '../state/gameState.js';
import { getLocationById, getNPCById } from '../state/staticData.js';
import { updateTimeDisplay, formatTime } from './timeDisplay.js';
import { showActionText, showMessage } from './messaging.js';
import { moveToLocation } from '../logic/mechanics.js';
import { updateAllDisplays } from './displayCoordinator.js';

// Module-level state
let lastDisplayedLocation = null;
let lastBuildingState = null;

/**
 * Fetch NPCs at current location from backend
 * @param {string} location - Location ID (e.g., "kingdom")
 * @param {string} district - District key (e.g., "center")
 * @param {string} building - Building ID (optional, empty string if not in building)
 * @returns {Promise<string[]>} Array of NPC IDs at this location
 */
async function fetchNPCsAtLocation(location, district, building = '') {
    try {
        const state = getGameStateSync();
        const timeOfDay = state.character?.time_of_day !== undefined ? state.character.time_of_day : 720;
        const districtId = `${location}-${district}`;

        const params = new URLSearchParams({
            location: location,
            district: districtId,
            time: timeOfDay.toString()
        });

        if (building) {
            params.append('building', building);
        }

        const response = await fetch(`/api/npcs/at-location?${params.toString()}`);
        if (!response.ok) {
            logger.error('Failed to fetch NPCs:', response.statusText);
            return [];
        }

        const data = await response.json();
        logger.debug('Fetched NPCs at location:', data);

        // Extract NPC IDs from response
        return data.map(npc => npc.npc_id);
    } catch (error) {
        logger.error('Error fetching NPCs at location:', error);
        return [];
    }
}

/**
 * Display current location with navigation, buildings, and NPCs
 * Main location rendering function
 */
export async function displayCurrentLocation() {
    const state = getGameStateSync();
    const cityId = state.location?.current;
    const districtKey = state.location?.district || 'center';

    if (!cityId) return;

    // Construct full district ID from city + district (e.g., "village-west-east")
    const districtId = `${cityId}-${districtKey}`;
    const currentLocationId = districtId;  // For compatibility with rest of function

    logger.debug('Display location:', { cityId, districtKey, districtId });

    // Get the city data (for image, music)
    const cityData = getLocationById(cityId);
    if (!cityData) {
        logger.error('City not found:', cityId);
        return;
    }

    // Get the district data (for description, buildings, connections)
    const locationData = getLocationById(districtId);
    if (!locationData) {
        logger.error('District not found:', districtId);
        return;
    }

    logger.debug('City data:', cityData);
    logger.debug('District data:', locationData);

    // Update scene image (use city's image for all districts)
    const sceneImage = document.getElementById('scene-image');
    if (sceneImage && cityData.image) {
        sceneImage.src = cityData.image;
        sceneImage.alt = cityData.name;
    }

    // Update music (use city's music for all districts)
    if (cityData.music && window.musicSystem) {
        window.musicSystem.playLocationMusic();
    }

    // Update city name (top of scene)
    const cityName = document.getElementById('city-name');
    if (cityName) {
        cityName.textContent = cityData.name;
    }

    // Update district name (bottom of scene)
    const districtName = document.getElementById('district-name');
    if (districtName) {
        districtName.textContent = locationData.name;
    }

    // Update time of day display
    updateTimeDisplay();

    // Show location description in action text (white color) - only when:
    // 1. District changes (moving between districts)
    // 2. Exiting a building (going from building to outdoors)
    // NOT when entering a building
    const currentBuildingId = state.location?.building || '';
    const districtOnlyKey = `${cityId}-${districtKey}`;
    const wasInBuilding = lastBuildingState !== null && lastBuildingState !== '';
    const isInBuilding = currentBuildingId !== '';
    const exitedBuilding = wasInBuilding && !isInBuilding;
    const districtChanged = lastDisplayedLocation !== districtOnlyKey;

    if (locationData.description && (districtChanged || exitedBuilding)) {
        showActionText(locationData.description, 'white');
        lastDisplayedLocation = districtOnlyKey;
    }

    // Track building state for next time
    lastBuildingState = currentBuildingId;

    // Generate location actions based on city district structure
    const navContainer = document.getElementById('navigation-buttons');
    const buildingContainer = document.getElementById('building-buttons');
    const npcContainer = document.getElementById('npc-buttons');

    // Clear building and NPC containers (not navigation - that's handled per-slot)
    if (buildingContainer) {
        const buildingButtonContainer = buildingContainer.querySelector('div');
        if (buildingButtonContainer) buildingButtonContainer.innerHTML = '';
    }
    if (npcContainer) {
        const npcButtonContainer = npcContainer.querySelector('div');
        if (npcButtonContainer) npcButtonContainer.innerHTML = '';
    }

    // Get district data from location
    let districtData = null;
    if (locationData.properties?.districts) {
        // Find the current district
        for (const district of Object.values(locationData.properties.districts)) {
            if (district.id === currentLocationId) {
                districtData = district;
                break;
            }
        }
    }

    // If we have district data, use it; otherwise fall back to location data
    const currentData = districtData || locationData;
    logger.debug('Current data for buttons:', currentData);

    // Get connections - check both direct and properties.connections
    const connections = currentData.connections || currentData.properties?.connections;
    let buildings = currentData.buildings || currentData.properties?.buildings;

    // Fetch NPCs from backend based on current location and time
    let npcs = [];
    await fetchNPCsAtLocation(cityId, districtKey, currentBuildingId).then(npcData => {
        npcs = npcData || [];
    });

    // Check if we're inside a building
    if (currentBuildingId && buildings) {
        // Find the current building data
        const currentBuilding = buildings.find(b => b.id === currentBuildingId);

        if (currentBuilding) {
            logger.debug('Inside building:', currentBuilding);

            // Override buildings to show only "Exit Building" button
            buildings = [{ id: '__exit__', name: 'Exit Building', isExit: true }];
        }
    }

    // 1. NAVIGATION BUTTONS (D-pad style with cardinal directions)
    logger.debug('Navigation connections:', connections);
    if (connections) {
        // Clear all D-pad slots first
        ['travel-n', 'travel-s', 'travel-e', 'travel-w', 'travel-center'].forEach(slotId => {
            const slot = document.getElementById(slotId);
            if (slot) slot.innerHTML = '';
        });

        Object.entries(connections).forEach(([direction, connectionId]) => {
            logger.debug(`Processing connection: ${direction} -> ${connectionId}`);
            const connectedLocation = getLocationById(connectionId);
            logger.debug(`Found location:`, connectedLocation);

            if (connectedLocation) {
                // Map direction to D-pad slot
                const slotMap = {
                    'north': 'travel-n',
                    'south': 'travel-s',
                    'east': 'travel-e',
                    'west': 'travel-w',
                    'center': 'travel-center'
                };

                const slotId = slotMap[direction.toLowerCase()];
                const slot = document.getElementById(slotId);

                if (slot) {
                    // Determine button type based on location_type
                    const buttonType = connectedLocation.location_type === 'environment' ? 'environment' : 'navigation';

                    const button = createLocationButton(
                        connectedLocation.name || direction.toUpperCase(),
                        () => moveToLocation(connectionId),
                        buttonType
                    );
                    button.className += ' w-full h-full';
                    slot.appendChild(button);
                } else {
                    logger.warn(`No slot found for ${direction} (${slotId})`);
                }
            } else {
                logger.warn(`No location found for ${direction} -> ${connectionId}`);
            }
        });
    } else {
        logger.debug('No connections found for this location');
    }

    // 2. BUILDING BUTTONS
    if (buildingContainer) {
        const buildingButtonContainer = buildingContainer.querySelector('div');

        if (buildings && buildings.length > 0) {
            // Get current time of day from game state (convert minutes to hours)
            const timeInMinutes = state.character?.time_of_day !== undefined ? state.character.time_of_day : 720;
            const currentTime = Math.floor(timeInMinutes / 60) % 24;

            buildings.forEach(building => {
                // Check if this is the special "Exit Building" button
                if (building.isExit) {
                    const button = createLocationButton(
                        building.name,
                        () => exitBuilding(),
                        'building'  // Use green for exit button
                    );
                    buildingButtonContainer.appendChild(button);

                    // Check if player has a rented room here - add Sleep button
                    const hasRentedRoom = checkIfRoomRented(currentBuildingId);
                    if (hasRentedRoom) {
                        const sleepButton = createLocationButton(
                            'Sleep',
                            () => sleepInRoom(),
                            'action'  // Special color for action
                        );
                        buildingButtonContainer.appendChild(sleepButton);
                    }

                    // Check if player has a booked show at the right time - add Play Show button
                    const hasShow = checkIfShowReady(currentBuildingId, timeInMinutes);
                    if (hasShow) {
                        const showButton = createLocationButton(
                            'Play Show',
                            () => performShow(),
                            'action'  // Special color for action
                        );
                        buildingButtonContainer.appendChild(showButton);
                    }

                    return;
                }

                // Check if building is currently open
                const isOpen = isBuildingOpen(building, currentTime);

                if (isOpen) {
                    // Open building - normal styling
                    const button = createLocationButton(
                        building.name,
                        () => enterBuilding(building.id),
                        'building'
                    );
                    buildingButtonContainer.appendChild(button);
                } else {
                    // Closed building - grey styling with different click handler
                    const button = createLocationButton(
                        building.name,
                        () => showBuildingClosedMessage(building),
                        'building-closed'
                    );
                    buildingButtonContainer.appendChild(button);
                }
            });
        } else {
            // No buildings in this district
            const emptyMessage = document.createElement('div');
            emptyMessage.className = 'text-gray-400 text-xs p-2 text-center italic';
            emptyMessage.textContent = 'No buildings here.';
            buildingButtonContainer.appendChild(emptyMessage);
        }
    }

    // 3. NPC BUTTONS (only district-level NPCs, not building NPCs)
    if (npcContainer) {
        const npcButtonContainer = npcContainer.querySelector('div');

        if (npcs && npcs.length > 0) {
            npcs.forEach(npcId => {
                const npcData = getNPCById(npcId);
                const displayName = npcData ? npcData.name : npcId.replace(/_/g, ' ');
                const button = createLocationButton(
                    displayName,
                    () => talkToNPC(npcId),
                    'npc'
                );
                npcButtonContainer.appendChild(button);
            });
        } else {
            // Show message when no NPCs in district (they're all in buildings)
            const emptyMessage = document.createElement('div');
            emptyMessage.className = 'text-gray-400 text-xs p-2 text-center italic';
            emptyMessage.textContent = 'No one here. Check buildings.';
            npcButtonContainer.appendChild(emptyMessage);
        }
    }
}

/**
 * Create an action button with custom styling
 * @param {string} text - Button text
 * @param {Function} onClick - Click handler
 * @param {string} classes - CSS classes for styling
 * @returns {HTMLButtonElement} Button element
 */
export function createActionButton(text, onClick, classes = 'bg-gray-600 hover:bg-gray-700') {
    const button = document.createElement('button');
    button.className = `${classes} text-white px-4 py-2 rounded text-sm font-medium transition-colors`;
    button.textContent = text;
    button.addEventListener('click', onClick);
    return button;
}

/**
 * Create a location button with consistent styling
 * @param {string} text - Button text
 * @param {Function} onClick - Click handler
 * @param {string} type - Button type: 'navigation', 'environment', 'building', 'building-closed', 'npc'
 * @returns {HTMLButtonElement} Styled button element
 */
export function createLocationButton(text, onClick, type = 'navigation') {
    const button = document.createElement('button');

    // Different muted colors for different types (Win95-style)
    const typeStyles = {
        navigation: '#6b7a9e',      // City districts - muted blue
        environment: '#9e6b6b',     // Outside city - muted red
        building: '#6b8e6b',        // Open buildings - muted green
        'building-closed': '#808080', // Closed buildings - grey
        npc: '#8b6b9e',             // NPCs - muted purple
        action: '#9e8b6b'           // Special actions (sleep, play show) - muted gold
    };

    const bgColor = typeStyles[type] || typeStyles.navigation;
    const textColor = type === 'building-closed' ? '#000000' : '#ffffff';

    button.className = 'text-white transition-all leading-tight text-center overflow-hidden';
    button.style.fontSize = '7px';
    button.style.background = bgColor;
    button.style.color = textColor;
    button.style.cursor = 'pointer';
    button.style.padding = '2px 4px';
    button.style.borderTop = '1px solid #ffffff';
    button.style.borderLeft = '1px solid #ffffff';
    button.style.borderRight = '1px solid #000000';
    button.style.borderBottom = '1px solid #000000';
    button.style.boxShadow = 'inset -1px -1px 0 #404040, inset 1px 1px 0 rgba(255, 255, 255, 0.3)';
    button.style.overflowWrap = 'break-word';
    button.style.hyphens = 'none';
    button.style.display = 'flex';
    button.style.alignItems = 'center';
    button.style.justifyContent = 'center';
    button.textContent = text;
    button.addEventListener('click', () => {
        // Auto-play time when taking any action
        if (window.timeClock && window.timeClock.play) {
            window.timeClock.play();
        }
        onClick();
    });
    return button;
}

/**
 * Check if a building is currently open based on time of day
 * @param {Object} building - Building data
 * @param {number} currentTime - Current time (0-23)
 * @returns {boolean} True if building is open
 */
export function isBuildingOpen(building, currentTime) {
    // Always open buildings
    if (building.open === 'always') {
        return true;
    }

    // Private buildings (never accessible)
    if (building.open === -1 || building.open < 0) {
        return false;
    }

    const openTime = building.open;
    const closeTime = building.close;

    // No hours specified - assume always open
    if (openTime === undefined) {
        return true;
    }

    // Convert old time values (0-11) to new 24-hour format if needed
    // Old buildings may still use 0-11 values, so we need to handle both
    // In 24-hour system: buildings should use actual hours (e.g., 6 = 6 AM, 14 = 2 PM)

    // Open rest of day (close is null)
    if (closeTime === null) {
        return currentTime >= openTime;
    }

    // Check if open hours wrap around midnight
    if (openTime < closeTime) {
        // Normal hours (e.g., 6-18: 6 AM to 6 PM)
        return currentTime >= openTime && currentTime < closeTime;
    } else {
        // Overnight hours (e.g., 20-6: 8 PM through night to 6 AM)
        return currentTime >= openTime || currentTime < closeTime;
    }
}

/**
 * Show message when clicking a closed building
 * @param {Object} building - Building data
 */
export function showBuildingClosedMessage(building) {
    const openTimeName = building.open === 'always' ? 'always' : formatTime(building.open);

    showMessage(`🔒 ${building.name} is closed. Opens at ${openTimeName}.`, 'error');
}

/**
 * Enter a building
 * @param {string} buildingId - Building ID to enter
 */
export async function enterBuilding(buildingId) {
    logger.debug('Entering building:', buildingId);

    try {
        await gameAPI.sendAction('enter_building', { building_id: buildingId });
        await refreshGameState();
        await updateAllDisplays();
    } catch (error) {
        logger.error('Failed to enter building:', error);
        showMessage('❌ Failed to enter building', 'error');
    }
}

/**
 * Exit the current building
 */
export async function exitBuilding() {
    logger.debug('Exiting building');

    try {
        await gameAPI.sendAction('exit_building', {});
        await refreshGameState();
        await updateAllDisplays();
    } catch (error) {
        logger.error('Failed to exit building:', error);
        showMessage('❌ Failed to exit building', 'error');
    }
}

/**
 * Initiate dialogue with an NPC
 * @param {string} npcId - NPC ID to talk to
 */
export async function talkToNPC(npcId) {
    logger.debug('Initiating dialogue with NPC:', npcId);

    try {
        const result = await gameAPI.sendAction('talk_to_npc', { npc_id: npcId });

        if (result.success && result.delta?.npc_dialogue) {
            showNPCDialogue(result.delta.npc_dialogue, result.message);
        } else {
            showMessage('❌ Failed to talk to NPC', 'error');
        }
    } catch (error) {
        logger.error('Error talking to NPC:', error);
        showMessage('❌ Failed to talk to NPC', 'error');
    }
}

/**
 * Show NPC dialogue UI - replaces bottom UI with dialogue options
 * @param {Object} dialogueData - Dialogue data with options
 * @param {string} npcMessage - NPC message to display
 */
export function showNPCDialogue(dialogueData, npcMessage) {
    logger.debug('Showing NPC dialogue:', dialogueData);

    // Show NPC message in yellow
    if (npcMessage) {
        showMessage(npcMessage, 'warning'); // warning = yellow
    }

    // Get the action-buttons container (parent of all three columns)
    const actionButtonsArea = document.getElementById('action-buttons');
    if (!actionButtonsArea) {
        logger.error('action-buttons container not found!');
        return;
    }

    // Hide the normal action buttons
    actionButtonsArea.style.display = 'none';

    // Create or get dialogue overlay container
    let dialogueOverlay = document.getElementById('npc-dialogue-overlay');
    if (!dialogueOverlay) {
        dialogueOverlay = document.createElement('div');
        dialogueOverlay.id = 'npc-dialogue-overlay';
        dialogueOverlay.className = 'p-4 bg-gray-800 border-t-4 border-yellow-500';
        dialogueOverlay.style.height = '125px'; // Match action-buttons height

        // Insert right after action-buttons
        actionButtonsArea.parentNode.insertBefore(dialogueOverlay, actionButtonsArea.nextSibling);
    }

    // Clear previous content
    dialogueOverlay.innerHTML = '';

    // Create dialogue options grid
    const optionsGrid = document.createElement('div');
    optionsGrid.className = 'grid grid-cols-3 gap-1';

    // Add dialogue option buttons
    if (dialogueData.options && dialogueData.options.length > 0) {
        dialogueData.options.forEach(optionKey => {
            const button = document.createElement('button');
            button.className = 'text-white transition-all';
            button.style.fontSize = '7px';
            button.style.background = '#9e8b6b'; // Muted yellow/tan
            button.style.color = '#ffffff';
            button.style.cursor = 'pointer';
            button.style.padding = '2px 4px';
            button.style.borderTop = '1px solid #ffffff';
            button.style.borderLeft = '1px solid #ffffff';
            button.style.borderRight = '1px solid #000000';
            button.style.borderBottom = '1px solid #000000';
            button.style.boxShadow = 'inset -1px -1px 0 #404040, inset 1px 1px 0 rgba(255, 255, 255, 0.3)';

            // Format option text (convert snake_case to readable)
            const optionText = formatDialogueOption(optionKey);
            button.textContent = optionText;

            button.addEventListener('click', () => selectDialogueOption(dialogueData.npc_id, optionKey));
            optionsGrid.appendChild(button);
        });
    } else {
        logger.warn('No dialogue options provided!');
    }

    dialogueOverlay.appendChild(optionsGrid);
    dialogueOverlay.style.display = 'block';

    logger.debug('Dialogue overlay created with', dialogueData.options?.length || 0, 'options');
}

/**
 * Format dialogue option key to readable text
 * @param {string} optionKey - Option key (snake_case)
 * @returns {string} Readable option text
 */
export function formatDialogueOption(optionKey) {
    const optionNames = {
        'ask_about_fee': 'Ask about fee',
        'ask_about_tribute': 'Ask about tribute',
        'pay_fee': 'Pay the fee',
        'pay_tribute': 'Pay tribute',
        'use_storage': 'Use storage',
        'maybe_later': 'Maybe later',
        'goodbye': 'Goodbye'
    };

    return optionNames[optionKey] || optionKey.replace(/_/g, ' ');
}

/**
 * Handle dialogue option selection
 * @param {string} npcId - NPC ID
 * @param {string} choice - Selected dialogue option
 */
export async function selectDialogueOption(npcId, choice) {
    logger.debug('Selected dialogue option:', choice);

    try {
        const result = await gameAPI.sendAction('npc_dialogue_choice', {
            npc_id: npcId,
            choice: choice
        });

        if (result.success) {
            // Refresh game state after successful dialogue action
            await refreshGameState();
            await updateAllDisplays();

            // Check if vault should open (check this first before close action)
            if (result.delta?.open_vault) {
                logger.debug('Opening vault with data:', result.delta.open_vault);
                closeNPCDialogue();
                // Show message before opening vault
                if (result.message) {
                    showMessage(result.message, 'warning');
                }
                showVaultUI(result.delta.open_vault);
            }
            // Check if shop should open
            else if (result.delta?.open_shop) {
                logger.debug('Opening shop for merchant:', result.delta.open_shop);
                closeNPCDialogue();
                // Show message before opening shop
                if (result.message) {
                    showMessage(result.message, 'warning');
                }
                // Open shop with optional tab selection
                const shopTab = result.delta.shop_tab || 'buy';
                window.openShop(result.delta.open_shop);
                if (shopTab === 'sell') {
                    window.switchShopTab('sell');
                }
            }
            // Check if dialogue should close
            else if (result.delta?.npc_dialogue?.action === 'close') {
                closeNPCDialogue();
                // Show message when closing dialogue
                if (result.message) {
                    showMessage(result.message, 'warning');
                }
            }
            // Continue dialogue with new options
            else if (result.delta?.npc_dialogue) {
                // When continuing dialogue, only show the message once
                // The message will be shown by showNPCDialogue, so pass it there
                showNPCDialogue(result.delta.npc_dialogue, result.message);
            }
            // No delta but has message - just show the message
            else if (result.message) {
                showMessage(result.message, 'warning');
            }
        } else {
            logger.error('Dialogue option failed:', result.error);
            showMessage(result.error || 'Dialogue option failed', 'error');
        }
    } catch (error) {
        logger.error('Error selecting dialogue option:', error);
        showMessage('❌ Failed to process dialogue', 'error');
    }
}

/**
 * Close NPC dialogue and restore normal UI
 */
export function closeNPCDialogue() {
    logger.debug('Closing NPC dialogue');

    // Hide dialogue overlay
    const dialogueOverlay = document.getElementById('npc-dialogue-overlay');
    if (dialogueOverlay) {
        dialogueOverlay.style.display = 'none';
    }

    // Restore action buttons
    const actionButtonsArea = document.getElementById('action-buttons');
    if (actionButtonsArea) {
        actionButtonsArea.style.display = 'grid'; // Restore grid display
    }

    logger.debug('Dialogue closed, action buttons restored');
}

/**
 * Show vault UI overlay (40 slots over main scene)
 * @param {Object} vaultData - Vault data with slots
 */
export function showVaultUI(vaultData) {
    // Get scene container to overlay on top of it
    const sceneContainer = document.getElementById('scene-container');
    if (!sceneContainer) return;

    // Create or get vault overlay
    let vaultOverlay = document.getElementById('vault-overlay');
    if (!vaultOverlay) {
        vaultOverlay = document.createElement('div');
        vaultOverlay.id = 'vault-overlay';
        vaultOverlay.className = 'absolute inset-0 bg-black bg-opacity-90 flex items-center justify-center';
        vaultOverlay.style.zIndex = '100';
        sceneContainer.appendChild(vaultOverlay);
    }

    // Clear previous content
    vaultOverlay.innerHTML = '';

    // Create vault container
    const vaultContainer = document.createElement('div');
    vaultContainer.className = 'p-2 w-full h-full flex flex-col';

    // Header
    const header = document.createElement('div');
    header.className = 'flex justify-between items-center mb-2';

    const title = document.createElement('h2');
    title.className = 'text-yellow-400 font-bold';
    title.style.fontSize = '12px';
    title.textContent = '🏦 Vault Storage';

    const closeButton = document.createElement('button');
    closeButton.className = 'text-white px-2 py-1 font-bold';
    closeButton.style.cssText = 'background: #dc2626; border-top: 2px solid #ef4444; border-left: 2px solid #ef4444; border-right: 2px solid #991b1b; border-bottom: 2px solid #991b1b; font-size: 10px;';
    closeButton.textContent = 'Close';
    closeButton.addEventListener('click', closeVaultUI);

    header.appendChild(title);
    header.appendChild(closeButton);
    vaultContainer.appendChild(header);

    // Vault slots grid (40 slots in 8x5 grid)
    const slotsGrid = document.createElement('div');
    slotsGrid.className = 'grid grid-cols-8 gap-1 flex-1';
    slotsGrid.id = 'vault-slots-grid';
    slotsGrid.style.gridAutoRows = '1fr';

    const slots = vaultData.slots || [];
    for (let i = 0; i < 40; i++) {
        const slotData = slots[i] || { slot: i, item: null, quantity: 0 };
        const slotElement = createVaultSlot(slotData, i, vaultData.building || vaultData.location);
        slotsGrid.appendChild(slotElement);
    }

    vaultContainer.appendChild(slotsGrid);

    // Instructions
    const instructions = document.createElement('div');
    instructions.className = 'text-gray-400 text-center mt-1';
    instructions.style.fontSize = '8px';
    instructions.textContent = 'Click inventory items to store. Click vault items to withdraw.';
    vaultContainer.appendChild(instructions);

    vaultOverlay.appendChild(vaultContainer);
    vaultOverlay.style.display = 'flex';

    // Mark vault as open
    window.vaultOpen = true;

    // Force DOM to update before binding events
    requestAnimationFrame(() => {
        if (window.inventoryInteractions && window.inventoryInteractions.bindInventoryEvents) {
            window.inventoryInteractions.bindInventoryEvents();
        }
    });
}

/**
 * Create a single vault slot element (styled like backpack slots)
 * @param {Object} slotData - Slot data with item and quantity
 * @param {number} slotIndex - Slot index (0-39)
 * @param {string} buildingId - Building ID for vault
 * @returns {HTMLElement} Vault slot element
 */
export function createVaultSlot(slotData, slotIndex, buildingId) {
    const slot = document.createElement('div');
    slot.className = 'vault-slot relative cursor-pointer hover:bg-gray-800 flex items-center justify-center';
    // Match backpack slot styling exactly
    slot.style.cssText = `aspect-ratio: 1; background: #1a1a1a; border-top: 2px solid #000000; border-left: 2px solid #000000; border-right: 2px solid #3a3a3a; border-bottom: 2px solid #3a3a3a; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);`;

    // Data attributes for drag-and-drop
    slot.setAttribute('data-vault-slot', slotIndex);
    slot.setAttribute('data-vault-building', buildingId);
    slot.setAttribute('data-slot-type', 'vault');

    if (slotData.item && slotData.quantity > 0) {
        slot.setAttribute('data-item-id', slotData.item);

        // Create image container
        const imgDiv = document.createElement('div');
        imgDiv.className = 'w-full h-full flex items-center justify-center p-1';
        const img = document.createElement('img');
        img.src = `/res/img/items/${slotData.item}.png`;
        img.alt = slotData.item;
        img.className = 'w-full h-full object-contain';
        img.style.imageRendering = 'pixelated';
        img.onerror = function() {
            if (!this.dataset.fallbackAttempted) {
                this.dataset.fallbackAttempted = 'true';
                this.src = '/res/img/items/unknown.png';
            }
        };
        imgDiv.appendChild(img);
        slot.appendChild(imgDiv);

        // Add quantity label if > 1
        if (slotData.quantity > 1) {
            const quantityLabel = document.createElement('div');
            quantityLabel.className = 'absolute bottom-0 right-0 text-white';
            quantityLabel.style.fontSize = '10px';
            quantityLabel.textContent = `${slotData.quantity}`;
            slot.appendChild(quantityLabel);
        }
    }

    return slot;
}

/**
 * Close vault UI
 */
export function closeVaultUI() {
    const vaultOverlay = document.getElementById('vault-overlay');
    if (vaultOverlay) {
        vaultOverlay.style.display = 'none';
    }

    // Mark vault as closed
    window.vaultOpen = false;

    // Refresh inventory display
    // Import updateCharacterDisplay dynamically to avoid circular dependency
    import('./characterDisplay.js').then(module => {
        module.updateCharacterDisplay();
    });
}

/**
 * Check if player has a rented room at the current building
 * @param {string} buildingId - Building ID to check
 * @returns {boolean} True if room is rented here
 */
export function checkIfRoomRented(buildingId) {
    const state = getGameStateSync();

    // Try multiple possible paths for rented_rooms
    const rentedRooms = state.rented_rooms || state.character?.rented_rooms || [];
    const currentDay = state.current_day || state.character?.current_day || 0;
    const currentTime = state.time_of_day || state.character?.time_of_day || 0;

    logger.debug('Checking rented room:', { buildingId, rentedRooms, currentDay, currentTime });

    // Check if there's a valid (non-expired) rented room at this building
    for (const room of rentedRooms) {
        if (room.building === buildingId) {
            const expDay = room.expiration_day || 0;
            const expTime = room.expiration_time || 0;

            logger.debug('Found rented room:', { room, expDay, expTime });

            // Check if not expired
            if (currentDay < expDay || (currentDay === expDay && currentTime <= expTime)) {
                logger.debug('Room is valid!');
                return true;
            }
        }
    }

    logger.debug('No valid room found');
    return false;
}

/**
 * Check if player has a show ready to perform at the current building
 * @param {string} buildingId - Building ID to check
 * @param {number} timeInMinutes - Current time in minutes
 * @returns {boolean} True if show is ready to perform
 */
export function checkIfShowReady(buildingId, timeInMinutes) {
    const state = getGameStateSync();

    // Try multiple possible paths for booked_shows
    const bookedShows = state.booked_shows || state.character?.booked_shows || [];
    const currentDay = state.current_day || state.character?.current_day || 0;

    logger.debug('Checking booked show:', { buildingId, bookedShows, currentDay, timeInMinutes });

    // Check if there's an unperformed show at this venue
    for (const show of bookedShows) {
        if (show.venue_id === buildingId && show.day === currentDay && !show.performed) {
            const showTime = show.show_time || 1260; // Default 9 PM
            const timeDiff = timeInMinutes - showTime;

            logger.debug('Found booked show:', { show, showTime, timeDiff });

            // Show is ready if it's between show time and 30 minutes after
            if (timeDiff >= 0 && timeDiff <= 30) {
                logger.debug('Show is ready to perform!');
                return true;
            }
        }
    }

    logger.debug('No show ready');
    return false;
}

/**
 * Sleep in rented room
 */
export async function sleepInRoom() {
    logger.debug('Sleeping in rented room');

    try {
        const result = await gameAPI.sendAction('sleep', {});

        if (result.success) {
            showMessage(result.message || 'You wake up refreshed!', 'success');
            await refreshGameState();
            await updateAllDisplays();
        } else {
            showMessage(result.error || 'Failed to sleep', 'error');
        }
    } catch (error) {
        logger.error('Error sleeping:', error);
        showMessage('❌ Failed to sleep', 'error');
    }
}

/**
 * Perform booked show
 */
export async function performShow() {
    logger.debug('Performing booked show');

    try {
        const result = await gameAPI.sendAction('play_show', {});

        if (result.success) {
            showMessage(result.message || 'Excellent performance!', 'success');
            await refreshGameState();
            await updateAllDisplays();
        } else {
            showMessage(result.error || 'Failed to perform show', 'error');
        }
    } catch (error) {
        logger.error('Error performing show:', error);
        showMessage('❌ Failed to perform show', 'error');
    }
}

logger.debug('Location display module loaded');
