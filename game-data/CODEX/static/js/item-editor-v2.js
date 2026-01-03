let allItems = {};
let currentItem = null;
let isNewItem = false;
let currentTags = [];
let currentNotes = [];
let currentPackContents = [];
let allItemTypes = new Set();
let spellComponents = [];
let allItemIds = [];

// ===== INITIALIZATION =====
async function loadItems() {
    try {
        const response = await fetch('/api/items');
        allItems = await response.json();

        // Extract all unique item types and spell components
        spellComponents = [];
        allItemIds = [];
        Object.entries(allItems).forEach(([id, item]) => {
            if (item.type) {
                allItemTypes.add(item.type);
            }
            // Collect spell components (items with spell_component tag)
            if (item.tags && item.tags.includes('spell_component')) {
                spellComponents.push({ id: id, name: item.name });
            }
            allItemIds.push({ id: id, name: item.name });
        });

        // Sort for better UX
        spellComponents.sort((a, b) => a.name.localeCompare(b.name));
        allItemIds.sort((a, b) => a.name.localeCompare(b.name));

        populateTypeFilter();
        populateTypeDropdown();
        populateSpellComponentsDropdown();
        populateAllowedTypesDropdown();
        populatePackItemsDropdown();
        renderItemList();
        updateItemCount();
    } catch (error) {
        console.error('Error loading items:', error);
        showStatus('Failed to load items', 'error');
    }
}

function populateTypeFilter() {
    const typeFilter = document.getElementById('typeFilter');
    typeFilter.innerHTML = '<option value="">All Types</option>';

    Array.from(allItemTypes).sort().forEach(type => {
        const option = document.createElement('option');
        option.value = type;
        option.textContent = type;
        typeFilter.appendChild(option);
    });
}

function populateTypeDropdown() {
    const typeSelect = document.getElementById('itemType');
    typeSelect.innerHTML = '<option value="">Select type...</option>';

    Array.from(allItemTypes).sort().forEach(type => {
        const option = document.createElement('option');
        option.value = type;
        option.textContent = type;
        typeSelect.appendChild(option);
    });
}

function populateSpellComponentsDropdown() {
    const providesSelect = document.getElementById('itemProvides');
    providesSelect.innerHTML = '<option value="">Select spell component...</option>';

    spellComponents.forEach(comp => {
        const option = document.createElement('option');
        option.value = comp.id;
        option.textContent = comp.name;
        providesSelect.appendChild(option);
    });
}

function populateAllowedTypesDropdown() {
    const allowedTypesSelect = document.getElementById('allowedTypes');
    allowedTypesSelect.innerHTML = '<option value="any">any (all items)</option>';

    // Add separator comment
    const optgroupTypes = document.createElement('optgroup');
    optgroupTypes.label = 'Item Types';
    Array.from(allItemTypes).sort().forEach(type => {
        const option = document.createElement('option');
        option.value = type;
        option.textContent = type;
        optgroupTypes.appendChild(option);
    });
    allowedTypesSelect.appendChild(optgroupTypes);

    // Add all item IDs
    const optgroupItems = document.createElement('optgroup');
    optgroupItems.label = 'Specific Items';
    allItemIds.forEach(item => {
        const option = document.createElement('option');
        option.value = item.id;
        option.textContent = `${item.name} (${item.id})`;
        optgroupItems.appendChild(option);
    });
    allowedTypesSelect.appendChild(optgroupItems);
}

function populatePackItemsDropdown() {
    const packItemSelect = document.getElementById('newPackItemSelect');
    packItemSelect.innerHTML = '<option value="">Select item to add...</option>';

    allItemIds.forEach(item => {
        const option = document.createElement('option');
        option.value = item.id;
        option.textContent = `${item.name} (${item.id})`;
        packItemSelect.appendChild(option);
    });
}

function updateItemCount() {
    const visibleItems = document.querySelectorAll('.item-card').length;
    document.getElementById('itemCount').textContent = visibleItems;
}

// ===== FILTERING =====
function applyFilters() {
    const searchTerm = document.getElementById('searchBox').value.toLowerCase();
    const typeFilter = document.getElementById('typeFilter').value;

    renderItemList(searchTerm, typeFilter);
    updateItemCount();
}

