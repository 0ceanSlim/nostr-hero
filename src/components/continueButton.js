/**
 * Continue Button Component
 *
 * Provides a reusable continue button for scenes and a helper
 * function to wait for a button click.
 *
 * @module components/continueButton
 */

/**
 * Create a continue button that appears after a delay.
 * @param {number} delay - Delay in milliseconds before button appears (default 7000).
 * @param {string} text - Button text (default "Continue →").
 * @returns {HTMLButtonElement} The button element.
 */
export function createContinueButton(delay = 7000, text = 'Continue →') {
  const button = document.createElement('button');
  button.className = 'pixel-continue-btn opacity-0 pointer-events-none';
  button.textContent = text;

  // Show button after delay
  setTimeout(() => {
    button.classList.remove('opacity-0', 'pointer-events-none');
    button.classList.add('scene-text'); // Apply fade-in animation
  }, delay);

  return button;
}

/**
 * Wait for a button click.
 * @param {HTMLButtonElement} button - The button element.
 * @returns {Promise<void>} A promise that resolves when the button is clicked.
 */
export function waitForButtonClick(button) {
  return new Promise(resolve => {
    button.onclick = resolve;
  });
}
