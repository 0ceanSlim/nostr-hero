#!/bin/bash
# Quick script to find Node.js installation

echo "Searching for Node.js..."
echo ""

# Method 1: Check common locations
echo "Checking common locations:"
test -f /usr/bin/node && echo "  ✓ Found: /usr/bin/node" || echo "  ✗ Not found: /usr/bin/node"
test -f /usr/local/bin/node && echo "  ✓ Found: /usr/local/bin/node" || echo "  ✗ Not found: /usr/local/bin/node"
test -f ~/.nvm/versions/node/*/bin/node && echo "  ✓ Found: ~/.nvm/versions/node/*/bin/node" || echo "  ✗ Not found: ~/.nvm/versions/node/*/bin/node"
test -f /snap/bin/node && echo "  ✓ Found: /snap/bin/node" || echo "  ✗ Not found: /snap/bin/node"
echo ""

# Method 2: Search entire system
echo "Searching entire system (this may take a moment)..."
find /usr /opt /snap ~/.nvm 2>/dev/null -name "node" -type f | head -5
echo ""

# Method 3: Check if node is in PATH for current user
echo "Checking if 'node' works for current user:"
which node 2>/dev/null || echo "  ✗ 'node' command not found in PATH"
node --version 2>/dev/null || echo "  ✗ 'node --version' failed"
echo ""

# Method 4: Check npm
echo "Checking npm:"
which npm 2>/dev/null || echo "  ✗ 'npm' command not found in PATH"
npm --version 2>/dev/null || echo "  ✗ 'npm --version' failed"
echo ""

echo "If Node.js is not installed, install it with:"
echo "  curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -"
echo "  sudo apt-get install -y nodejs"