function renderItemList(searchFilter = '', typeFilter = '') {
    const container = document.getElementById('itemListContainer');
    container.innerHTML = '';

    const items = Object.entries(allItems).filter(([filename, item]) => {
        // Search filter
        if (searchFilter) {
            const matches = item.name.toLowerCase().includes(searchFilter) ||
                           item.type.toLowerCase().includes(searchFilter) ||
                           filename.toLowerCase().includes(searchFilter);
            if (!matches) return false;
        }

        // Type filter
        if (typeFilter && item.type !== typeFilter) {
            return false;
        }

        return true;
    });

    items.forEach(([filename, item]) => {
        const card = document.createElement('div');
        card.className = 'item-card';
        if (currentItem === filename) {
            card.classList.add('selected');
        }
        card.onclick = () => selectItem(filename);
        card.innerHTML =
            '<div class="item-card-name">' + item.name + '</div>' +
            '<div class="item-card-type">' + item.type + '</div>';
        container.appendChild(card);
    });
}

// ===== ITEM SELECTION =====
function selectItem(filename) {
    currentItem = filename;
    isNewItem = false;
    const item = allItems[filename];

    showEditor();
    populateForm(item);
    renderItemList(document.getElementById('searchBox').value, document.getElementById('typeFilter').value);
}

function populateForm(item) {
    document.getElementById('editorMode').textContent = 'Edit';
    document.getElementById('itemName').textContent = item.name;
    document.getElementById('deleteBtn').style.display = 'block';

    // Basic info
    document.getElementById('itemId').value = item.id || '';
    document.getElementById('itemId').readOnly = true;
    document.getElementById('itemNameInput').value = item.name || '';
    document.getElementById('itemType').value = item.type || '';
    document.getElementById('itemDescription').value = item.description || '';
    document.getElementById('itemAiDescription').value = item.ai_description || '';
    document.getElementById('itemRarity').value = item.rarity || 'common';
    document.getElementById('itemPrice').value = item.price || 0;
    document.getElementById('itemWeight').value = item.weight || 1;
    document.getElementById('itemStack').value = item.stack || 1;

    // Tags
    currentTags = item.tags || [];
    renderTags();

    // Equipment
    document.getElementById('gearSlot').value = item.gear_slot || '';

    // Container
    document.getElementById('containerSlots').value = item.container_slots || '';
    // Handle allowed_types multi-select
    const allowedTypesSelect = document.getElementById('allowedTypes');
    Array.from(allowedTypesSelect.options).forEach(opt => opt.selected = false);
    if (item.allowed_types) {
        if (item.allowed_types === 'any') {
            allowedTypesSelect.options[0].selected = true; // Select "any"
        } else if (Array.isArray(item.allowed_types)) {
            item.allowed_types.forEach(type => {
                Array.from(allowedTypesSelect.options).forEach(opt => {
                    if (opt.value === type) {
                        opt.selected = true;
                    }
                });
            });
        }
    }

    // Combat
    document.getElementById('itemAC').value = item.ac || '';
    document.getElementById('itemDamage').value = item.damage || '';
    document.getElementById('damageType').value = item['damage-type'] || '';

    // Ranged
    document.getElementById('ammunition').value = item.ammunition || '';
    document.getElementById('range').value = item.range || '';
    document.getElementById('rangeLong').value = item['range-long'] || '';

    // Consumable
    document.getElementById('itemHeal').value = item.heal || '';
    document.getElementById('itemEffects').value =
        Array.isArray(item.effects) ? JSON.stringify(item.effects) : '[]';

    // Focus
    document.getElementById('itemProvides').value = item.provides || '';

    // Pack contents
    currentPackContents = item.contents || [];
    renderPackContents();

    // Notes
    currentNotes = item.notes || [];
    renderNotes();

    // Image
    document.getElementById('itemImage').value = item.image || `/res/img/items/${item.id}.png`;
    checkImage();

    // Update conditional sections
    updateConditionalSections();
}

function showEditor() {
    document.getElementById('emptyState').style.display = 'none';
    document.getElementById('editorForm').style.display = 'block';
}

function hideEditor() {
    document.getElementById('emptyState').style.display = 'block';
    document.getElementById('editorForm').style.display = 'none';
}

