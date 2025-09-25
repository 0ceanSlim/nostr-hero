// Nostr Hero Startup Sequence
// Comprehensive initialization following gnostream patterns

class NostrHeroStartup {
    constructor() {
        this.initializationSteps = [
            { name: 'Session Manager', fn: this.initSessionManager },
            { name: 'Authentication', fn: this.initAuthentication },
            { name: 'Game Systems', fn: this.initGameSystems },
            { name: 'UI Components', fn: this.initUIComponents }
        ];
        this.currentStep = 0;
        this.isInitialized = false;
    }

    async initialize() {
        console.log('🎮 Starting Nostr Hero initialization sequence...');

        try {
            for (let i = 0; i < this.initializationSteps.length; i++) {
                const step = this.initializationSteps[i];
                this.currentStep = i;

                console.log(`🔄 Step ${i + 1}/${this.initializationSteps.length}: ${step.name}`);
                this.updateLoadingIndicator(step.name, (i + 1) / this.initializationSteps.length);

                await step.fn.call(this);
                console.log(`✅ Step ${i + 1} complete: ${step.name}`);
            }

            this.isInitialized = true;
            console.log('🎉 Nostr Hero initialization complete!');
            this.onInitializationComplete();

        } catch (error) {
            console.error('❌ Initialization failed at step:', this.initializationSteps[this.currentStep]?.name, error);
            this.onInitializationFailed(error);
        }
    }

    async initSessionManager() {
        return new Promise((resolve, reject) => {
            const checkSessionManager = () => {
                if (window.sessionManager) {
                    console.log('✅ SessionManager ready');
                    resolve();
                } else {
                    console.log('⏳ Waiting for SessionManager...');
                    setTimeout(checkSessionManager, 50);
                }
            };
            checkSessionManager();

            // Timeout after 5 seconds
            setTimeout(() => {
                if (!window.sessionManager) {
                    reject(new Error('SessionManager failed to load within 5 seconds'));
                }
            }, 5000);
        });
    }

    async initAuthentication() {
        if (typeof initializeAuthentication === 'function') {
            initializeAuthentication();
            console.log('✅ Authentication system initialized');
        } else {
            throw new Error('Authentication initialization function not found');
        }
    }

    async initGameSystems() {
        // Game systems are initialized when the game starts after authentication
        // Just verify the functions are available
        const requiredGameFunctions = [
            'getGameState',
            'updateGameState',
            'initializeGame',
            'showMessage'
        ];

        for (const funcName of requiredGameFunctions) {
            if (typeof window[funcName] !== 'function') {
                throw new Error(`Required game function ${funcName} not found`);
            }
        }

        console.log('✅ Game systems ready');
    }

    async initUIComponents() {
        // Check that required DOM elements exist
        const requiredElements = [
            'game-app'
        ];

        for (const elementId of requiredElements) {
            const element = document.getElementById(elementId);
            if (!element) {
                throw new Error(`Required DOM element #${elementId} not found`);
            }
        }

        // Initialize UI event listeners
        this.setupGlobalEventListeners();
        console.log('✅ UI components initialized');
    }

    setupGlobalEventListeners() {
        // Global error handler
        window.addEventListener('error', (event) => {
            console.error('🚨 Global error:', event.error);
            showMessage('❌ An error occurred: ' + event.error.message, 'error');
        });

        // Unhandled promise rejection handler
        window.addEventListener('unhandledrejection', (event) => {
            console.error('🚨 Unhandled promise rejection:', event.reason);
            showMessage('❌ An error occurred: ' + (event.reason?.message || event.reason), 'error');
        });

        // Session storage events
        window.addEventListener('storage', (event) => {
            if (event.key === 'nostr_hero_session_meta') {
                console.log('📡 Session storage changed, refreshing session');
                if (window.sessionManager) {
                    window.sessionManager.checkExistingSession();
                }
            }
        });

        // Visibility change handler for session monitoring
        document.addEventListener('visibilitychange', () => {
            if (!document.hidden && window.sessionManager) {
                // Check session when tab becomes visible
                window.sessionManager.checkExistingSession();
            }
        });

        // Before unload handler for cleanup
        window.addEventListener('beforeunload', () => {
            console.log('🧹 Cleaning up before page unload');
            // Any cleanup logic here
        });
    }

