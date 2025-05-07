// This file previously contained WebSocket connection logic.
// It is no longer needed as Wails uses its own JS Bridge and event system
// for frontend-backend communication.

// You can remove imports of this file from other components.
// Communication logic should be updated to use Wails runtime functions:
// - `window.go.main.App.*` for calling Go methods.
// - `runtime.EventsOn()` for listening to Go events.

export {}; // Add this line if the file becomes empty to satisfy TypeScript module requirements