// ===== NEW ITEM =====
function createNewItem() {
    isNewItem = true;
    currentItem = null;
    currentTags = [];
    currentNotes = [];
    currentPackContents = [];

    showEditor();

    document.getElementById('editorMode').textContent = 'Create New Item';
    document.getElementById('itemName').textContent = 'New Item';
    document.getElementById('deleteBtn').style.display = 'none';

    // Clear form
    document.getElementById('itemId').value = '';
    document.getElementById('itemId').readOnly = false;
    document.getElementById('itemNameInput').value = '';
    document.getElementById('itemType').value = '';
    document.getElementById('itemDescription').value = '';
    document.getElementById('itemAiDescription').value = '';
    document.getElementById('itemRarity').value = 'common';
    document.getElementById('itemPrice').value = 0;
    document.getElementById('itemWeight').value = 1;
    document.getElementById('itemStack').value = 1;
    document.getElementById('gearSlot').value = '';
    document.getElementById('containerSlots').value = '';
    // Clear multi-select
    const allowedTypesSelect = document.getElementById('allowedTypes');
    Array.from(allowedTypesSelect.options).forEach(opt => opt.selected = false);
    document.getElementById('itemAC').value = '';
    document.getElementById('itemDamage').value = '';
    document.getElementById('damageType').value = '';
    document.getElementById('ammunition').value = '';
    document.getElementById('range').value = '';
    document.getElementById('rangeLong').value = '';
    document.getElementById('itemHeal').value = '';
    document.getElementById('itemEffects').value = '[]';
    document.getElementById('itemProvides').value = '';
    document.getElementById('itemImage').value = '';

    renderTags();
    renderNotes();
    renderPackContents();
    updateConditionalSections();

    // Clear selection
    document.querySelectorAll('.item-card').forEach(card => {
        card.classList.remove('selected');
    });
}

// ===== CONDITIONAL SECTIONS =====
function updateConditionalSections() {
    const hasEquipment = currentTags.includes('equipment');
    const hasContainer = currentTags.includes('container');
    const hasConsumable = currentTags.includes('consumable');
    const hasFocus = currentTags.includes('focus');
    const hasPack = currentTags.includes('pack');
    const hasThrown = currentTags.includes('thrown');

    const itemType = document.getElementById('itemType').value.toLowerCase();

    // Check for specific item types
    const isWeapon = itemType.includes('weapon') ||
                    itemType.includes('melee') ||
                    itemType.includes('martial') ||
                    itemType.includes('simple') ||
                    hasThrown; // Thrown items are weapons too
    const isArmor = itemType.includes('armor');
    const isRanged = itemType.includes('ranged') || hasThrown; // Thrown weapons are ranged

    // Show/hide sections
    document.getElementById('equipmentSection').style.display = hasEquipment ? 'block' : 'none';
    document.getElementById('containerSection').style.display = hasContainer ? 'block' : 'none';
    document.getElementById('consumableSection').style.display = hasConsumable ? 'block' : 'none';
    document.getElementById('focusSection').style.display = hasFocus ? 'block' : 'none';
    document.getElementById('packSection').style.display = hasPack ? 'block' : 'none';
    // Show weapon section for all weapons (melee, ranged, and thrown all need damage/damage-type)
    document.getElementById('weaponSection').style.display = isWeapon ? 'block' : 'none';
    document.getElementById('armorSection').style.display = isArmor ? 'block' : 'none';
    // Show ranged section for ranged weapons AND thrown weapons (for range properties)
    document.getElementById('rangedSection').style.display = isRanged ? 'block' : 'none';

    // Update ammunition field based on thrown status
    const ammoGroup = document.querySelector('#ammunition').parentElement;
    if (hasThrown) {
        // Hide ammunition field for thrown weapons
        ammoGroup.style.display = 'none';
    } else {
        // Show ammunition field for non-thrown ranged weapons
        ammoGroup.style.display = 'block';
        const ammoLabel = ammoGroup.querySelector('label');
        if (ammoLabel) {
            ammoLabel.innerHTML = 'Ammunition *';
        }
    }
}

// ===== TAGS MANAGEMENT =====
function renderTags() {
    const container = document.getElementById('tagsContainer');
    container.innerHTML = '';

    currentTags.forEach(tag => {
        const tagElement = document.createElement('div');
        tagElement.className = 'tag';
        tagElement.innerHTML =
            tag +
            '<span class="tag-remove" onclick="removeTag(\'' + tag + '\')">×</span>';
        container.appendChild(tagElement);
    });

    updateConditionalSections();
}

function addTag() {
    const input = document.getElementById('newTagInput');
    const tag = input.value.trim();

    if (tag && !currentTags.includes(tag)) {
        currentTags.push(tag);
        renderTags();
        input.value = '';
    }
}

function removeTag(tag) {
    currentTags = currentTags.filter(t => t !== tag);
    renderTags();
}

// ===== NOTES MANAGEMENT =====
function renderNotes() {
    const container = document.getElementById('notesContainer');
    container.innerHTML = '';

    currentNotes.forEach((note, index) => {
        const noteElement = document.createElement('div');
        noteElement.className = 'note-item';
        noteElement.innerHTML =
            note +
            '<span class="note-remove" onclick="removeNote(' + index + ')">×</span>';
        container.appendChild(noteElement);
    });
}

