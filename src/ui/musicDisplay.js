/**
 * Music Display UI Module
 *
 * Handles the music tab UI rendering and interactions.
 *
 * @module ui/musicDisplay
 */

import { logger } from '../lib/logger.js';
import { getAllTracks, getCurrentTrack, getPlaybackMode, isLoopEnabled, playTrack, setPlaybackMode, toggleLoop, getVolume } from '../systems/musicSystem.js';

/**
 * Update music tab display
 */
export function updateMusicDisplay() {
    const tracks = getAllTracks();
    const currentTrack = getCurrentTrack();
    const mode = getPlaybackMode();
    const loop = isLoopEnabled();

    logger.debug('Updating music display:', {
        trackCount: tracks.length,
        currentTrack: currentTrack?.title,
        mode,
        loop
    });

    // Update mode buttons
    updateModeButtons(mode);

    // Update loop button
    updateLoopButton(loop);

    // Update track list
    updateTrackList(tracks, currentTrack);

    logger.debug('Music display updated');
}

/**
 * Update mode buttons (auto/manual)
 * @param {string} mode - Current mode ('auto' or 'manual')
 */
function updateModeButtons(mode) {
    const autoBtn = document.getElementById('music-auto-btn');
    const manualBtn = document.getElementById('music-manual-btn');

    if (!autoBtn || !manualBtn) return;

    // Reset both buttons to raised state
    autoBtn.style.borderTop = '1px solid #ffffff';
    autoBtn.style.borderLeft = '1px solid #ffffff';
    autoBtn.style.borderRight = '1px solid #000000';
    autoBtn.style.borderBottom = '1px solid #000000';
    autoBtn.style.boxShadow = 'inset -1px -1px 0 #404040, inset 1px 1px 0 rgba(255, 255, 255, 0.3)';

    manualBtn.style.borderTop = '1px solid #ffffff';
    manualBtn.style.borderLeft = '1px solid #ffffff';
    manualBtn.style.borderRight = '1px solid #000000';
    manualBtn.style.borderBottom = '1px solid #000000';
    manualBtn.style.boxShadow = 'inset -1px -1px 0 #404040, inset 1px 1px 0 rgba(255, 255, 255, 0.3)';

    // Set active button to pressed state
    if (mode === 'auto') {
        autoBtn.style.borderTop = '1px solid #000000';
        autoBtn.style.borderLeft = '1px solid #000000';
        autoBtn.style.borderRight = '1px solid #ffffff';
        autoBtn.style.borderBottom = '1px solid #ffffff';
        autoBtn.style.boxShadow = 'inset 1px 1px 0 #404040, inset -1px -1px 0 rgba(255, 255, 255, 0.3)';
    } else {
        manualBtn.style.borderTop = '1px solid #000000';
        manualBtn.style.borderLeft = '1px solid #000000';
        manualBtn.style.borderRight = '1px solid #ffffff';
        manualBtn.style.borderBottom = '1px solid #ffffff';
        manualBtn.style.boxShadow = 'inset 1px 1px 0 #404040, inset -1px -1px 0 rgba(255, 255, 255, 0.3)';
    }
}

/**
 * Update loop button
 * @param {boolean} loop - Whether loop is enabled
 */
function updateLoopButton(loop) {
    const loopBtn = document.getElementById('music-loop-btn');
    if (!loopBtn) return;

    if (loop) {
        loopBtn.textContent = 'Loop: ON';
        loopBtn.style.background = '#6b8e6b'; // Muted green
        loopBtn.style.borderTop = '1px solid #000000';
        loopBtn.style.borderLeft = '1px solid #000000';
        loopBtn.style.borderRight = '1px solid #ffffff';
        loopBtn.style.borderBottom = '1px solid #ffffff';
        loopBtn.style.boxShadow = 'inset 1px 1px 0 #404040, inset -1px -1px 0 rgba(255, 255, 255, 0.3)';
    } else {
        loopBtn.textContent = 'Loop: OFF';
        loopBtn.style.background = '#8b9aaa'; // Muted gray
        loopBtn.style.borderTop = '1px solid #ffffff';
        loopBtn.style.borderLeft = '1px solid #ffffff';
        loopBtn.style.borderRight = '1px solid #000000';
        loopBtn.style.borderBottom = '1px solid #000000';
        loopBtn.style.boxShadow = 'inset -1px -1px 0 #404040, inset 1px 1px 0 rgba(255, 255, 255, 0.3)';
    }
}

