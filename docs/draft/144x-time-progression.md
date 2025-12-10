# 144x Time Progression Implementation

## Requirements
- 144x real-time speed (1 game day = 10 real minutes)
- Continuous time flow up to 12 game hours idle, then pause until next action
- 8 game hour cap per action (prevent extreme jumps)
- No offline progression (frozen when logged out)

## Implementation

### Backend: `src/api/game_actions.go`
Add to all action handlers:
1. Calculate elapsed real-time since last action
2. Convert to game segments: `segments = min((elapsed_ms * 144) / 3600000, 8)`
3. Apply 12-hour idle pause: if `segments > 12`, cap at 12
4. Call existing `handleAdvanceTimeAction(state, segments)`
5. Update last action timestamp

### Session Tracking: Store timestamp
- Add `LastActionTime time.Time` to session/state
- Initialize on session start, update on every action

### Files to Modify
- `src/api/game_actions.go` - Add time calculation to action handlers
- `src/api/saves.go` - Add LastActionTime field to SaveFile struct
- Save file format includes timestamp for persistence

## Key Logic
```
elapsed_ms = now - last_action_time
if elapsed_ms > 0:
  segments = (elapsed_ms * 144) / 3600000  // 144x multiplier
  segments = min(segments, 8)              // 8 hour cap
  if segments > 12:                        // Pause after 12 hours idle
    segments = 12
  advanceTime(segments)
  last_action_time = now
```

## No Frontend Changes Needed
- Display already handles any time value
- Auto-updates when state changes