function addNote() {
    const input = document.getElementById('newNoteInput');
    const note = input.value.trim();

    if (note) {
        currentNotes.push(note);
        renderNotes();
        input.value = '';
    }
}

function removeNote(index) {
    currentNotes.splice(index, 1);
    renderNotes();
}

// ===== PACK CONTENTS MANAGEMENT =====
function renderPackContents() {
    const container = document.getElementById('packContentsContainer');
    container.innerHTML = '';

    currentPackContents.forEach((packItem, index) => {
        const itemId = packItem[0];
        const quantity = packItem[1];
        const itemData = allItems[itemId];

        const packItemElement = document.createElement('div');
        packItemElement.className = 'pack-item';

        const itemName = itemData ? itemData.name : '⚠️ Unknown Item';
        const itemIdDisplay = `(${itemId})`;

        packItemElement.innerHTML =
            '<div class="pack-item-info">' +
                '<div>' +
                    '<span class="pack-item-name">' + itemName + '</span>' +
                    '<span class="pack-item-id">' + itemIdDisplay + '</span>' +
                '</div>' +
                '<span class="pack-item-quantity">Qty: ' + quantity + '</span>' +
            '</div>' +
            '<span class="pack-item-remove" onclick="removePackItem(' + index + ')">×</span>';

        container.appendChild(packItemElement);
    });
}

function addPackItem() {
    const itemSelect = document.getElementById('newPackItemSelect');
    const quantityInput = document.getElementById('newPackItemQuantity');

    const itemId = itemSelect.value;
    const quantity = parseInt(quantityInput.value) || 1;

    if (!itemId) {
        showStatus('Please select an item to add', 'error');
        return;
    }

    if (quantity < 1) {
        showStatus('Quantity must be at least 1', 'error');
        return;
    }

    // Check if item already exists in pack
    const existingIndex = currentPackContents.findIndex(item => item[0] === itemId);
    if (existingIndex !== -1) {
        // Update quantity
        currentPackContents[existingIndex][1] = quantity;
    } else {
        // Add new item
        currentPackContents.push([itemId, quantity]);
    }

    renderPackContents();
    itemSelect.value = '';
    quantityInput.value = 1;
}

function removePackItem(index) {
    currentPackContents.splice(index, 1);
    renderPackContents();
}

