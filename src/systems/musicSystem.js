/**
 * Music System Module
 *
 * Handles music playback with auto/manual modes and loop controls.
 * Tracks unlocked music and provides UI integration.
 *
 * @module systems/musicSystem
 */

import { logger } from '../lib/logger.js';
import { getGameStateSync } from '../state/gameState.js';

// Music system state
let currentAudio = null;
let currentTrack = null;
let playbackMode = 'auto'; // 'auto' or 'manual'
let loopEnabled = true;
let allTracks = []; // Loaded from music.json

/**
 * Initialize the music system
 * @param {Array} tracks - Music tracks from game data API
 */
export function initMusicSystem(tracks = []) {
    // Store tracks
    allTracks = tracks;

    // Load playback preferences from localStorage
    const savedMode = localStorage.getItem('musicMode');
    const savedLoop = localStorage.getItem('musicLoop');

    if (savedMode) playbackMode = savedMode;
    if (savedLoop !== null) loopEnabled = savedLoop === 'true';
}

/**
 * Get all music tracks with unlock status
 * @returns {Array} Array of tracks with unlocked status
 */
export function getAllTracks() {
    const state = getGameStateSync();
    const unlockedTracks = state.character?.music_tracks_unlocked || [];

    logger.debug('Getting all tracks:', {
        totalTracks: allTracks.length,
        unlockedTracks: unlockedTracks.length
    });

    return allTracks.map(track => ({
        ...track,
        unlocked: track.auto_unlock || unlockedTracks.includes(track.title)
    }));
}

/**
 * Get currently playing track info
 * @returns {Object|null} Current track info or null
 */
export function getCurrentTrack() {
    return currentTrack;
}

/**
 * Get playback mode
 * @returns {string} 'auto' or 'manual'
 */
export function getPlaybackMode() {
    return playbackMode;
}

/**
 * Get loop status
 * @returns {boolean} Whether loop is enabled
 */
export function isLoopEnabled() {
    return loopEnabled;
}

/**
 * Set playback mode
 * @param {string} mode - 'auto' or 'manual'
 */
export function setPlaybackMode(mode) {
    playbackMode = mode;
    localStorage.setItem('musicMode', mode);
    logger.debug('Playback mode set to:', mode);

    // Trigger mode change event
    document.dispatchEvent(new CustomEvent('musicModeChange', {
        detail: { mode, loop: loopEnabled }
    }));

    // If switching to auto, play location music
    if (mode === 'auto') {
        playLocationMusic();
    }
}

/**
 * Toggle loop on/off
 */
export function toggleLoop() {
    loopEnabled = !loopEnabled;
    localStorage.setItem('musicLoop', loopEnabled.toString());
    logger.debug('Loop toggled:', loopEnabled);

    // Update current audio if playing
    if (currentAudio) {
        currentAudio.loop = loopEnabled;
    }

    // Trigger loop change event
    document.dispatchEvent(new CustomEvent('musicLoopChange', {
        detail: { mode: playbackMode, loop: loopEnabled }
    }));
}

/**
 * Play a specific track
 * @param {Object} track - Track object from getAllTracks()
 */
export function playTrack(track) {
    if (!track.unlocked) {
        logger.warn('Cannot play locked track:', track.title);
        return;
    }

    // Don't restart if same track is already playing
    if (currentTrack && currentTrack.title === track.title && currentAudio && !currentAudio.paused) {
        logger.debug('Track already playing:', track.title);
        return;
    }

    // Stop current audio
    stopMusic();

    // Create new audio
    currentAudio = new Audio(track.file);
    currentAudio.loop = loopEnabled;
    currentAudio.volume = 0.5;
    currentTrack = track;

    // Play
    currentAudio.play().catch(err => {
        logger.debug('Music autoplay prevented:', err);
    });

    logger.debug('Playing track:', track.title);

    // Trigger track change event
    document.dispatchEvent(new CustomEvent('musicTrackChange', {
        detail: { track, mode: playbackMode, loop: loopEnabled }
    }));
}

/**
 * Play the current location's music (auto mode)
 */
export function playLocationMusic() {
    const state = getGameStateSync();
    const currentLocation = state.location?.current;

    if (!currentLocation) {
        return;
    }

    // Only play in auto mode
    if (playbackMode !== 'auto') {
        return;
    }

    // Find the track that unlocks at this location
    const locationTrack = allTracks.find(t =>
        t.unlocks_at === currentLocation &&
        (t.auto_unlock || state.character?.music_tracks_unlocked?.includes(t.title))
    );

    if (locationTrack) {
        // Add unlocked property since we already verified it should be unlocked
        const unlockedTrack = { ...locationTrack, unlocked: true };
        playTrack(unlockedTrack);
    } else {
        // Play a default track if no location-specific track
        const defaultTrack = allTracks.find(t => t.auto_unlock);
        if (defaultTrack) {
            // Add unlocked property for auto-unlock tracks
            const unlockedTrack = { ...defaultTrack, unlocked: true };
            playTrack(unlockedTrack);
        }
    }
}

/**
 * Stop current music
 */
export function stopMusic() {
    if (currentAudio) {
        currentAudio.pause();
        currentAudio.currentTime = 0;
        currentAudio = null;
    }
    currentTrack = null;
}

/**
 * Pause current music
 */
export function pauseMusic() {
    if (currentAudio) {
        currentAudio.pause();
    }
}

/**
 * Resume current music
 */
export function resumeMusic() {
    if (currentAudio) {
        currentAudio.play().catch(err => {
            logger.debug('Music play prevented:', err);
        });
    }
}

/**
 * Toggle play/pause
 */
export function togglePlayPause() {
    if (!currentAudio) {
        return;
    }

    if (currentAudio.paused) {
        resumeMusic();
    } else {
        pauseMusic();
    }

    // Trigger event to update UI
    document.dispatchEvent(new CustomEvent('musicPlayPauseToggle', {
        detail: { paused: currentAudio.paused }
    }));
}

/**
 * Check if a track should be unlocked at current location
 * and unlock it if needed
 */
export function checkAndUnlockLocationMusic() {
    const state = getGameStateSync();
    const currentLocation = state.location?.current;
    const unlockedTracks = state.character?.music_tracks_unlocked || [];

    if (!currentLocation) return;

    // Find track for this location
    const locationTrack = allTracks.find(t => t.unlocks_at === currentLocation);

    if (locationTrack && !locationTrack.auto_unlock && !unlockedTracks.includes(locationTrack.title)) {
        // Unlock the track
        state.character.music_tracks_unlocked = [...unlockedTracks, locationTrack.title];
        logger.info('Unlocked music track:', locationTrack.title);

        // Trigger unlock event
        document.dispatchEvent(new CustomEvent('musicUnlocked', {
            detail: { track: locationTrack }
        }));
    }
}

// Export for global access
window.musicSystem = {
    playTrack,
    playLocationMusic,
    stopMusic,
    pauseMusic,
    resumeMusic,
    togglePlayPause,
    setPlaybackMode,
    toggleLoop,
    getPlaybackMode,
    isLoopEnabled,
    getCurrentTrack,
    getAllTracks
};

logger.debug('Music system module loaded');
