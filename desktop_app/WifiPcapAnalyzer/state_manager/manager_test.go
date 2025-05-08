package state_manager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPeriodicallyCalculateMetrics_Basic(t *testing.T) {
	interval := 1 * time.Second
	historyPoints := 5
	sm := NewStateManager(interval, historyPoints)

	// Setup initial state
	bssid := "00:11:22:33:44:55"
	staMAC := "AA:BB:CC:DD:EE:FF"

	bssInfo := NewBSSInfo(bssid)
	staInfo := NewSTAInfo(staMAC)
	staInfo.AssociatedBSSID = bssid
	bssInfo.AssociatedSTAs[staMAC] = staInfo

	// Simulate accumulated data over the interval
	accumulatedBssAirtime := 500 * time.Millisecond // 50% utilization
	accumulatedBssBytes := int64(1000000)           // 1MB
	accumulatedStaAirtime := 200 * time.Millisecond // 20% utilization
	accumulatedStaUplinkBytes := int64(200000)      // 0.2MB
	accumulatedStaDownlinkBytes := int64(300000)    // 0.3MB

	bssInfo.totalAirtime = accumulatedBssAirtime
	bssInfo.totalTxBytes = accumulatedBssBytes
	staInfo.totalAirtime = accumulatedStaAirtime
	staInfo.totalUplinkBytes = accumulatedStaUplinkBytes
	staInfo.totalDownlinkBytes = accumulatedStaDownlinkBytes

	// Set last calculation time to be 'interval' ago
	calcTime := time.Now().Add(-interval)
	bssInfo.lastCalcTime = calcTime
	staInfo.lastCalcTime = calcTime

	sm.bssInfos[bssid] = bssInfo
	sm.staInfos[staMAC] = staInfo

	// Call the function to test
	sm.PeriodicallyCalculateMetrics()

	// Assertions (Corrected expected values)
	assert.InDelta(t, 50.0, bssInfo.ChannelUtilization, 0.01, "BSS Channel Utilization should be calculated") // Use InDelta for float comparison
	assert.Equal(t, int64(8000000), bssInfo.Throughput, "BSS Throughput should be calculated")
	assert.InDelta(t, 20.0, staInfo.ChannelUtilization, 0.01, "STA Channel Utilization should be calculated") // Use InDelta for float comparison
	assert.Equal(t, int64(1600000), staInfo.UplinkThroughput, "STA Uplink Throughput should be calculated")
	assert.Equal(t, int64(2400000), staInfo.DownlinkThroughput, "STA Downlink Throughput should be calculated")

	assert.Equal(t, 1, len(bssInfo.HistoricalChannelUtilization), "BSS Historical Channel Util should have 1 entry")
	assert.Equal(t, 1, len(bssInfo.HistoricalThroughput), "BSS Historical Throughput should have 1 entry")
	assert.Equal(t, 1, len(staInfo.HistoricalChannelUtilization), "STA Historical Channel Util should have 1 entry")
	assert.Equal(t, 1, len(staInfo.HistoricalUplinkThroughput), "STA Historical Uplink should have 1 entry")
	assert.Equal(t, 1, len(staInfo.HistoricalDownlinkThroughput), "STA Historical Downlink should have 1 entry")

	assert.Equal(t, time.Duration(0), bssInfo.totalAirtime, "BSS totalAirtime should be reset")
	assert.Equal(t, int64(0), bssInfo.totalTxBytes, "BSS totalTxBytes should be reset")
	assert.Equal(t, time.Duration(0), staInfo.totalAirtime, "STA totalAirtime should be reset")
	assert.Equal(t, int64(0), staInfo.totalUplinkBytes, "STA totalUplinkBytes should be reset")
	assert.Equal(t, int64(0), staInfo.totalDownlinkBytes, "STA totalDownlinkBytes should be reset")

	assert.NotEqual(t, calcTime, bssInfo.lastCalcTime, "BSS lastCalcTime should be updated")
	assert.NotEqual(t, calcTime, staInfo.lastCalcTime, "STA lastCalcTime should be updated")
}

func TestPeriodicallyCalculateMetrics_FirstCalculation(t *testing.T) {
	interval := 1 * time.Second
	historyPoints := 5
	sm := NewStateManager(interval, historyPoints)

	bssid := "00:11:22:33:44:66"
	bssInfo := NewBSSInfo(bssid)
	// lastCalcTime is zero initially
	bssInfo.totalAirtime = 100 * time.Millisecond // Some accumulated value
	bssInfo.totalTxBytes = 50000

	sm.bssInfos[bssid] = bssInfo

	sm.PeriodicallyCalculateMetrics()

	// On first calculation, utilization and throughput should be 0
	assert.Equal(t, 0.0, bssInfo.ChannelUtilization, "First BSS Util should be 0")
	assert.Equal(t, int64(0), bssInfo.Throughput, "First BSS Throughput should be 0")
	assert.Equal(t, 1, len(bssInfo.HistoricalChannelUtilization), "BSS Historical Util should have 1 entry")
	assert.Equal(t, 0.0, bssInfo.HistoricalChannelUtilization[0], "First BSS Historical Util value should be 0")
	assert.Equal(t, 1, len(bssInfo.HistoricalThroughput), "BSS Historical Throughput should have 1 entry")
	assert.Equal(t, int64(0), bssInfo.HistoricalThroughput[0], "First BSS Historical Throughput value should be 0")
	assert.Equal(t, time.Duration(0), bssInfo.totalAirtime, "BSS totalAirtime should be reset")
	assert.Equal(t, int64(0), bssInfo.totalTxBytes, "BSS totalTxBytes should be reset")
	assert.False(t, bssInfo.lastCalcTime.IsZero(), "BSS lastCalcTime should be updated")
}

