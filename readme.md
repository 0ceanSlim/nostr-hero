# ⚔️ Nostr Hero

## Overview

Nostr Hero is a web application that derives a unique RPG-style character from your Nostr public key. Using a deterministic algorithm, the app transforms your identity into a digital champion with a race, class, background, alignment, and stats. This allows for a consistent and reproducible hero based on your cryptographic identity.

### How It Works

#### Go Backend

Runs the web server and serves HTML/CSS/JS files.

Processes Nostr public keys to deterministically generate character attributes.

Stores and serves character data via JSON.

#### HTMX Frontend

Fetches character data dynamically from the Go backend.

Uses AJAX requests to update the UI without full page reloads.

#### Hyperscript Enhancements

Provides lightweight client-side interactivity.

Handles UI interactions efficiently.

#### TailwindCSS Styling

Ensures a clean, responsive, and modern UI.

### Features

Deterministic Character Generation: Every Nostr key produces the same hero.

Race & Class Weighing: Different races and classes have unique distributions.

Background & Alignment System: Provides a backstory and moral compass.

D&D-Style Stats: Randomized yet balanced character stats.

Dynamic Frontend: No need for full page reloads thanks to HTMX.

### Contributing

Feel free to fork this repo, submit issues, or suggest improvements!

### License

This project is Open Source and licensed under the MIT License. See the [LICENSE](license) file for details.

### Acknowledgments

Special thanks to the Nostr community for their continuous support and contributions.

Open Source and made with 💦 by [OceanSlim](https://njump.me/npub1zmc6qyqdfnllhnzzxr5wpepfpnzcf8q6m3jdveflmgruqvd3qa9sjv7f60)