/**
 * Update track list
 * @param {Array} tracks - All tracks with unlock status
 * @param {Object|null} currentTrack - Currently playing track
 */
function updateTrackList(tracks, currentTrack) {
    const trackList = document.getElementById('music-track-list');
    if (!trackList) return;

    trackList.innerHTML = '';

    tracks.forEach(track => {
        const trackItem = document.createElement('div');
        trackItem.className = 'flex items-center justify-between p-1';
        trackItem.style.borderBottom = '1px solid #404040';

        // Track name
        const trackName = document.createElement('span');
        trackName.textContent = track.title;
        trackName.style.fontSize = '7px';
        trackName.style.cursor = track.unlocked ? 'pointer' : 'default';

        // Color based on unlock status
        if (track.unlocked) {
            trackName.style.color = '#6b8e6b'; // Green for unlocked
        } else {
            trackName.style.color = '#9e6b6b'; // Red for locked
        }

        // Highlight current track
        if (currentTrack && currentTrack.title === track.title) {
            trackName.style.fontWeight = 'bold';
            trackName.style.color = '#ffffff'; // White for currently playing
        }

        // Click handler for unlocked tracks in manual mode
        if (track.unlocked) {
            trackName.addEventListener('click', () => {
                const mode = getPlaybackMode();
                if (mode === 'manual') {
                    // Force play when manually selecting a track
                    playTrack(track, true);
                }
            });

            trackName.addEventListener('mouseenter', () => {
                if (getPlaybackMode() === 'manual') {
                    trackName.style.textDecoration = 'underline';
                }
            });

            trackName.addEventListener('mouseleave', () => {
                trackName.style.textDecoration = 'none';
            });
        }

        trackItem.appendChild(trackName);
        trackList.appendChild(trackItem);
    });
}

/**
 * Initialize music display event listeners
 */
export function initMusicDisplay() {
    logger.debug('Initializing music display');

    // Play/Pause button
    const playPauseBtn = document.getElementById('music-play-pause-btn');
    if (playPauseBtn) {
        // Update button text on play/pause toggle
        document.addEventListener('musicPlayPauseToggle', (e) => {
            playPauseBtn.textContent = e.detail.paused ? '▶' : '⏸';
        });
    }

    // Volume slider
    const volumeSlider = document.getElementById('music-volume-slider');
    if (volumeSlider) {
        // Set initial value
        const currentVolume = window.musicSystem.getVolume();
        volumeSlider.value = Math.round(currentVolume * 100);

        // Listen for volume changes from other sources
        document.addEventListener('musicVolumeChange', (e) => {
            volumeSlider.value = Math.round(e.detail.volume * 100);
        });
    }

    // Mode buttons
    const autoBtn = document.getElementById('music-auto-btn');
    const manualBtn = document.getElementById('music-manual-btn');
    const loopBtn = document.getElementById('music-loop-btn');

    if (autoBtn) {
        autoBtn.addEventListener('click', () => {
            setPlaybackMode('auto');
        });
    }

    if (manualBtn) {
        manualBtn.addEventListener('click', () => {
            setPlaybackMode('manual');
        });
    }

    if (loopBtn) {
        loopBtn.addEventListener('click', () => {
            toggleLoop();
        });
    }

    // Listen for music system events
    document.addEventListener('musicModeChange', () => updateMusicDisplay());
    document.addEventListener('musicLoopChange', () => updateMusicDisplay());
    document.addEventListener('musicTrackChange', () => updateMusicDisplay());
    document.addEventListener('musicUnlocked', () => updateMusicDisplay());

    // Initial display update
    updateMusicDisplay();

    logger.debug('Music display initialized');
}

logger.debug('Music display module loaded');
