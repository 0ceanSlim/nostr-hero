// ============================================================================
// BACK BUTTON COMPONENT
// Small pixelated gray button for going back in equipment selection
// ============================================================================

/**
 * Create a back button for equipment selection
 * @param {Function} onClick - Callback function when clicked
 * @returns {HTMLButtonElement} The button element
 */
function createBackButton(onClick) {
  const button = document.createElement('button');
  button.className = 'pixel-back-btn';
  button.innerHTML = '← Back';
  button.onclick = onClick;

  return button;
}