    updateLoadingIndicator(stepName, progress) {
        const gameContainer = document.getElementById('game-app');
        if (gameContainer && !this.isInitialized) {
            const progressPercent = Math.round(progress * 100);
            gameContainer.innerHTML = `
                <div class="flex items-center justify-center min-h-screen">
                    <div class="text-center">
                        <div class="mb-8">
                            <h1 class="text-4xl font-bold text-yellow-400 mb-2">⚔️ Nostr Hero ⚔️</h1>
                            <p class="text-gray-400">A decentralized text-based RPG</p>
                        </div>

                        <div class="mb-6">
                            <div class="w-64 bg-gray-700 rounded-full h-2 mx-auto mb-2">
                                <div class="bg-yellow-400 h-2 rounded-full transition-all duration-300"
                                     style="width: ${progressPercent}%"></div>
                            </div>
                            <p class="text-sm text-gray-300">${progressPercent}% - ${stepName}</p>
                        </div>

                        <div class="text-xs text-gray-500">
                            <p>🔐 Powered by Nostr</p>
                            <p>🌐 Your keys, your character</p>
                        </div>
                    </div>
                </div>
            `;
        }
    }

    onInitializationComplete() {
        // Clear any loading indicators
        this.hideLoadingIndicator();

        // Emit initialization complete event
        window.dispatchEvent(new CustomEvent('nostrHeroReady', {
            detail: { timestamp: Date.now() }
        }));

        console.log('🎮 Nostr Hero is ready to play!');
    }

    onInitializationFailed(error) {
        const gameContainer = document.getElementById('game-app');
        if (gameContainer) {
            gameContainer.innerHTML = `
                <div class="flex items-center justify-center min-h-screen">
                    <div class="text-center max-w-md mx-auto p-6">
                        <div class="mb-6">
                            <h1 class="text-4xl font-bold text-red-400 mb-2">⚠️ Initialization Failed</h1>
                            <p class="text-gray-400 mb-4">Failed to start Nostr Hero</p>
                        </div>

                        <div class="bg-red-900 bg-opacity-50 border border-red-600 rounded-lg p-4 mb-6">
                            <p class="text-red-200 text-sm">${error.message}</p>
                        </div>

                        <button onclick="window.location.reload()"
                                class="bg-yellow-600 hover:bg-yellow-700 text-gray-900 px-6 py-3 rounded-lg font-medium">
                            🔄 Retry
                        </button>

                        <div class="mt-6 text-xs text-gray-500">
                            <p>If this problem persists, please check:</p>
                            <ul class="mt-2 text-left">
                                <li>• JavaScript is enabled</li>
                                <li>• No browser extensions are blocking scripts</li>
                                <li>• Your internet connection is stable</li>
                            </ul>
                        </div>
                    </div>
                </div>
            `;
        }

        console.error('💥 Nostr Hero initialization failed:', error);
    }

    hideLoadingIndicator() {
        // Loading indicator will be replaced by game interface or login interface
        // This is handled by the authentication system
    }

    // Public API
    isReady() {
        return this.isInitialized;
    }

    getCurrentStep() {
        return this.currentStep;
    }

    getTotalSteps() {
        return this.initializationSteps.length;
    }
}

// Create global startup instance
window.nostrHeroStartup = new NostrHeroStartup();

// Auto-initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    console.log('📱 DOM loaded, starting Nostr Hero initialization...');
    window.nostrHeroStartup.initialize();
});

// Public API for checking if the game is ready
window.isNostrHeroReady = function() {
    return window.nostrHeroStartup?.isReady() || false;
};

console.log('🚀 Nostr Hero startup system loaded');