package utils

// Add utility functions here.
// For example, functions for:
// - MAC address formatting.
// - Converting byte slices to hex strings.
// - Logging helpers.

// FrequencyToChannel converts a frequency in MHz to a channel number.
// This is a simplified version and might need adjustments for different bands/regions.
func FrequencyToChannel(freqMHz int) int {
	if freqMHz >= 2412 && freqMHz <= 2472 { // 2.4 GHz band (Channels 1-13)
		return (freqMHz-2412)/5 + 1
	} else if freqMHz == 2484 { // Channel 14 for Japan
		return 14
	} else if freqMHz >= 5170 && freqMHz <= 5825 { // 5 GHz band
		// Formula for 5 GHz is (freq - 5000) / 5
		// Example: 5180 -> (5180-5000)/5 = 180/5 = 36
		// Example: 5745 -> (5745-5000)/5 = 745/5 = 149
		return (freqMHz - 5000) / 5
	} else if freqMHz >= 5955 && freqMHz <= 7115 { // 6 GHz band (Wi-Fi 6E)
		// Formula for 6 GHz (U-NII-5 to U-NII-8) is (freq - 5950) / 5 + 1 (for 20MHz channels)
		// Example: 5955 -> (5955-5950)/5 + 1 = 5/5 + 1 = 2
		// Example: 6415 -> (6415-5950)/5 + 1 = 465/5 + 1 = 93 + 1 = 94
		// This is a common way, but channel numbering can vary.
		return (freqMHz - 5950) / 5 // This might need +1 depending on convention
	}
	// Add more bands if necessary (e.g., 60 GHz)
	return 0 // Unknown or out of common Wi-Fi range
}
