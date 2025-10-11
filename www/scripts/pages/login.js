document.getElementById("login-button").onclick = async function () {
  if (window.nostr) {
    try {
      const publicKeyHex = await window.nostr.getPublicKey();

      // Step 1: Initialize the user session
      const initResponse = await fetch("/init-user", {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: new URLSearchParams({ publicKey: publicKeyHex }).toString(),
      });

      if (!initResponse.ok) {
        console.error("User initialization failed.");
        return;
      }
    } catch (err) {
      console.error("Login process failed:", err);
    }
  } else {
    alert("Nostr extension not available.");
  }
};

document;
document
  .getElementById("login-button")
  .addEventListener("click", async function () {
    document.getElementById("login-button").style.display = "none";
    document.getElementById("spinner").style.display = "block";
  });

// Function to show spinner when starting to load content
document.addEventListener("htmx:beforeRequest", function (event) {
  const spinnerId = event.target.getAttribute("hx-target");
  const spinner = document.querySelector(spinnerId + " .spinner");
  if (spinner) {
    spinner.style.display = "block"; // Show the spinner
  }
});

// Function to hide spinner when the content is fully loaded
document.addEventListener("htmx:afterOnLoad", function (event) {
  const spinnerId = event.target.getAttribute("hx-target");
  const spinner = document.querySelector(spinnerId + " .spinner");
  if (spinner) {
    spinner.style.display = "none"; // Hide the spinner
  }
});