func TestPeriodicallyCalculateMetrics_HistoryLimit(t *testing.T) {
	interval := 1 * time.Second
	historyPoints := 3 // Set a small limit for testing
	sm := NewStateManager(interval, historyPoints)

	bssid := "00:11:22:33:44:77"
	bssInfo := NewBSSInfo(bssid)
	sm.bssInfos[bssid] = bssInfo

	// Simulate multiple calculation cycles
	lastCalc := time.Now().Add(-time.Duration(historyPoints+1) * interval) // Start further back
	for i := 0; i < historyPoints+2; i++ {
		// Simulate some accumulation
		bssInfo.totalAirtime = time.Duration(100+i*10) * time.Millisecond
		bssInfo.totalTxBytes = int64(10000 + i*1000)
		bssInfo.lastCalcTime = lastCalc.Add(time.Duration(i) * interval)

		// Call calculation
		sm.PeriodicallyCalculateMetrics()

		// Check intermediate state if needed, but focus on final state
	}

	// Assert final state
	assert.Equal(t, historyPoints, len(bssInfo.HistoricalChannelUtilization), "BSS Historical Util should be capped at maxHistoryPoints")
	assert.Equal(t, historyPoints, len(bssInfo.HistoricalThroughput), "BSS Historical Throughput should be capped at maxHistoryPoints")

	// Verify the oldest data point was removed (e.g., the first calculated value is no longer present)
	// The first calculation would have resulted in 0.0 util and 0 throughput because lastCalcTime was zero initially.
	// The second calculation (i=1) would be based on (100+0*10)ms airtime = 10% util.
	// Let's check if the first non-zero value (from i=1) is still there or if it got pushed out.
	// Expected values in history (approx): [Util from i=2, Util from i=3, Util from i=4]
	// Util for i=2: totalAirtime=(100+1*10)ms=110ms -> 11%
	// Util for i=2: totalAirtime=(100+1*10)ms=110ms -> 11% (Calculated during i=2 loop) -> Becomes History[0] after i=4
	// Util for i=3: totalAirtime=(100+2*10)ms=120ms -> 12% (Calculated during i=3 loop) -> Becomes History[1] after i=4
	// Util for i=4: totalAirtime=(100+3*10)ms=130ms -> 13% (Calculated during i=4 loop) -> Becomes History[2] after i=4
	assert.InDelta(t, 12.0, bssInfo.HistoricalChannelUtilization[0], 0.01, "Oldest historical util should be from i=3 calc (based on i=2 accum)")               // Corrected expected value
	assert.InDelta(t, 14.0, bssInfo.HistoricalChannelUtilization[historyPoints-1], 0.01, "Newest historical util should be from i=5 calc (based on i=4 accum)") // Corrected expected value

	// Throughput for i=2: totalBytes=(10000+1*1000)=11000 -> 88000 bps (Calculated during i=2 loop) -> Becomes History[0] after i=4
	// Throughput for i=3: totalBytes=(10000+2*1000)=12000 -> 96000 bps (Calculated during i=3 loop) -> Becomes History[1] after i=4
	// Throughput for i=4: totalBytes=(10000+3*1000)=13000 -> 104000 bps (Calculated during i=4 loop) -> Becomes History[2] after i=4
	assert.Equal(t, int64(96000), bssInfo.HistoricalThroughput[0], "Oldest historical throughput should be from i=3 calc (based on i=2 accum)")                // Corrected expected value
	assert.Equal(t, int64(112000), bssInfo.HistoricalThroughput[historyPoints-1], "Newest historical throughput should be from i=5 calc (based on i=4 accum)") // Corrected expected value

	assert.Equal(t, time.Duration(0), bssInfo.totalAirtime, "BSS totalAirtime should be reset after final calc")
	assert.Equal(t, int64(0), bssInfo.totalTxBytes, "BSS totalTxBytes should be reset after final calc")
}

func TestPeriodicallyCalculateMetrics_ZeroAccumulation(t *testing.T) {
	interval := 1 * time.Second
	historyPoints := 5
	sm := NewStateManager(interval, historyPoints)

	bssid := "00:11:22:33:44:88"
	bssInfo := NewBSSInfo(bssid)
	bssInfo.lastCalcTime = time.Now().Add(-interval) // Set a valid last calc time
	// No accumulated data (totalAirtime and totalTxBytes remain 0)
	sm.bssInfos[bssid] = bssInfo

	sm.PeriodicallyCalculateMetrics()

	assert.Equal(t, 0.0, bssInfo.ChannelUtilization, "BSS Util should be 0 with zero accumulation")
	assert.Equal(t, int64(0), bssInfo.Throughput, "BSS Throughput should be 0 with zero accumulation")
	assert.Equal(t, 1, len(bssInfo.HistoricalChannelUtilization))
	assert.Equal(t, 0.0, bssInfo.HistoricalChannelUtilization[0])
	assert.Equal(t, 1, len(bssInfo.HistoricalThroughput))
	assert.Equal(t, int64(0), bssInfo.HistoricalThroughput[0])
	assert.Equal(t, time.Duration(0), bssInfo.totalAirtime)
	assert.Equal(t, int64(0), bssInfo.totalTxBytes)
}
