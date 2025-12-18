/**
 * Back Button Component
 *
 * Provides a reusable back button component.
 *
 * @module components/backButton
 */

/**
 * Create a back button for equipment selection.
 * @param {Function} onClick - Callback function when clicked.
 * @returns {HTMLButtonElement} The button element.
 */
export function createBackButton(onClick) {
  const button = document.createElement('button');
  button.className = 'pixel-back-btn';
  button.innerHTML = '← Back';
  button.onclick = onClick;

  return button;
}
