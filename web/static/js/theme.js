// Function to set the theme
function setTheme(theme) {
  document.documentElement.setAttribute("data-theme", theme);
  localStorage.setItem("theme", theme);
}

// On page load, check for saved theme in localStorage
document.addEventListener("DOMContentLoaded", () => {
  const savedTheme = localStorage.getItem("theme");
  if (savedTheme) {
    setTheme(savedTheme);
  }
});

// Add click event listeners to theme buttons
document.querySelectorAll(".swatch").forEach((button) => {
  button.addEventListener("click", () => {
    const newTheme = button.dataset.theme;
    setTheme(newTheme);

    // Hide dropdown after theme selection
    document.getElementById("themeDropdown").classList.add("hidden");
  });
});