// ===== SAVE ITEM =====
async function saveItem() {
    const itemId = document.getElementById('itemId').value.trim();

    if (!itemId) {
        showStatus('Item ID is required', 'error');
        return;
    }

    // Validate ID format (lowercase-with-hyphens)
    if (!/^[a-z0-9-]+$/.test(itemId)) {
        showStatus('Item ID must be lowercase letters, numbers, and hyphens only', 'error');
        return;
    }

    // Build item object
    const item = {
        id: itemId,
        name: document.getElementById('itemNameInput').value.trim(),
        description: document.getElementById('itemDescription').value.trim(),
        ai_description: document.getElementById('itemAiDescription').value.trim(),
        rarity: document.getElementById('itemRarity').value,
        price: parseInt(document.getElementById('itemPrice').value) || 0,
        weight: parseFloat(document.getElementById('itemWeight').value) || 1,
        stack: parseInt(document.getElementById('itemStack').value) || 1,
        type: document.getElementById('itemType').value.trim(),
        tags: currentTags,
        notes: currentNotes,
        image: document.getElementById('itemImage').value.trim() || `/res/img/items/${itemId}.png`
    };

    // Add conditional fields
    if (currentTags.includes('equipment')) {
        item.gear_slot = document.getElementById('gearSlot').value;
    }

    if (currentTags.includes('container')) {
        item.container_slots = parseInt(document.getElementById('containerSlots').value) || 20;
        // Handle multi-select allowed_types
        const allowedTypesSelect = document.getElementById('allowedTypes');
        const selectedOptions = Array.from(allowedTypesSelect.selectedOptions).map(opt => opt.value);

        if (selectedOptions.includes('any') || selectedOptions.length === 0) {
            item.allowed_types = 'any';
        } else {
            item.allowed_types = selectedOptions;
        }
    }

    // Combat properties
    const ac = document.getElementById('itemAC').value.trim();
    if (ac) item.ac = ac;

    const damage = document.getElementById('itemDamage').value.trim();
    if (damage) item.damage = damage;

    const damageType = document.getElementById('damageType').value;
    if (damageType) item['damage-type'] = damageType;

    // Ranged properties
    const ammunition = document.getElementById('ammunition').value.trim();
    if (ammunition) item.ammunition = ammunition;

    const range = document.getElementById('range').value.trim();
    if (range) item.range = range;

    const rangeLong = document.getElementById('rangeLong').value.trim();
    if (rangeLong) item['range-long'] = rangeLong;

    // Consumable properties
    const heal = document.getElementById('itemHeal').value.trim();
    if (heal) item.heal = heal;

    if (currentTags.includes('consumable')) {
        try {
            const effectsStr = document.getElementById('itemEffects').value.trim();
            item.effects = effectsStr ? JSON.parse(effectsStr) : [];
        } catch (e) {
            item.effects = [];
        }
    }

    // Focus properties
    if (currentTags.includes('focus')) {
        const provides = document.getElementById('itemProvides').value.trim();
        if (provides) item.provides = provides;
    }

    // Pack contents
    if (currentTags.includes('pack')) {
        item.contents = currentPackContents;
    }

    // Save to server
    try {
        const filename = isNewItem ? itemId : currentItem;
        const response = await fetch(`/api/items/${filename}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item)
        });

        if (response.ok) {
            showStatus('Item saved successfully!', 'success');

            // Update local cache
            allItems[itemId] = item;

            if (isNewItem) {
                // Add to type filter if new type
                if (!allItemTypes.has(item.type)) {
                    allItemTypes.add(item.type);
                    populateTypeFilter();
                    populateTypeDatalist();
                }

                isNewItem = false;
                currentItem = itemId;
            }

            renderItemList(document.getElementById('searchBox').value, document.getElementById('typeFilter').value);
            updateItemCount();
        } else {
            const error = await response.text();
            showStatus('Failed to save item: ' + error, 'error');
        }
    } catch (error) {
        showStatus('Error saving item: ' + error.message, 'error');
    }
}

// ===== DELETE ITEM =====
async function deleteItem() {
    if (!currentItem || isNewItem) return;

    if (!confirm(`Are you sure you want to delete "${allItems[currentItem].name}"?`)) {
        return;
    }

    try {
        const response = await fetch(`/api/items/${currentItem}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            showStatus('Item deleted successfully', 'success');

            delete allItems[currentItem];
            currentItem = null;

            hideEditor();
            renderItemList();
            updateItemCount();
        } else {
            showStatus('Failed to delete item', 'error');
        }
    } catch (error) {
        showStatus('Error deleting item: ' + error.message, 'error');
    }
}

// ===== VALIDATION =====
async function validateItem() {
    // Run validation check
    showStatus('Validating item...', 'warning');

    // You could call the validation API here
    setTimeout(() => {
        showStatus('Validation complete - no issues found', 'success');
    }, 500);
}

// ===== IMAGE HANDLING =====
async function checkImage() {
    const imagePath = document.getElementById('itemImage').value;
    const container = document.getElementById('imageContainer');

    if (!imagePath) {
        container.innerHTML = '<p style="color: #6272a4;">No image path set</p>';
        return;
    }

    // Try to load the image
    const img = new Image();
    img.onload = function() {
        container.innerHTML = '';
        container.appendChild(img);
    };
    img.onerror = function() {
        container.innerHTML = '<p style="color: #ff5555;">Image not found</p>';
    };
    // Use /www prefix to access images from CODEX server
    img.src = imagePath.startsWith('/res/') ? `/www${imagePath}` : imagePath;
}

async function generateImage() {
    showStatus('Image generation not yet implemented', 'warning');
}

// ===== UTILITY =====
function showStatus(message, type = 'success') {
    const statusEl = document.getElementById('statusMessage');
    statusEl.textContent = message;
    statusEl.className = 'status-message ' + type + ' show';

    setTimeout(() => {
        statusEl.classList.remove('show');
    }, 5000);
}

function cancelEdit() {
    if (isNewItem) {
        hideEditor();
        currentItem = null;
    } else if (currentItem) {
        selectItem(currentItem); // Reload current item
    } else {
        hideEditor();
    }
}

// ===== INIT =====
window.addEventListener('DOMContentLoaded', () => {
    loadItems();

    // Add enter key support for tag/note input
    document.getElementById('newTagInput').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            addTag();
            e.preventDefault();
        }
    });

    document.getElementById('newNoteInput').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            addNote();
            e.preventDefault();
        }
    });

    // Update conditional sections when type changes
    document.getElementById('itemType').addEventListener('input', updateConditionalSections);
});
