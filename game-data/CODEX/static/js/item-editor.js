let allItems = {};
let currentItem = null;
let refactorPreviewData = null;

async function loadItems() {
    try {
        const response = await fetch('/api/items');
        allItems = await response.json();
        renderItemList();
    } catch (error) {
        console.error('Error loading items:', error);
    }
}

function renderItemList(filter = '') {
    const container = document.getElementById('itemListContainer');
    container.innerHTML = '';

    const items = Object.entries(allItems).filter(([filename, item]) => {
        if (!filter) return true;
        const searchLower = filter.toLowerCase();
        return item.name.toLowerCase().includes(searchLower) ||
               item.type.toLowerCase().includes(searchLower) ||
               filename.toLowerCase().includes(searchLower);
    });

    items.forEach(([filename, item]) => {
        const card = document.createElement('div');
        card.className = 'item-card';
        if (currentItem === filename) {
            card.classList.add('selected');
        }
        card.onclick = () => selectItem(filename);
        card.innerHTML = '<div class="item-card-name">' + item.name + '</div>' +
                         '<div class="item-card-type">' + item.type + '</div>';
        container.appendChild(card);
    });
}

function selectItem(filename) {
    currentItem = filename;
    const item = allItems[filename];

    document.getElementById('emptyState').style.display = 'none';
    document.getElementById('editorForm').style.display = 'block';

    document.getElementById('itemName').textContent = item.name;
    document.getElementById('itemId').value = item.id;
    document.getElementById('itemNameInput').value = item.name;
    document.getElementById('itemType').value = item.type || '';
    document.getElementById('itemPrice').value = item.price || 0;
    document.getElementById('itemWeight').value = item.weight || 0;
    document.getElementById('itemStack').value = item.stack || 1;
    document.getElementById('itemRarity').value = item.rarity || 'common';
    document.getElementById('itemDescription').value = item.description || '';
    document.getElementById('itemAiDescription').value = item.ai_description || '';

    renderItemList();
    checkImage();
}

async function saveItem() {
    if (!currentItem) return;

    const item = {
        id: document.getElementById('itemId').value,
        name: document.getElementById('itemNameInput').value,
        type: document.getElementById('itemType').value,
        price: parseInt(document.getElementById('itemPrice').value) || 0,
        weight: parseFloat(document.getElementById('itemWeight').value) || 0,
        stack: parseInt(document.getElementById('itemStack').value) || 1,
        rarity: document.getElementById('itemRarity').value,
        description: document.getElementById('itemDescription').value,
        ai_description: document.getElementById('itemAiDescription').value
    };

    try {
        const response = await fetch('/api/items/' + currentItem, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item)
        });

        const result = await response.json();

        if (result.status === 'success') {
            showStatus('Item saved successfully!', 'success');
            allItems[currentItem] = item;
            renderItemList();
        } else {
            showStatus('Error saving item', 'error');
        }
    } catch (error) {
        showStatus('Error: ' + error.message, 'error');
    }
}

async function checkImage() {
    if (!currentItem) return;

    try {
        const response = await fetch('/api/items/' + currentItem + '/image');
        const result = await response.json();

        const container = document.getElementById('imageContainer');
        if (result.exists) {
            container.innerHTML = '<img src="../../www/res/img/items/' + allItems[currentItem].id + '.png" alt="Item image" />';
        } else {
            container.innerHTML = '<p style="color: #6272a4;">No image available</p>';
        }
    } catch (error) {
        console.error('Error checking image:', error);
    }
}

async function generateImage() {
    if (!currentItem) return;

    showStatus('Generating image...', 'success');

    try {
        const response = await fetch('/api/items/' + currentItem + '/generate-image', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ model: 'bitforge' })
        });

        const result = await response.json();

        if (result.success) {
            showStatus('Image generated! Cost: $' + result.cost.toFixed(4), 'success');
            const container = document.getElementById('imageContainer');
            container.innerHTML = '<img src="data:image/png;base64,' + result.imageData + '" alt="Generated image" />' +
                                  '<div style="margin-top: 10px;">' +
                                  '<button class="btn btn-primary" onclick="acceptImage(\'' + result.imageData + '\')">Accept & Save</button>' +
                                  '</div>';
        } else {
            showStatus('Error generating image', 'error');
        }
    } catch (error) {
        showStatus('Error: ' + error.message, 'error');
    }
}

async function acceptImage(imageData) {
    if (!currentItem) return;

    try {
        const response = await fetch('/api/items/' + currentItem + '/accept-image', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ imageData: imageData })
        });

        const result = await response.json();

        if (result.success) {
            showStatus('Image accepted and saved!', 'success');
            checkImage();
        } else {
            showStatus('Error accepting image', 'error');
        }
    } catch (error) {
        showStatus('Error: ' + error.message, 'error');
    }
}

async function previewRefactor() {
    if (!currentItem) return;

    const newId = document.getElementById('newItemId').value.trim();
    if (!newId) {
        showStatus('Please enter a new ID', 'error');
        return;
    }

    try {
        const response = await fetch('/api/refactor/preview', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                filename: currentItem,
                oldId: allItems[currentItem].id,
                newId: newId
            })
        });

        refactorPreviewData = await response.json();

        const previewDiv = document.getElementById('refactorPreview');
        const refList = document.getElementById('referenceList');

        if (refactorPreviewData.references.length === 0) {
            refList.innerHTML = '<p style="color: #6272a4;">No references found</p>';
        } else {
            refList.innerHTML = '';
            refactorPreviewData.references.forEach(ref => {
                const item = document.createElement('div');
                item.className = 'reference-item';
                item.textContent = ref.file + ' → ' + ref.location;
                refList.appendChild(item);
            });
        }

        previewDiv.style.display = 'block';
        document.getElementById('applyRefactorBtn').disabled = false;
        showStatus('Found ' + refactorPreviewData.references.length + ' references', 'success');
    } catch (error) {
        showStatus('Error: ' + error.message, 'error');
    }
}

async function applyRefactor() {
    if (!refactorPreviewData) return;

    if (!confirm('This will rename the item and update all references. Continue?')) {
        return;
    }

    try {
        const response = await fetch('/api/refactor/apply', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                filename: currentItem,
                preview: refactorPreviewData
            })
        });

        const result = await response.json();

        if (result.status === 'success') {
            showStatus('Refactoring complete! Reloading items...', 'success');
            setTimeout(() => {
                window.location.reload();
            }, 1500);
        } else {
            showStatus('Error applying refactor', 'error');
        }
    } catch (error) {
        showStatus('Error: ' + error.message, 'error');
    }
}

function showStatus(message, type) {
    const status = document.getElementById('statusMessage');
    status.textContent = message;
    status.className = 'status-message ' + type;
    status.style.display = 'block';
    setTimeout(() => {
        status.style.display = 'none';
    }, 5000);
}

document.getElementById('searchBox').addEventListener('input', (e) => {
    renderItemList(e.target.value);
});

loadItems();
