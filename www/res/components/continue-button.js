// ============================================================================
// CONTINUE BUTTON COMPONENT
// Pixelated yellow button with 7-second delay
// ============================================================================

/**
 * Create a continue button that appears after a delay
 * @param {number} delay - Delay in milliseconds before button appears (default 7000)
 * @param {string} text - Button text (default "Continue →")
 * @returns {HTMLButtonElement} The button element
 */
function createContinueButton(delay = 7000, text = 'Continue →') {
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
 * Wait for button click
 * @param {HTMLButtonElement} button - The button element
 * @returns {Promise<void>}
 */
function waitForButtonClick(button) {
  return new Promise(resolve => {
    button.onclick = resolve;
  });
}
