package frame_parser

import (
	"WifiPcapAnalyzer/logger"
	"WifiPcapAnalyzer/utils"
	"bufio"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	// "github.com/google/gopacket/layers" // No longer directly used for parsing
)

// HTCapabilityInfo stores parsed HT capabilities.
type HTCapabilityInfo struct {
	ChannelWidth40MHz      bool   `json:"channel_width_40mhz"`
	ShortGI20MHz           bool   `json:"short_gi_20mhz"`
	ShortGI40MHz           bool   `json:"short_gi_40mhz"`
	SupportedMCSSet        []byte `json:"supported_mcs_set"` // Raw 16 bytes
	PrimaryChannel         uint8  `json:"primary_channel"`
	SecondaryChannelOffset string `json:"secondary_channel_offset"`
	// Additional fields from tshark
	LDPCCoding         bool   `json:"ldpc_coding"`
	FortyMhzIntolerant bool   `json:"40mhz_intolerant"`
	TxSTBC             bool   `json:"tx_stbc"`
	RxSTBC             uint8  `json:"rx_stbc"`
	MaxAMSDULength     uint16 `json:"max_amsdu_length"`
	DSSCck             bool   `json:"dsss_cck_mode_40mhz"`
	HTDelayedBlockAck  bool   `json:"delayed_block_ack"`
	MaxAMPDULength     uint32 `json:"max_ampdu_length"`
}

// VHTCapabilityInfo stores parsed VHT capabilities.
type VHTCapabilityInfo struct {
	MaxMPDULength            uint8  `json:"max_mpdu_length"`
	SupportedChannelWidthSet uint8  `json:"supported_channel_width_set"`
	ShortGI80MHz             bool   `json:"short_gi_80mhz"`
	ShortGI160MHz            bool   `json:"short_gi_160mhz"`
	SUBeamformerCapable      bool   `json:"su_beamformer_capable"`
	MUBeamformerCapable      bool   `json:"mu_beamformer_capable"`
	RxMCSMap                 uint16 `json:"rx_mcs_map"`
	RxHighestLongGIRate      uint16 `json:"rx_highest_long_gi_rate"`
	TxMCSMap                 uint16 `json:"tx_mcs_map"`
	TxHighestLongGIRate      uint16 `json:"tx_highest_long_gi_rate"`
	ChannelWidth             string `json:"channel_width"` // e.g., "20", "40", "80", "160", "80+80"
	ChannelCenter0           uint8  `json:"channel_center_0"`
	ChannelCenter1           uint8  `json:"channel_center_1"` // Added: channel_center_1
	// Additional fields from tshark
	RxLDPC               bool  `json:"rx_ldpc"`
	TxSTBC               bool  `json:"tx_stbc"`
	RxSTBC               uint8 `json:"rx_stbc"`
	SUBeamformee         bool  `json:"su_beamformee"`
	MUBeamformee         bool  `json:"mu_beamformee"`
	BeamformeeSTS        uint8 `json:"beamformee_sts"`
	SoundingDimensions   uint8 `json:"sounding_dimensions"`
	MaxAMPDULengthExp    uint8 `json:"max_ampdu_length_exp"`
	RxPatternConsistency bool  `json:"rx_pattern_consistency"`
	TxPatternConsistency bool  `json:"tx_pattern_consistency"`
	VHTHTCCapability     bool  `json:"vht_htc_capability"`     // 修正：添加vht_htc_capability字段
	VHTTXOPPSCapability  bool  `json:"vht_txop_ps_capability"` // 添加：txop省电能力
}

// HECapabilityInfo stores parsed HE capabilities.
type HECapabilityInfo struct {
	// 只保留JSON示例中确认存在的字段
	BSSColor string `json:"bss_color"`
	// MAC能力字段
	HTCHESupport        bool `json:"htc_he_support"`
	TwtRequesterSupport bool `json:"twt_requester_support"`
	TwtResponderSupport bool `json:"twt_responder_support"`

	// PHY能力字段
	SUBeamformer bool `json:"su_beamformer"`
	SUBeamformee bool `json:"su_beamformee"`

	// 支持的通道宽度能力 - 从示例JSON确认存在的字段
	ChannelWidth160MHz       bool `json:"channel_width_160mhz"`         // 对应 wlan.ext_tag.he_phy_cap.chan_width_set.160_in_5ghz
	ChannelWidth80Plus80MHz  bool `json:"channel_width_80plus80mhz"`    // 对应 wlan.ext_tag.he_phy_cap.chan_width_set.160_80_80_in_5ghz
	ChannelWidth40_80MHzIn5G bool `json:"channel_width_40_80mhz_in_5g"` // 对应 wlan.ext_tag.he_phy_cap.chan_width_set.40_80_in_5ghz

	// MCS相关字段
	MaxMCSForOneSS   uint8  `json:"max_mcs_for_1_ss"`
	MaxMCSForTwoSS   uint8  `json:"max_mcs_for_2_ss"`
	MaxMCSForThreeSS uint8  `json:"max_mcs_for_3_ss"`
	MaxMCSForFourSS  uint8  `json:"max_mcs_for_4_ss"`
	RxHEMCSMap       uint16 `json:"rx_he_mcs_map"`
	TxHEMCSMap       uint16 `json:"tx_he_mcs_map"`
}

// ParsedFrameInfo holds extracted information from a single 802.11 frame.
// Fields are now populated from tshark output.
type ParsedFrameInfo struct {
	Timestamp              time.Time
	FrameType              string // e.g., "Beacon", "ProbeResp", "Data", "QoSData" (derived from wlan.fc.type_subtype)
	WlanFcType             uint8  // WLAN Frame Type (integer)
	WlanFcSubtype          uint8  // WLAN Frame Subtype (integer)
	BSSID                  net.HardwareAddr
	SA                     net.HardwareAddr
	DA                     net.HardwareAddr
	RA                     net.HardwareAddr
	TA                     net.HardwareAddr
	Channel                int      // Derived from radiotap.channel.freq or wlan.ds.current_channel
	Frequency              int      // radiotap.channel.freq
	SignalStrength         int      // radiotap.dbm_antsignal
	NoiseLevel             int      // radiotap.dbm_antnoise
	Bandwidth              string   // Derived from HT/VHT/HE capabilities
	SSID                   string   // wlan.ssid
	SupportedRates         []string // From relevant IEs, if parsed
	DSSetChannel           uint8    // wlan.ds.current_channel
	TIM                    []byte   // wlan.tim (raw bytes or parsed structure)
	RSNRaw                 []byte   // wlan.rsn.* (raw bytes or parsed structure)
	Security               string   // Security information
	IsQoSData              bool     // Derived from frame type/subtype
	ParsedHTCaps           *HTCapabilityInfo
	ParsedVHTCaps          *VHTCapabilityInfo
	ParsedHECaps           *HECapabilityInfo // New
	FrameLength            int               // frame.len (original frame length)
	FrameCapLength         int               // frame.cap_len (captured frame length)
	PHYRateMbps            float64           // Estimated PHY rate in Mbps
	IsShortPreamble        bool              // Potentially from radiotap flags (if available) or inferred
	IsShortGI              bool              // From Radiotap MCS/HT/VHT/HE flags or capabilities
	TransportPayloadLength int               // L4+ payload length (ip.len, ipv6.plen, tcp.len, udp.length)
	MACDurationID          uint16            // wlan.duration
	RetryFlag              bool              // wlan.flags.retry
	// Fields from radiotap.mcs.*, radiotap.vht.*, radiotap.he.* for PhyRateCalculator
	RadiotapDataRate   float64 // radiotap.datarate (legacy)
	RadiotapMCSIndex   uint8   // radiotap.mcs.index
	RadiotapMCSBw      uint8   // radiotap.mcs.bw (20, 40)
	RadiotapMCSGI      bool    // radiotap.mcs.gi (short GI)
	RadiotapVHTMCS     uint8   // radiotap.vht.mcs
	RadiotapVHTNSS     uint8   // radiotap.vht.nss
	RadiotapVHTBw      string  // radiotap.vht.bw (e.g., "20", "40", "80", "160", "80+80")
	RadiotapVHTShortGI bool    // radiotap.vht.gi
	RadiotapHEMCS      uint8   // radiotap.he.mcs
	RadiotapHENSS      uint8   // radiotap.he.nss
	RadiotapHEBw       string  // radiotap.he.bw (e.g., "20MHz", "40MHz", "80MHz", "HE_MU_80MHz")
	RadiotapHEGI       string  // radiotap.he.gi (e.g., "0.8us", "1.6us", "3.2us")
	BitRate            float64 // STA BitRate

	// Raw tshark fields for debugging or further processing if needed
	RawFields map[string]string
}

// PacketInfoHandler is a function that processes parsed frame information.
type PacketInfoHandler func(info *ParsedFrameInfo)

// TSharkExecutor manages the tshark process.
type TSharkExecutor struct {
	cmd *exec.Cmd
}

// Start launches the tshark process.
func (tse *TSharkExecutor) Start(pcapFilePath string, tsharkPath string, fields []string) (io.ReadCloser, io.ReadCloser, error) {
	if tsharkPath == "" {
		tsharkPath = "tshark" // Default to tshark in PATH
	}
	args := []string{
		"-r", pcapFilePath,
		"-T", "fields",
		"-E", "header=y",
		"-E", "separator=,",
		"-E", "quote=d",
		"-E", "occurrence=a", // Get all occurrences for multi-value fields
	}
	for _, field := range fields {
		args = append(args, "-e", field)
	}

	logger.Log.Info().Msgf("INFO_TSHARK_EXEC: Starting tshark with command: %s %s", tsharkPath, strings.Join(args, " "))
	tse.cmd = exec.Command(tsharkPath, args...)

	stdout, err := tse.cmd.StdoutPipe()
	if err != nil {
		logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_EXEC: Failed to get stdout pipe")
		return nil, nil, err
	}
	stderr, err := tse.cmd.StderrPipe()
	if err != nil {
		logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_EXEC: Failed to get stderr pipe")
		return nil, nil, err
	}

	if err := tse.cmd.Start(); err != nil {
		logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_EXEC: Failed to start tshark")
		return nil, nil, err
	}
	logger.Log.Info().Msgf("INFO_TSHARK_EXEC: tshark process started (PID: %d)", tse.cmd.Process.Pid)
	return stdout, stderr, nil
}

// StartStream launches the tshark process with input from an io.Reader.
func (tse *TSharkExecutor) StartStream(pcapStream io.Reader, tsharkPath string, fields []string) (io.ReadCloser, io.ReadCloser, error) {
	if tsharkPath == "" {
		tsharkPath = "tshark" // Default to tshark in PATH
	}
	args := []string{
		"-r", "-", // Read from stdin
		"-T", "fields",
		"-E", "header=y",
		"-E", "separator=,",
		"-E", "quote=d",
		"-E", "occurrence=a",
	}
	for _, field := range fields {
		args = append(args, "-e", field)
	}

	logger.Log.Info().Msgf("INFO_TSHARK_EXEC: Starting tshark with command (streaming): %s %s", tsharkPath, strings.Join(args, " "))
	tse.cmd = exec.Command(tsharkPath, args...)
	tse.cmd.Stdin = pcapStream // Set stdin to the provided stream

	stdout, err := tse.cmd.StdoutPipe()
	if err != nil {
		logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_EXEC: Failed to get stdout pipe (streaming)")
		return nil, nil, err
	}
	stderr, err := tse.cmd.StderrPipe()
	if err != nil {
		logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_EXEC: Failed to get stderr pipe (streaming)")
		return nil, nil, err
	}

	if err := tse.cmd.Start(); err != nil {
		logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_EXEC: Failed to start tshark (streaming)")
		return nil, nil, err
	}
	logger.Log.Info().Msgf("INFO_TSHARK_EXEC: tshark process started (streaming) (PID: %d)", tse.cmd.Process.Pid)
	return stdout, stderr, nil
}

// Stop terminates the tshark process.
func (tse *TSharkExecutor) Stop() {
	if tse.cmd != nil && tse.cmd.Process != nil {
		logger.Log.Info().Msgf("INFO_TSHARK_EXEC: Stopping tshark process (PID: %d)", tse.cmd.Process.Pid)
		if err := tse.cmd.Process.Kill(); err != nil {
			logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_EXEC: Failed to kill tshark process")
		}
		tse.cmd.Wait() // Wait for the command to exit and release resources
		logger.Log.Info().Msgf("INFO_TSHARK_EXEC: tshark process stopped.")
	}
}

// CSVParser parses CSV data from tshark.
type CSVParser struct {
	reader     *csv.Reader
	HeaderMap  map[string]int
	HeaderList []string
}

// NewCSVParser creates a new CSV parser.
func NewCSVParser(r io.Reader) (*CSVParser, error) {
	csvReader := csv.NewReader(r)
	header, err := csvReader.Read()
	if err != nil {
		logger.Log.Error().Err(err).Msgf("ERROR_CSV_PARSE: Failed to read CSV header")
		return nil, err
	}

	headerMap := make(map[string]int)
	for i, colName := range header {
		headerMap[colName] = i
	}
	logger.Log.Info().Msgf("INFO_CSV_PARSE: CSV Header parsed: %v", header)
	return &CSVParser{reader: csvReader, HeaderMap: headerMap, HeaderList: header}, nil
}

// ReadFrame reads a single CSV row (frame).
func (p *CSVParser) ReadFrame() (map[string]string, error) {
	record, err := p.reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		logger.Log.Error().Err(err).Msgf("ERROR_CSV_PARSE: Failed to read CSV row")
		return nil, err
	}

	frameData := make(map[string]string)
	for fieldName, index := range p.HeaderMap {
		if index < len(record) {
			frameData[fieldName] = record[index]
		} else {
			frameData[fieldName] = "" // Field not present in this row
		}
	}
	return frameData, nil
}

// FrameProcessor converts CSV rows to ParsedFrameInfo.
type FrameProcessor struct {
	headerMap map[string]int
}

// NewFrameProcessor creates a new frame processor.
func NewFrameProcessor(headerMap map[string]int) *FrameProcessor {
	return &FrameProcessor{headerMap: headerMap}
}

// Helper functions for safe field extraction and conversion
func getString(row map[string]string, fieldName string) string {
	return row[fieldName]
}

func getInt(row map[string]string, fieldName string) (int, error) {
	valStr := strings.TrimSpace(row[fieldName])
	if valStr == "" {
		return 0, fmt.Errorf("field %s is empty", fieldName)
	}
	// Handle multiple values if present (e.g., from -E occurrence=a)
	// For simplicity, take the first one if multiple are comma-separated.
	// A more robust solution might involve specific logic per field.
	if strings.Contains(valStr, ",") {
		valStr = strings.Split(valStr, ",")[0]
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse int for field %s, value '%s': %w", fieldName, valStr, err)
	}
	return val, nil
}

func getUint8(row map[string]string, fieldName string) (uint8, error) {
	valStr := strings.TrimSpace(row[fieldName])
	if valStr == "" {
		return 0, fmt.Errorf("field %s is empty", fieldName)
	}
	if strings.Contains(valStr, ",") {
		valStr = strings.Split(valStr, ",")[0]
	}
	val, err := strconv.ParseUint(valStr, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint8 for field %s, value '%s': %w", fieldName, valStr, err)
	}
	return uint8(val), nil
}

func getUint16(row map[string]string, fieldName string) (uint16, error) {
	valStr := strings.TrimSpace(row[fieldName])
	if valStr == "" {
		return 0, fmt.Errorf("field %s is empty", fieldName)
	}
	if strings.Contains(valStr, ",") {
		valStr = strings.Split(valStr, ",")[0]
	}
	val, err := strconv.ParseUint(valStr, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint16 for field %s, value '%s': %w", fieldName, valStr, err)
	}
	return uint16(val), nil
}

func getFloat64(row map[string]string, fieldName string) (float64, error) {
	valStr := strings.TrimSpace(row[fieldName])
	if valStr == "" {
		return 0, fmt.Errorf("field %s is empty", fieldName)
	}
	if strings.Contains(valStr, ",") {
		valStr = strings.Split(valStr, ",")[0]
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float64 for field %s, value '%s': %w", fieldName, valStr, err)
	}
	return val, nil
}

func getUint32(row map[string]string, fieldName string) (uint32, error) {
	valStr := strings.TrimSpace(row[fieldName])
	if valStr == "" {
		return 0, fmt.Errorf("field %s is empty", fieldName)
	}
	if strings.Contains(valStr, ",") {
		valStr = strings.Split(valStr, ",")[0]
	}
	val, err := strconv.ParseUint(valStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint32 for field %s, value '%s': %w", fieldName, valStr, err)
	}
	return uint32(val), nil
}

func getMAC(row map[string]string, fieldName string) (net.HardwareAddr, error) {
	valStr := strings.TrimSpace(row[fieldName])
	if valStr == "" {
		return nil, fmt.Errorf("field %s is empty", fieldName)
	}
	if strings.Contains(valStr, ",") {
		valStr = strings.Split(valStr, ",")[0]
	}
	mac, err := net.ParseMAC(valStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MAC for field %s, value '%s': %w", fieldName, valStr, err)
	}
	return mac, nil
}

func getBool(row map[string]string, fieldName string) (bool, error) {
	valStr := strings.TrimSpace(row[fieldName])
	if valStr == "" {
		return false, fmt.Errorf("field %s is empty", fieldName)
	}
	if strings.Contains(valStr, ",") {
		valStr = strings.Split(valStr, ",")[0]
	}
	switch strings.ToLower(valStr) {
	case "1", "true", "yes":
		return true, nil
	case "0", "false", "no":
		return false, nil
	default:
		return false, fmt.Errorf("failed to parse bool for field %s, value '%s'", fieldName, valStr)
	}
}

func parseFrameTypeSubtype(hexVal string) (uint8, uint8, string, error) {
	if hexVal == "" {
		return 0, 0, "Unknown", fmt.Errorf("wlan.fc.type_subtype is empty")
	}
	// Example: "0x08" for Beacon
	val, err := strconv.ParseUint(strings.TrimPrefix(hexVal, "0x"), 16, 8)
	if err != nil {
		return 0, 0, "Unknown", fmt.Errorf("failed to parse wlan.fc.type_subtype '%s': %w", hexVal, err)
	}
	// Correctly extract Type and Subtype from the combined field
	// Type is bits 2-3, Subtype is bits 4-7 of the 8-bit value
	// Example: Beacon is Type 0, Subtype 8. wlan.fc.type_subtype = 0x08 (0000 1000)
	// Type = (val & 0b00001100) >> 2
	// Subtype = (val & 0b11110000) >> 4
	// This logic was incorrect.
	// Instead, we should use the direct tshark fields "wlan.fc.type" and "wlan.fc.subtype"
	// which are already requested. The parseFrameTypeSubtype function will be simplified.

	// This function will now take typeVal and subtypeVal directly if they are parsed from separate fields.
	// For now, let's assume this function is called with pre-parsed type and subtype.
	// The calling code in ProcessRow needs to be updated.
	// This function's signature and purpose will change.

	// The original logic based on combined field was:
	// typeVal := (uint8(val) & 0b00001100) >> 2
	// subtypeVal := (uint8(val) & 0b11110000) >> 4
	// This is being replaced by direct field usage.
	// The parameters typeVal and subtypeVal will be passed directly.
	// So, the function signature should be:
	// func formatFrameTypeString(typeVal uint8, subtypeVal uint8) string { ... }
	// And the call site will parse "wlan.fc.type" and "wlan.fc.subtype"

	// For now, let's keep the existing structure but acknowledge the parsing error source.
	// The fix will be to use direct fields in ProcessRow.
	// This function will be simplified or removed if direct fields are used.

	// Re-evaluating: The function is called with typeSubtypeHex.
	// It *must* parse this hex value.
	// The issue is the bitwise extraction.
	// Correct extraction from an 8-bit combined value (like 0x08 for Beacon):
	// Type (bits 2,3)
	// Subtype (bits 4,5,6,7)

	// Example: 0x08 = 0000 1000
	// Type = (0000 1000 & 0000 1100) >> 2 = (0000 0010) >> 2 = 0000 0000 = 0 (Incorrect for Mgmt)
	// Subtype = (0000 1000 & 1111 0000) >> 4 = (0000 0000) >> 4 = 0 (Incorrect for Beacon)

	// Correct interpretation of wlan.fc.type_subtype (e.g., 0x08):
	// Bits 0-1: Protocol Version (usually 0)
	// Bits 2-3: Type
	// Bits 4-7: Subtype

	// So, for 0x08 (00001000):
	// Type bits are at index 2 and 3 (from right, 0-indexed): these are '00' -> Type 0 (Management)
	// Subtype bits are at index 4,5,6,7: these are '1000' -> Subtype 8 (Beacon)

	// Corrected extraction:
	typeVal := (uint8(val) >> 2) & 0x03
	subtypeVal := (uint8(val) >> 4) & 0x0F

	var typeStr string
	switch typeVal {
	case 0: // Management
		typeStr = "Mgmt"
		switch subtypeVal {
		case 0:
			typeStr = "MgmtAssocReq"
		case 1:
			typeStr = "MgmtAssocResp"
		case 2:
			typeStr = "MgmtReassocReq"
		case 3:
			typeStr = "MgmtReassocResp"
		case 4:
			typeStr = "MgmtProbeReq"
		case 5:
			typeStr = "MgmtProbeResp"
		case 8:
			typeStr = "MgmtBeacon"
		case 9:
			typeStr = "MgmtATIM"
		case 10:
			typeStr = "MgmtDisassoc"
		case 11:
			typeStr = "MgmtAuth"
		case 12:
			typeStr = "MgmtDeauth"
		case 13:
			typeStr = "MgmtAction"
		default:
			typeStr = fmt.Sprintf("MgmtSubType%d", subtypeVal)
		}
	case 1: // Control
		typeStr = "Ctrl"
		switch subtypeVal {
		case 8:
			typeStr = "CtrlBlockAckReq"
		case 9:
			typeStr = "CtrlBlockAck"
		case 10:
			typeStr = "CtrlPSPoll"
		case 11:
			typeStr = "CtrlRTS"
		case 12:
			typeStr = "CtrlCTS"
		case 13:
			typeStr = "CtrlAck"
		case 14:
			typeStr = "CtrlCFEnd"
		case 15:
			typeStr = "CtrlCFEndAck"
		default:
			typeStr = fmt.Sprintf("CtrlSubType%d", subtypeVal)
		}
	case 2: // Data
		typeStr = "Data"
		switch subtypeVal {
		case 0:
			typeStr = "Data"
		case 4:
			typeStr = "DataNull" // Null data
		case 8:
			typeStr = "QoSData"
		case 12:
			typeStr = "QoSNull" // QoS Null
		default:
			typeStr = fmt.Sprintf("DataSubType%d", subtypeVal)
		}
	default:
		typeStr = fmt.Sprintf("Type%dSubType%d", typeVal, subtypeVal)
	}
	return typeVal, subtypeVal, typeStr, nil
}

// ProcessRow converts a CSV row to ParsedFrameInfo.
func (fp *FrameProcessor) ProcessRow(row map[string]string) (*ParsedFrameInfo, error) {
	info := &ParsedFrameInfo{RawFields: row}
	// var err error
	var parseErrors []string

	// 添加INFO级别日志，记录每一行处理开始
	logger.Log.Info().Msgf("INFO_PROCESS_ROW: Processing new frame row, checking for security fields")

	// 记录关键安全字段的存在情况
	privacyBit := getString(row, "wlan.fixed.capabilities.privacy")
	rsnAkms := getString(row, "wlan.rsn.akms.type")
	frameType := getString(row, "wlan.fc.type")
	frameSubType := getString(row, "wlan.fc.subtype")

	logger.Log.Info().Msgf("INFO_SECURITY_FIELDS: privacy_bit='%s', rsn_akms='%s', frame_type='%s', frame_subtype='%s'",
		privacyBit, rsnAkms, frameType, frameSubType)

	// 5.1. Frame basic information
	fieldName := "frame.time_epoch"
	rawValue := getString(row, fieldName)
	if val, e := getFloat64(row, fieldName); e == nil {
		sec, dec := int64(val), int64((val-float64(int64(val)))*1e9)
		info.Timestamp = time.Unix(sec, dec)
	} else {
		errStr := fmt.Sprintf("frame.time_epoch: %v", e)
		parseErrors = append(parseErrors, errStr)
		// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse field '%s' with value '%s': %v", fieldName, rawValue, e)
		info.Timestamp = time.Time{} // Fallback to zero time, log error
	}

	fieldName = "frame.len"
	rawValue = getString(row, fieldName)
	if val, e := getInt(row, fieldName); e == nil {
		info.FrameLength = val
	} else {
		errStr := fmt.Sprintf("frame.len: %v", e)
		parseErrors = append(parseErrors, errStr)
		// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse field '%s' with value '%s': %v", fieldName, rawValue, e)
	}

	fieldName = "frame.cap_len"
	rawValue = getString(row, fieldName)
	if val, e := getInt(row, fieldName); e == nil {
		info.FrameCapLength = val
	} else {
		// Optional, so don't add to parseErrors if missing
		if rawValue != "" { // Log error only if field was present but unparsable
			// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
		}
	}

	fieldName = "wlan.fc.type_subtype"
	typeSubtypeHex := getString(row, fieldName)
	if typeVal, subtypeVal, typeStr, e := parseFrameTypeSubtype(typeSubtypeHex); e == nil {
		info.WlanFcType = typeVal // This will now be correctly 0 for Mgmt, 1 for Ctrl, 2 for Data
		info.WlanFcSubtype = subtypeVal
		info.FrameType = typeStr
		if info.WlanFcType == 2 && (info.WlanFcSubtype == 8 || info.WlanFcSubtype == 12) { // QoSData or QoSNull
			info.IsQoSData = true
		}
	} else {
		errStr := fmt.Sprintf("wlan.fc.type_subtype: %v", e)
		parseErrors = append(parseErrors, errStr)
		// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse field '%s' with value '%s': %v", fieldName, typeSubtypeHex, e)
	}

	// Alternative: Use direct wlan.fc.type and wlan.fc.subtype if available
	// This is preferred if tshark provides them reliably.
	// Check if "wlan.fc.type" and "wlan.fc.subtype" are in defaultTsharkFields
	// They are: "wlan.fc.type", "wlan.fc.subtype"
	// So, let's prioritize these direct fields.

	typeValStr := getString(row, "wlan.fc.type")
	subtypeValStr := getString(row, "wlan.fc.subtype")

	if typeInt, typeErr := strconv.Atoi(typeValStr); typeErr == nil {
		if subtypeInt, subtypeErr := strconv.Atoi(subtypeValStr); subtypeErr == nil {
			info.WlanFcType = uint8(typeInt)
			info.WlanFcSubtype = uint8(subtypeInt)
			// Regenerate FrameType string based on these direct values
			_, _, info.FrameType, _ = parseFrameTypeSubtype(typeSubtypeHex) // Keep using combined for string for now, or adapt formatFrameTypeString
			// Better: adapt parseFrameTypeSubtype to take numeric type/subtype for string formatting
			// For now, the string might be based on the potentially miscalculated combined field if direct parsing fails later.
			// Let's refine FrameType string generation after confirming direct type/subtype
			// This is a bit redundant if parseFrameTypeSubtype is correct.
			// The main goal is to get WlanFcType correct for state_manager.

			// Re-generate FrameType string using the directly parsed type and subtype
			var tempTypeStr string
			switch info.WlanFcType {
			case 0: // Management
				tempTypeStr = "Mgmt"
				switch info.WlanFcSubtype {
				case 0:
					tempTypeStr = "MgmtAssocReq"
				case 1:
					tempTypeStr = "MgmtAssocResp"
				case 2:
					tempTypeStr = "MgmtReassocReq"
				case 3:
					tempTypeStr = "MgmtReassocResp"
				case 4:
					tempTypeStr = "MgmtProbeReq"
				case 5:
					tempTypeStr = "MgmtProbeResp"
				case 8:
					tempTypeStr = "MgmtBeacon"
				case 9:
					tempTypeStr = "MgmtATIM"
				case 10:
					tempTypeStr = "MgmtDisassoc"
				case 11:
					tempTypeStr = "MgmtAuth"
				case 12:
					tempTypeStr = "MgmtDeauth"
				case 13:
					tempTypeStr = "MgmtAction"
				default:
					tempTypeStr = fmt.Sprintf("MgmtSubType%d", info.WlanFcSubtype)
				}
			case 1: // Control
				tempTypeStr = "Ctrl"
				switch info.WlanFcSubtype {
				case 8:
					tempTypeStr = "CtrlBlockAckReq"
				case 9:
					tempTypeStr = "CtrlBlockAck"
				case 10:
					tempTypeStr = "CtrlPSPoll"
				case 11:
					tempTypeStr = "CtrlRTS"
				case 12:
					tempTypeStr = "CtrlCTS"
				case 13:
					tempTypeStr = "CtrlAck"
				case 14:
					tempTypeStr = "CtrlCFEnd"
				case 15:
					tempTypeStr = "CtrlCFEndAck"
				default:
					tempTypeStr = fmt.Sprintf("CtrlSubType%d", info.WlanFcSubtype)
				}
			case 2: // Data
				tempTypeStr = "Data"
				switch info.WlanFcSubtype {
				case 0:
					tempTypeStr = "Data"
				case 4:
					tempTypeStr = "DataNull"
				case 8:
					tempTypeStr = "QoSData"
					info.IsQoSData = true
				case 12:
					tempTypeStr = "QoSNull"
					info.IsQoSData = true
				default:
					tempTypeStr = fmt.Sprintf("DataSubType%d", info.WlanFcSubtype)
				}
			default:
				tempTypeStr = fmt.Sprintf("Type%dSubType%d", info.WlanFcType, info.WlanFcSubtype)
			}
			info.FrameType = tempTypeStr
			// logger.Log.Debug().Msgf("DEBUG_PARSE_DIRECT_TYPE_SUBTYPE: Successfully parsed direct wlan.fc.type=%d, wlan.fc.subtype=%d. FrameType set to: %s", info.WlanFcType, info.WlanFcSubtype, info.FrameType)

		} else if subtypeValStr != "" {
			// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse direct wlan.fc.subtype '%s': %v", subtypeValStr, subtypeErr)
		}
	} else if typeValStr != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse direct wlan.fc.type '%s': %v", typeValStr, typeErr)
	}

	fieldName = "wlan.fc.retry"
	rawValue = getString(row, fieldName)
	if val, e := getBool(row, fieldName); e == nil {
		info.RetryFlag = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	// 5.2. MAC Addresses
	fieldsToParse := []string{"wlan.ra", "wlan.da", "wlan.ta", "wlan.sa", "wlan.bssid"}
	macDestinations := []*net.HardwareAddr{&info.RA, &info.DA, &info.TA, &info.SA, &info.BSSID}

	for i, fieldName := range fieldsToParse {
		rawValue := getString(row, fieldName)
		if mac, e := getMAC(row, fieldName); e == nil {
			*macDestinations[i] = mac
		} else if rawValue != "" {
			errStr := fmt.Sprintf("%s: %v", fieldName, e)
			parseErrors = append(parseErrors, errStr)
			// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse field '%s' with value '%s': %v", fieldName, rawValue, e)
		}
	}

	// 5.3. Radiotap and Physical Layer Information
	fieldName = "radiotap.channel.freq"
	rawValue = getString(row, fieldName)
	if val, e := getInt(row, fieldName); e == nil {
		info.Frequency = val
		info.Channel = utils.FrequencyToChannel(val)
	} else if rawValue != "" {
		// Don't add to parseErrors, let it be default. Critical for BSSID/SA/DA etc.
		// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse field '%s' with value '%s' (will use default): %v", fieldName, rawValue, e)
	}

	// 检查并记录RSN相关字段
	rsnVersion := getString(row, "wlan.rsn.version")
	privacy := getString(row, "wlan.fixed.capabilities.privacy")
	akmsType := getString(row, "wlan.rsn.akms.type")

	// 如果有任何安全相关字段，设置RSNRaw为非空
	if rsnVersion != "" || (privacy == "True" && akmsType != "") {
		info.RSNRaw = []byte{1} // 至少有一个非零元素表示存在RSN
		logger.Log.Info().Msgf("INFO_RSN_DATA: RSN data found - version='%s', privacy='%s', akms.type='%s'",
			rsnVersion, privacy, akmsType)
	} else {
		info.RSNRaw = []byte{} // 空slice
		logger.Log.Info().Msgf("INFO_RSN_DATA: No RSN data found in this frame")
	}

	fieldName = "radiotap.dbm_antsignal"
	rawValue = getString(row, fieldName)
	if val, e := getInt(row, fieldName); e == nil {
		info.SignalStrength = val
	} else if rawValue != "" {
		// Don't add to parseErrors, let it be default.
		// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse field '%s' with value '%s' (will use default): %v", fieldName, rawValue, e)
	}

	fieldName = "radiotap.dbm_antnoise"
	rawValue = getString(row, fieldName)
	if val, e := getInt(row, fieldName); e == nil {
		info.NoiseLevel = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	// Radiotap PHY fields for PhyRateCalculator
	fieldName = "radiotap.datarate"
	rawValue = getString(row, fieldName)
	if val, e := getFloat64(row, fieldName); e == nil {
		info.RadiotapDataRate = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	fieldName = "radiotap.mcs.index"
	rawValue = getString(row, fieldName)
	if val, e := getUint8(row, fieldName); e == nil {
		info.RadiotapMCSIndex = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	fieldName = "radiotap.mcs.bw"
	rawValue = getString(row, fieldName)
	if val, e := getUint8(row, fieldName); e == nil {
		info.RadiotapMCSBw = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	fieldName = "radiotap.mcs.gi"
	rawValue = getString(row, fieldName)
	if val, e := getBool(row, fieldName); e == nil {
		info.RadiotapMCSGI = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	fieldName = "radiotap.vht.mcs"
	rawValue = getString(row, fieldName)
	if val, e := getUint8(row, fieldName); e == nil {
		info.RadiotapVHTMCS = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	fieldName = "radiotap.vht.nss"
	rawValue = getString(row, fieldName)
	if val, e := getUint8(row, fieldName); e == nil {
		info.RadiotapVHTNSS = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	info.RadiotapVHTBw = getString(row, "radiotap.vht.bw") // String: "20", "40", "80", "160", "80+80"

	fieldName = "radiotap.vht.gi"
	rawValue = getString(row, fieldName)
	if val, e := getBool(row, fieldName); e == nil {
		info.RadiotapVHTShortGI = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	fieldName = "radiotap.he.mcs"
	rawValue = getString(row, fieldName)
	if val, e := getUint8(row, fieldName); e == nil {
		info.RadiotapHEMCS = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	fieldName = "radiotap.he.nss"
	rawValue = getString(row, fieldName)
	if val, e := getUint8(row, fieldName); e == nil {
		info.RadiotapHENSS = val
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	// 5.5.1 BitRate
	// 首先尝试从wlan.supported_rates获取比特率
	fieldName = "wlan.supported_rates"
	rawValue = getString(row, fieldName)
	if rawValue != "" {
		// 支持的速率可能有多个值，我们取第一个非零值
		rates := strings.Split(rawValue, ",")
		for _, rate := range rates {
			if val, e := strconv.ParseFloat(strings.TrimSpace(rate), 64); e == nil && val > 0 {
				info.BitRate = val
				logger.Log.Debug().Msgf("DEBUG_BITRATE_FROM_SUPPORTED_RATES: Using wlan.supported_rates=%v to set BitRate=%.1f Mbps", val, info.BitRate)
				break
			}
		}
	}

	// 如果wlan.supported_rates未设置BitRate，尝试使用wlan_radio.data_rate
	if info.BitRate == 0 {
		fieldName = "wlan_radio.data_rate"
		rawValue = getString(row, fieldName)
		if val, e := getFloat64(row, fieldName); e == nil && val > 0 {
			info.BitRate = val
			logger.Log.Debug().Msgf("DEBUG_BITRATE_FROM_WLAN_RADIO: Using wlan_radio.data_rate=%v to set BitRate=%.1f Mbps", val, info.BitRate)
		} else if rawValue != "" {
			logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
		}
	}

	info.RadiotapHEBw = getString(row, "radiotap.he.bw") // String: "20MHz", "40MHz", "80MHz", "HE_MU_80MHz" etc.
	info.RadiotapHEGI = getString(row, "radiotap.he.gi") // String: "0.8us", "1.6us", "3.2us"

	// Calculate PHY Rate
	info.PHYRateMbps = getPHYRateMbps(info) // Pass the partially filled info

	// 5.4. BSS Information
	fieldName = "wlan.ssid"
	rawSsidStr := getString(row, fieldName)

	// Determine security
	security := "Open" // Default
	capPrivacyStr := getString(row, "wlan.fixed.capabilities.privacy")
	rsnAkmsType := getString(row, "wlan.rsn.akms.type")

	logger.Log.Info().Msgf("INFO_SECURITY_PARSE: Begin security parsing for BSSID %s", info.BSSID)
	logger.Log.Info().Msgf("INFO_SECURITY_PARSE: wlan.fixed.capabilities.privacy = '%s'", capPrivacyStr)
	logger.Log.Info().Msgf("INFO_SECURITY_PARSE: wlan.rsn.akms.type = '%s'", rsnAkmsType)

	// 增加解析BSSID相关字段
	bssidStr := getString(row, "wlan.bssid")
	if bssidStr != "" {
		logger.Log.Info().Msgf("INFO_SECURITY_PARSE: Found BSSID in data: %s", bssidStr)
	}

	// 检查管理帧相关字段
	ssidStr := getString(row, "wlan.ssid")
	if ssidStr != "" {
		logger.Log.Info().Msgf("INFO_SECURITY_PARSE: Found SSID in data: %s", ssidStr)
	}

	if capPrivacyStr == "1" {
		security = "WEP"
		logger.Log.Info().Msgf("INFO_SECURITY_PARSE: Privacy bit set, default to WEP")
	}

	if rsnAkmsType != "" {
		// 详细解析wlan.rsn.akms.type以确定具体加密类型
		if strings.Contains(rsnAkmsType, "2") {
			security = "WPA2-PSK"
			logger.Log.Info().Msgf("INFO_SECURITY_PARSE: RSN/AKM type contains '2', setting to WPA2-PSK")
		} else if strings.Contains(rsnAkmsType, "8") {
			security = "WPA3-SAE"
			logger.Log.Info().Msgf("INFO_SECURITY_PARSE: RSN/AKM type contains '8', setting to WPA3-SAE")
		} else {
			security = "WPA2/WPA3"
			logger.Log.Info().Msgf("INFO_SECURITY_PARSE: RSN/AKM type present but not recognized, using generic WPA2/WPA3")
		}
	}

	logger.Log.Info().Msgf("INFO_SECURITY_PARSE: Final security type: %s", security)
	info.Security = security

	if rawSsidStr != "" {
		// Attempt to decode from hex first
		decodedHexSsid, err := hex.DecodeString(rawSsidStr)
		processedSsidStr := ""
		if err == nil {
			processedSsidStr = string(decodedHexSsid)
			// logger.Log.Debug().Msgf("DEBUG_SSID_HEX_DECODE_SUCCESS: Field '%s' was hex-decoded to: '%s'", fieldName, processedSsidStr)
		} else {
			// Not valid hex, or some other error, use original string
			processedSsidStr = rawSsidStr
			// logger.Log.Debug().Msgf("DEBUG_SSID_HEX_DECODE_FAIL: Field '%s' not valid hex ('%s'), using as is. Error: %v", fieldName, rawSsidStr, err)
		}

		if utf8.ValidString(processedSsidStr) {
			info.SSID = processedSsidStr
		} else {
			info.SSID = "<Invalid SSID Encoding>"
		}
	}
	// logger.Log.Debug().Msgf("DEBUG_SSID_DECODED: Decoded SSID for BSSID %s: %s", info.SSID, info.SSID)
	// Add the requested log after SSID processing
	// logger.Log.Debug().Msgf("DEBUG_SSID_DECODED: Decoded SSID for BSSID %s: %s", info.BSSID, info.SSID)

	fieldName = "wlan.ds.current_channel"
	rawValue = getString(row, fieldName)
	if val, e := getUint8(row, fieldName); e == nil {
		info.DSSetChannel = val
		if info.Channel == 0 && val > 0 { // If radiotap channel was missing, use DSSet
			info.Channel = int(val)
		}
	} else if rawValue != "" {
		// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
	}

	// HT Capabilities
	if htCapsStr := getString(row, "wlan.ht.capabilities"); htCapsStr != "" {
		info.ParsedHTCaps = &HTCapabilityInfo{}
		fieldName := "wlan.ht.info.primarychannel"
		rawValue := getString(row, fieldName)
		if val, e := getUint8(row, fieldName); e == nil {
			info.ParsedHTCaps.PrimaryChannel = val
		} else if rawValue != "" {
			logger.Log.Warn().Msgf("WARN_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
		}

		info.ParsedHTCaps.SecondaryChannelOffset = getString(row, "wlan.ht.info.secchanoffset") // e.g. "above", "below"

		// Parse additional HT capability fields
		if val, e := getBool(row, "wlan.ht.capabilities.ldpccoding"); e == nil {
			info.ParsedHTCaps.LDPCCoding = val
		}

		if val, e := getBool(row, "wlan.ht.capabilities.width"); e == nil && val {
			info.ParsedHTCaps.ChannelWidth40MHz = val
		}

		if val, e := getBool(row, "wlan.ht.capabilities.short20"); e == nil {
			info.ParsedHTCaps.ShortGI20MHz = val
		}

		if val, e := getBool(row, "wlan.ht.capabilities.short40"); e == nil {
			info.ParsedHTCaps.ShortGI40MHz = val
		}

		if val, e := getBool(row, "wlan.ht.capabilities.txstbc"); e == nil {
			info.ParsedHTCaps.TxSTBC = val
		}

		if val, e := getUint8(row, "wlan.ht.capabilities.rxstbc"); e == nil {
			info.ParsedHTCaps.RxSTBC = val
		}

		if val, e := getBool(row, "wlan.ht.capabilities.delayedblockack"); e == nil {
			info.ParsedHTCaps.HTDelayedBlockAck = val
		}

		if val, e := getBool(row, "wlan.ht.capabilities.dsscck"); e == nil {
			info.ParsedHTCaps.DSSCck = val
		}

		if val, e := getBool(row, "wlan.ht.capabilities.40mhzintolerant"); e == nil {
			info.ParsedHTCaps.FortyMhzIntolerant = val
		}

		// Parse AMPDU parameters
		if val, e := getUint32(row, "wlan.ht.ampduparam.maxlength"); e == nil {
			info.ParsedHTCaps.MaxAMPDULength = val
		}

		// Prioritize radiotap.mcs.bw for bandwidth if available
		if info.RadiotapMCSBw == 1 { // 40MHz
			info.Bandwidth = "40MHz"
		} else if info.RadiotapMCSBw == 0 { // 20MHz
			info.Bandwidth = "20MHz"
		} else {
			// If radiotap.mcs.bw is not available, infer from channel offset
			if info.ParsedHTCaps.SecondaryChannelOffset != "" || info.ParsedHTCaps.ChannelWidth40MHz {
				info.Bandwidth = "40MHz" // Infer 40MHz if secondary channel offset is present
			} else {
				info.Bandwidth = "20MHz" // Default to 20MHz if no offset
			}
		}
		if info.RadiotapMCSGI {
			info.IsShortGI = true
		} else {
			// Potentially infer Short GI from HT Capabilities, if not in radiotap
			if info.ParsedHTCaps.ShortGI20MHz || info.ParsedHTCaps.ShortGI40MHz {
				info.IsShortGI = true
			}
		}
	} // HT Capabilities 结束

	// VHT Capabilities
	if vhtCapsStr := getString(row, "wlan.vht.capabilities"); vhtCapsStr != "" {
		info.ParsedVHTCaps = &VHTCapabilityInfo{}
		vhtChannelWidth := getString(row, "wlan.vht.op.channelwidth")
		info.ParsedVHTCaps.ChannelWidth = vhtChannelWidth

		// Parse SupportedChannelWidthSet
		if val, e := getUint8(row, "wlan.vht.capabilities.supportedchanwidthset"); e == nil {
			info.ParsedVHTCaps.SupportedChannelWidthSet = val
		}

		// Parse MaxMPDULength
		if val, e := getUint8(row, "wlan.vht.capabilities.maxmpdulength"); e == nil {
			info.ParsedVHTCaps.MaxMPDULength = val
		}

		// Parse short GI flags
		if val, e := getBool(row, "wlan.vht.capabilities.short80"); e == nil {
			info.ParsedVHTCaps.ShortGI80MHz = val
		}

		if val, e := getBool(row, "wlan.vht.capabilities.short160"); e == nil {
			info.ParsedVHTCaps.ShortGI160MHz = val
		}

		// Parse beamforming capabilities
		if val, e := getBool(row, "wlan.vht.capabilities.subeamformer"); e == nil {
			info.ParsedVHTCaps.SUBeamformerCapable = val
		}

		if val, e := getBool(row, "wlan.vht.capabilities.subeamformee"); e == nil {
			info.ParsedVHTCaps.SUBeamformee = val
		}

		if val, e := getBool(row, "wlan.vht.capabilities.mubeamformer"); e == nil {
			info.ParsedVHTCaps.MUBeamformerCapable = val
		}

		if val, e := getBool(row, "wlan.vht.capabilities.mubeamformee"); e == nil {
			info.ParsedVHTCaps.MUBeamformee = val
		}

		// Parse STBC capabilities
		if val, e := getBool(row, "wlan.vht.capabilities.txstbc"); e == nil {
			info.ParsedVHTCaps.TxSTBC = val
		}

		if val, e := getUint8(row, "wlan.vht.capabilities.rxstbc"); e == nil {
			info.ParsedVHTCaps.RxSTBC = val
		}

		// Parse LDPC capability
		if val, e := getBool(row, "wlan.vht.capabilities.rxldpc"); e == nil {
			info.ParsedVHTCaps.RxLDPC = val
		}

		// Parse beamforming parameters
		if val, e := getUint8(row, "wlan.vht.capabilities.beamformee_sts_cap"); e == nil {
			info.ParsedVHTCaps.BeamformeeSTS = val
		}

		if val, e := getUint8(row, "wlan.vht.capabilities.soundingdimensions"); e == nil {
			info.ParsedVHTCaps.SoundingDimensions = val
		}

		// Parse consistency flags
		if val, e := getBool(row, "wlan.vht.capabilities.rxpatconsist"); e == nil {
			info.ParsedVHTCaps.RxPatternConsistency = val
		}

		if val, e := getBool(row, "wlan.vht.capabilities.txpatconsist"); e == nil {
			info.ParsedVHTCaps.TxPatternConsistency = val
		}

		// Parse AMPDU length exponent
		if val, e := getUint8(row, "wlan.vht.capabilities.maxampdu"); e == nil {
			info.ParsedVHTCaps.MaxAMPDULengthExp = val
		}

		// Parse VHT MCS maps
		if val, e := getUint16(row, "wlan.vht.mcsset.rxmcsmap"); e == nil {
			info.ParsedVHTCaps.RxMCSMap = val
		}

		if val, e := getUint16(row, "wlan.vht.mcsset.txmcsmap"); e == nil {
			info.ParsedVHTCaps.TxMCSMap = val
		}

		if val, e := getUint16(row, "wlan.vht.mcsset.rxhighestlonggirate"); e == nil {
			info.ParsedVHTCaps.RxHighestLongGIRate = val
		}

		if val, e := getUint16(row, "wlan.vht.mcsset.txhighestlonggirate"); e == nil {
			info.ParsedVHTCaps.TxHighestLongGIRate = val
		}

		// Parse VHT HTC capability
		if val, e := getBool(row, "wlan.vht.capabilities.vhthtc"); e == nil {
			info.ParsedVHTCaps.VHTHTCCapability = val
		}

		// Parse VHT TXOP PS capability
		if val, e := getBool(row, "wlan.vht.capabilities.vhttxopps"); e == nil {
			info.ParsedVHTCaps.VHTTXOPPSCapability = val
		}

		fieldName := "wlan.vht.op.channelcenter0"
		rawValue := getString(row, fieldName)
		if val, e := getUint8(row, fieldName); e == nil {
			info.ParsedVHTCaps.ChannelCenter0 = val
		} else if rawValue != "" {
			logger.Log.Warn().Msgf("WARN_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
		}

		fieldName = "wlan.vht.op.channelcenter1"
		rawValue = getString(row, fieldName)
		if val, e := getUint8(row, fieldName); e == nil {
			info.ParsedVHTCaps.ChannelCenter1 = val
		} else if rawValue != "" {
			logger.Log.Warn().Msgf("WARN_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, e)
		}

		// Prioritize radiotap.vht.bw for bandwidth if available
		if info.RadiotapVHTBw != "" {
			info.Bandwidth = info.RadiotapVHTBw // Use radiotap.vht.bw directly
		} else {
			// Infer from VHT channel width if radiotap is missing
			switch vhtChannelWidth {
			case "1":
				info.Bandwidth = "80MHz"
			case "2":
				info.Bandwidth = "160MHz"
			case "3":
				info.Bandwidth = "80+80MHz"
			case "0":
				// Differentiate 20/40 MHz based on HT or other info if possible
				if info.Bandwidth != "40MHz" {
					info.Bandwidth = "20MHz"
				}
			}
			// Alternative inference from supportedChannelWidthSet if available
			if info.ParsedVHTCaps.SupportedChannelWidthSet >= 1 && info.Bandwidth == "" {
				switch info.ParsedVHTCaps.SupportedChannelWidthSet {
				case 1:
					info.Bandwidth = "80MHz"
				case 2, 3:
					info.Bandwidth = "160MHz"
				}
			}
		}

		if info.RadiotapVHTShortGI {
			info.IsShortGI = true
		} else {
			// Can potentially infer Short GI from VHT Capabilities
			if info.ParsedVHTCaps.ShortGI80MHz || info.ParsedVHTCaps.ShortGI160MHz {
				info.IsShortGI = true
			}
		}
	}

	// HE Capabilities (Less Simplified)
	if heMacCapsStr := getString(row, "wlan.ext_tag.he_mac_caps"); heMacCapsStr != "" {
		info.ParsedHECaps = &HECapabilityInfo{}
		info.ParsedHECaps.BSSColor = getString(row, "wlan.ext_tag.bss_color_information.bss_color")

		// Parse HE MAC capabilities
		if val, e := getBool(row, "wlan.ext_tag.he_mac_cap.htc_he_support"); e == nil {
			info.ParsedHECaps.HTCHESupport = val
		}

		if val, e := getBool(row, "wlan.ext_tag.he_mac_cap.twt_req_support"); e == nil {
			info.ParsedHECaps.TwtRequesterSupport = val
		}

		if val, e := getBool(row, "wlan.ext_tag.he_mac_cap.twt_rsp_support"); e == nil {
			info.ParsedHECaps.TwtResponderSupport = val
		}

		// Parse HE PHY capabilities
		if val, e := getBool(row, "wlan.ext_tag.he_phy_cap.su_beamformer"); e == nil {
			info.ParsedHECaps.SUBeamformer = val
		}

		if val, e := getBool(row, "wlan.ext_tag.he_phy_cap.su_beamformee"); e == nil {
			info.ParsedHECaps.SUBeamformee = val
		}

		// 解析HE通道宽度相关能力
		if val, e := getBool(row, "wlan.ext_tag.he_phy_cap.chan_width_set.160_in_5ghz"); e == nil {
			info.ParsedHECaps.ChannelWidth160MHz = val
		}

		if val, e := getBool(row, "wlan.ext_tag.he_phy_cap.chan_width_set.160_80_80_in_5ghz"); e == nil {
			info.ParsedHECaps.ChannelWidth80Plus80MHz = val
		}

		if val, e := getBool(row, "wlan.ext_tag.he_phy_cap.chan_width_set.40_80_in_5ghz"); e == nil {
			info.ParsedHECaps.ChannelWidth40_80MHzIn5G = val
		}

		// Parse HE MCS sets
		if val, e := getUint8(row, "wlan.ext_tag.he_operation.max_he_mcs_for_1_ss"); e == nil {
			info.ParsedHECaps.MaxMCSForOneSS = val
		}

		if val, e := getUint8(row, "wlan.ext_tag.he_operation.max_he_mcs_for_2_ss"); e == nil {
			info.ParsedHECaps.MaxMCSForTwoSS = val
		}

		if val, e := getUint8(row, "wlan.ext_tag.he_operation.max_he_mcs_for_3_ss"); e == nil {
			info.ParsedHECaps.MaxMCSForThreeSS = val
		}

		if val, e := getUint8(row, "wlan.ext_tag.he_operation.max_he_mcs_for_4_ss"); e == nil {
			info.ParsedHECaps.MaxMCSForFourSS = val
		}

		if val, e := getUint16(row, "wlan.ext_tag.he_mcs_map.rx_he_mcs_map_lte_80"); e == nil {
			info.ParsedHECaps.RxHEMCSMap = val
		}

		if val, e := getUint16(row, "wlan.ext_tag.he_mcs_map.tx_he_mcs_map_lte_80"); e == nil {
			info.ParsedHECaps.TxHEMCSMap = val
		}

		// 带宽判断逻辑：从HE能力推断带宽
		// 优先使用radiotap.he.bw（如果存在）
		heBw := info.RadiotapHEBw
		if strings.Contains(heBw, "160MHz") {
			info.Bandwidth = "160MHz"
		} else if strings.Contains(heBw, "80MHz") {
			info.Bandwidth = "80MHz"
		} else if strings.Contains(heBw, "40MHz") {
			info.Bandwidth = "40MHz"
		} else if strings.Contains(heBw, "20MHz") {
			info.Bandwidth = "20MHz"
		} else {
			// 如果radiotap未提供HE带宽信息，从HE能力推断
			if info.ParsedHECaps.ChannelWidth160MHz {
				info.Bandwidth = "160MHz"
			} else if info.ParsedHECaps.ChannelWidth80Plus80MHz {
				info.Bandwidth = "80+80MHz"
			} else if info.ParsedHECaps.ChannelWidth40_80MHzIn5G {
				// 如果支持5G中的40MHz和80MHz，默认使用最高的能力
				info.Bandwidth = "80MHz"
			}
		}

		// 处理HE短保护间隔
		heGI := info.RadiotapHEGI
		if heGI != "" && heGI != "0.8us" {
			info.IsShortGI = true
		} else {
			// 可以从HE MAC Caps字符串推断Short GI
			if strings.Contains(heMacCapsStr, "Short GI") {
				info.IsShortGI = true
			}
		}
	}

	// 5.6. Throughput calculation parameters
	info.TransportPayloadLength = -1 // Default if not found
	fieldName = "ip.len"
	rawValue = getString(row, fieldName)
	if val, e := getInt(row, fieldName); e == nil {
		info.TransportPayloadLength = val
		// Further checks for tcp.len or udp.length can be added here with similar logging
	} else if rawValue != "" {
		// Try ipv6.plen if ip.len failed or was not present
		fieldName = "ipv6.plen"
		rawValue = getString(row, fieldName)
		if valIPv6, eIPv6 := getInt(row, fieldName); eIPv6 == nil {
			info.TransportPayloadLength = valIPv6
		} else if rawValue != "" {
			// logger.Log.Debug().Msgf("DEBUG_PARSE_FIELD_OPTIONAL_ERROR: Optional field '%s' with value '%s' failed to parse: %v", fieldName, rawValue, eIPv6)
		}
	}

	// 5.7. Frame duration/airtime calculation parameters
	fieldName = "wlan.duration"
	rawValue = getString(row, fieldName)
	if val, e := getUint16(row, fieldName); e == nil {
		info.MACDurationID = val
	} else if rawValue != "" {
		// Don't add to parseErrors, let it be default.
		// logger.Log.Debug().Msgf("DEBUG_PARSE_ERROR: Failed to parse field '%s' with value '%s' (will use default): %v", fieldName, rawValue, e)
	}

	// 带宽优先级：HE > VHT > HT
	// 确保最终带宽使用最高标准支持的宽度
	if info.ParsedHECaps != nil {
		// HE已经设置了带宽，优先级最高
		// 在HE部分已经处理
	} else if info.ParsedVHTCaps != nil {
		// 如果没有HE带宽信息，使用VHT带宽
		// 在VHT部分已经处理
	} else if info.ParsedHTCaps != nil {
		// 如果既没有HE也没有VHT带宽信息，使用HT带宽
		// 在HT部分已经处理
	}

	// 确保带宽信息始终有值
	if info.Bandwidth == "" {
		info.Bandwidth = "20MHz" // 默认值
	}

	if len(parseErrors) > 0 {
		// Combined error log for all parsing issues in this row
		// logger.Log.Debug().Msgf("DEBUG_PARSE_ROW_ERRORS: Errors parsing frame row: %s. Raw row: %v", strings.Join(parseErrors, "; "), row)
		return info, fmt.Errorf("errors parsing frame row: %s", strings.Join(parseErrors, "; "))
	}

	// logger.Log.Debug().Msgf("DEBUG_PARSED_FRAME: Successfully parsed frame: Timestamp=%s, SA=%s, DA=%s, BSSID=%s, SSID='%s', Signal=%d, FrameType=%s, Channel=%d, BW=%s",
	// 	info.Timestamp.Format(time.RFC3339Nano),
	// 	info.SA, info.DA, info.BSSID, info.SSID, info.SignalStrength, info.FrameType, info.Channel, info.Bandwidth)
	return info, nil
}

// ProcessPcapFile is the main entry point for parsing a pcap file using tshark.
func ProcessPcapFile(pcapFilePath string, tsharkPath string, pktHandler PacketInfoHandler) error {
	// Define all necessary tshark fields based on the specification
	fields := []string{
		"frame.time_epoch", "frame.len", "frame.cap_len", "wlan.fc.type_subtype",
		"wlan.fc.type", "wlan.fc.subtype", "wlan.fc.retry", // Corrected: wlan.flags.retry -> wlan.fc.retry
		"wlan.ra", "wlan.da", "wlan.ta", "wlan.sa", "wlan.bssid",
		"radiotap.channel.freq", "radiotap.dbm_antsignal", "radiotap.dbm_antnoise",
		"radiotap.datarate", "radiotap.mcs.index", "radiotap.mcs.bw", "radiotap.mcs.gi",
		"radiotap.vht.bw", "radiotap.vht.gi", "wlan_radio.data_rate", "wlan.supported_rates",
		"wlan.ssid", "wlan.fixed.beacon", "wlan.fixed.capabilities.ess", "wlan.fixed.capabilities.ibss", "wlan.fixed.capabilities.privacy",
		"wlan.ds.current_channel", "wlan.country_info.code",
		"wlan.rsn.akms.type", "wlan.rsn.pcs.type", "wlan.rsn.gcs.type",

		// HT Capabilities
		"wlan.ht.capabilities", "wlan.ht.info.primarychannel", "wlan.ht.info.secchanoffset",
		"wlan.ht.capabilities.ldpccoding", "wlan.ht.capabilities.width",
		"wlan.ht.capabilities.sm", "wlan.ht.capabilities.green",
		"wlan.ht.capabilities.short20", "wlan.ht.capabilities.short40",
		"wlan.ht.capabilities.txstbc", "wlan.ht.capabilities.rxstbc",
		"wlan.ht.capabilities.delayedblockack", "wlan.ht.capabilities.amsdu",
		"wlan.ht.capabilities.dsscck", "wlan.ht.capabilities.psmp",
		"wlan.ht.capabilities.40mhzintolerant", "wlan.ht.capabilities.lsig",
		"wlan.ht.ampduparam.maxlength", "wlan.ht.ampduparam.mpdudensity",
		"wlan.ht.mcsset.rxbitmask.0to7", "wlan.ht.mcsset.rxbitmask.8to15",
		"wlan.ht.mcsset.rxbitmask.16to23", "wlan.ht.mcsset.rxbitmask.24to31",

		// VHT Capabilities
		"wlan.vht.capabilities", "wlan.vht.op.channelwidth", "wlan.vht.op.channelcenter0", "wlan.vht.op.channelcenter1",
		"wlan.vht.capabilities.maxmpdulength", "wlan.vht.capabilities.supportedchanwidthset",
		"wlan.vht.capabilities.rxldpc", "wlan.vht.capabilities.short80", "wlan.vht.capabilities.short160",
		"wlan.vht.capabilities.txstbc", "wlan.vht.capabilities.rxstbc",
		"wlan.vht.capabilities.subeamformer", "wlan.vht.capabilities.subeamformee",
		"wlan.vht.capabilities.beamformee_sts_cap", "wlan.vht.capabilities.soundingdimensions",
		"wlan.vht.capabilities.mubeamformer", "wlan.vht.capabilities.mubeamformee",
		"wlan.vht.capabilities.vhttxopps", "wlan.vht.capabilities.vhthtc",
		"wlan.vht.capabilities.maxampdu", "wlan.vht.capabilities.linkadapt",
		"wlan.vht.capabilities.rxpatconsist", "wlan.vht.capabilities.txpatconsist",
		"wlan.vht.mcsset.rxmcsmap", "wlan.vht.mcsset.rxhighestlonggirate",
		"wlan.vht.mcsset.txmcsmap", "wlan.vht.mcsset.txhighestlonggirate",

		// HE Capabilities - 仅使用示例JSON中确认存在的字段
		"wlan.ext_tag.he_mac_caps", "wlan.ext_tag.he_mac_cap.htc_he_support",
		"wlan.ext_tag.he_mac_cap.twt_req_support", "wlan.ext_tag.he_mac_cap.twt_rsp_support",
		"wlan.ext_tag.he_operation.max_he_mcs_for_1_ss", "wlan.ext_tag.he_operation.max_he_mcs_for_2_ss",
		"wlan.ext_tag.he_operation.max_he_mcs_for_3_ss", "wlan.ext_tag.he_operation.max_he_mcs_for_4_ss",
		"wlan.ext_tag.he_mcs_map.rx_he_mcs_map_lte_80", "wlan.ext_tag.he_mcs_map.tx_he_mcs_map_lte_80",
		"wlan.ext_tag.he_phy_cap.su_beamformer", "wlan.ext_tag.he_phy_cap.su_beamformee",
		"wlan.ext_tag.bss_color_information.bss_color",
		// 使用示例中确认存在的HE通道宽度相关字段
		"wlan.ext_tag.he_phy_cap.chan_width_set.160_in_5ghz",
		"wlan.ext_tag.he_phy_cap.chan_width_set.160_80_80_in_5ghz",
		"wlan.ext_tag.he_phy_cap.chan_width_set.40_80_in_5ghz",

		"wlan.tim.dtim_count", "wlan.tim.dtim_period", "wlan.tim.bmapctl.multicast",
		"ip.len", "ipv6.plen", "tcp.len", "udp.length", "wlan.qos.tid",
		"wlan.duration",
	}

	executor := &TSharkExecutor{}
	stdout, stderr, err := executor.Start(pcapFilePath, tsharkPath, fields)
	if err != nil {
		return fmt.Errorf("failed to start tshark executor: %w", err)
	}
	defer executor.Stop()

	// Goroutine to log stderr from tshark
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logger.Log.Warn().Msgf("TSHARK_STDERR: %s", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_STDERR_SCAN")
		}
	}()

	csvParser, err := NewCSVParser(stdout)
	if err != nil {
		return fmt.Errorf("failed to create CSV parser: %w", err)
	}

	frameProcessor := NewFrameProcessor(csvParser.HeaderMap)
	frameCount := 0
	errorCount := 0

	// log.Println("INFO_PCAP_PROCESS: Starting to process frames from tshark output...")
	for {
		row, err := csvParser.ReadFrame()
		if err == io.EOF {
			// log.Println("INFO_PCAP_PROCESS: EOF reached in tshark CSV output.")
			break
		}
		if err != nil {
			logger.Log.Error().Err(err).Msgf("ERROR_PCAP_PROCESS: Failed to read/parse CSV row")
			errorCount++
			if errorCount > 100 && float64(errorCount)/float64(frameCount+errorCount) > 0.5 {
				logger.Log.Error().Msgf("ERROR_PCAP_PROCESS: High error rate (%d errors in %d attempts), stopping.", errorCount, frameCount+errorCount)
				return fmt.Errorf("too many CSV parsing errors")
			}
			continue // Skip this problematic row
		}

		parsedInfo, procErr := frameProcessor.ProcessRow(row)
		if procErr != nil {
			logger.Log.Warn().Err(procErr).Msgf("WARN_FRAME_PROCESS: Failed to process frame row. Raw row: %v", row)
			// Optionally, log only a subset of raw row if too verbose
			errorCount++
			continue // Skip this frame
		}

		if parsedInfo != nil {
			pktHandler(parsedInfo)
		}
		frameCount++
		if frameCount%1000 == 0 {
			logger.Log.Info().Msgf("INFO_PCAP_PROCESS: Processed %d frames...", frameCount)
		}
	}

	logger.Log.Info().Msgf("INFO_PCAP_PROCESS: Finished processing. Total frames processed: %d, Errors: %d", frameCount, errorCount)
	return nil
}

// ProcessPcapStream is the main entry point for parsing a pcap stream using tshark.
func ProcessPcapStream(pcapStream io.Reader, tsharkPath string, pktHandler PacketInfoHandler) error {
	// Define all necessary tshark fields based on the specification
	fields := []string{
		"frame.time_epoch", "frame.len", "frame.cap_len", "wlan.fc.type_subtype",
		"wlan.fc.type", "wlan.fc.subtype", "wlan.fc.retry", // Corrected: wlan.flags.retry -> wlan.fc.retry
		"wlan.ra", "wlan.da", "wlan.ta", "wlan.sa", "wlan.bssid",
		"radiotap.channel.freq", "radiotap.dbm_antsignal", "radiotap.dbm_antnoise",
		"radiotap.datarate", "radiotap.mcs.index", "radiotap.mcs.bw", "radiotap.mcs.gi",
		"radiotap.vht.bw", "radiotap.vht.gi", "wlan_radio.data_rate", "wlan.supported_rates",
		"wlan.ssid", "wlan.fixed.beacon", "wlan.fixed.capabilities.ess", "wlan.fixed.capabilities.ibss", "wlan.fixed.capabilities.privacy",
		"wlan.ds.current_channel", "wlan.country_info.code",
		"wlan.rsn.akms.type", "wlan.rsn.pcs.type", "wlan.rsn.gcs.type",

		// HT Capabilities
		"wlan.ht.capabilities", "wlan.ht.info.primarychannel", "wlan.ht.info.secchanoffset",
		"wlan.ht.capabilities.ldpccoding", "wlan.ht.capabilities.width",
		"wlan.ht.capabilities.sm", "wlan.ht.capabilities.green",
		"wlan.ht.capabilities.short20", "wlan.ht.capabilities.short40",
		"wlan.ht.capabilities.txstbc", "wlan.ht.capabilities.rxstbc",
		"wlan.ht.capabilities.delayedblockack", "wlan.ht.capabilities.amsdu",
		"wlan.ht.capabilities.dsscck", "wlan.ht.capabilities.psmp",
		"wlan.ht.capabilities.40mhzintolerant", "wlan.ht.capabilities.lsig",
		"wlan.ht.ampduparam.maxlength", "wlan.ht.ampduparam.mpdudensity",
		"wlan.ht.mcsset.rxbitmask.0to7", "wlan.ht.mcsset.rxbitmask.8to15",
		"wlan.ht.mcsset.rxbitmask.16to23", "wlan.ht.mcsset.rxbitmask.24to31",

		// VHT Capabilities
		"wlan.vht.capabilities", "wlan.vht.op.channelwidth", "wlan.vht.op.channelcenter0", "wlan.vht.op.channelcenter1",
		"wlan.vht.capabilities.maxmpdulength", "wlan.vht.capabilities.supportedchanwidthset",
		"wlan.vht.capabilities.rxldpc", "wlan.vht.capabilities.short80", "wlan.vht.capabilities.short160",
		"wlan.vht.capabilities.txstbc", "wlan.vht.capabilities.rxstbc",
		"wlan.vht.capabilities.subeamformer", "wlan.vht.capabilities.subeamformee",
		"wlan.vht.capabilities.beamformee_sts_cap", "wlan.vht.capabilities.soundingdimensions",
		"wlan.vht.capabilities.mubeamformer", "wlan.vht.capabilities.mubeamformee",
		"wlan.vht.capabilities.vhttxopps", "wlan.vht.capabilities.vhthtc",
		"wlan.vht.capabilities.maxampdu", "wlan.vht.capabilities.linkadapt",
		"wlan.vht.capabilities.rxpatconsist", "wlan.vht.capabilities.txpatconsist",
		"wlan.vht.mcsset.rxmcsmap", "wlan.vht.mcsset.rxhighestlonggirate",
		"wlan.vht.mcsset.txmcsmap", "wlan.vht.mcsset.txhighestlonggirate",

		// HE Capabilities - 仅使用示例JSON中确认存在的字段
		"wlan.ext_tag.he_mac_caps", "wlan.ext_tag.he_mac_cap.htc_he_support",
		"wlan.ext_tag.he_mac_cap.twt_req_support", "wlan.ext_tag.he_mac_cap.twt_rsp_support",
		"wlan.ext_tag.he_operation.max_he_mcs_for_1_ss", "wlan.ext_tag.he_operation.max_he_mcs_for_2_ss",
		"wlan.ext_tag.he_operation.max_he_mcs_for_3_ss", "wlan.ext_tag.he_operation.max_he_mcs_for_4_ss",
		"wlan.ext_tag.he_mcs_map.rx_he_mcs_map_lte_80", "wlan.ext_tag.he_mcs_map.tx_he_mcs_map_lte_80",
		"wlan.ext_tag.he_phy_cap.su_beamformer", "wlan.ext_tag.he_phy_cap.su_beamformee",
		"wlan.ext_tag.bss_color_information.bss_color",
		// 使用示例中确认存在的HE通道宽度相关字段
		"wlan.ext_tag.he_phy_cap.chan_width_set.160_in_5ghz",
		"wlan.ext_tag.he_phy_cap.chan_width_set.160_80_80_in_5ghz",
		"wlan.ext_tag.he_phy_cap.chan_width_set.40_80_in_5ghz",

		"wlan.tim.dtim_count", "wlan.tim.dtim_period", "wlan.tim.bmapctl.multicast",
		"ip.len", "ipv6.plen", "tcp.len", "udp.length", "wlan.qos.tid",
		"wlan.duration",
	}

	executor := &TSharkExecutor{}
	// Use StartStream instead of Start
	stdout, stderr, err := executor.StartStream(pcapStream, tsharkPath, fields)
	if err != nil {
		return fmt.Errorf("failed to start tshark executor for stream: %w", err)
	}
	defer executor.Stop()

	// Goroutine to log stderr from tshark
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logger.Log.Warn().Msgf("TSHARK_STDERR_STREAM: %s", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			logger.Log.Error().Err(err).Msgf("ERROR_TSHARK_STDERR_STREAM_SCAN")
		}
	}()

	csvParser, err := NewCSVParser(stdout)
	if err != nil {
		return fmt.Errorf("failed to create CSV parser for stream: %w", err)
	}

	frameProcessor := NewFrameProcessor(csvParser.HeaderMap)
	frameCount := 0
	errorCount := 0

	// log.Println("INFO_PCAP_STREAM_PROCESS: Starting to process frames from tshark stream output...")
	for {
		row, err := csvParser.ReadFrame()
		if err == io.EOF {
			// log.Println("INFO_PCAP_STREAM_PROCESS: EOF reached in tshark CSV stream output.")
			break
		}
		if err != nil {
			logger.Log.Error().Err(err).Msgf("ERROR_PCAP_STREAM_PROCESS: Failed to read/parse CSV row from stream")
			errorCount++
			if errorCount > 100 && float64(errorCount)/float64(frameCount+errorCount) > 0.5 {
				logger.Log.Error().Msgf("ERROR_PCAP_STREAM_PROCESS: High error rate (%d errors in %d attempts) from stream, stopping.", errorCount, frameCount+errorCount)
				return fmt.Errorf("too many CSV parsing errors from stream")
			}
			continue // Skip this problematic row
		}

		parsedInfo, procErr := frameProcessor.ProcessRow(row)
		if procErr != nil {
			logger.Log.Warn().Err(procErr).Msgf("WARN_FRAME_STREAM_PROCESS: Failed to process frame row from stream. Raw row: %v", row)
			errorCount++
			continue // Skip this frame
		}

		if parsedInfo != nil {
			pktHandler(parsedInfo)
		}
		frameCount++
		if frameCount%1000 == 0 {
			logger.Log.Info().Msgf("INFO_PCAP_STREAM_PROCESS: Processed %d frames from stream...", frameCount)
		}
	}

	logger.Log.Info().Msgf("INFO_PCAP_STREAM_PROCESS: Finished processing stream. Total frames processed: %d, Errors: %d", frameCount, errorCount)
	return nil
}

// getPHYRateMbps estimates the PHY rate in Mbps based on tshark fields.
func getPHYRateMbps(info *ParsedFrameInfo) float64 {
	// Priority: HE -> VHT -> HT -> Legacy (radiotap.datarate)
	// This is a simplified MVP version. More complex rate calculations exist.

	// VHT
	if info.RadiotapVHTNSS > 0 && info.RadiotapVHTMCS <= 9 { // Max MCS for VHT is 9
		nss := float64(info.RadiotapVHTNSS)
		mcs := float64(info.RadiotapVHTMCS)
		var baseRate float64

		// Determine base rate based on BW (simplified from 802.11ac tables)
		// These are for single stream, MCS0, Long GI.
		switch info.RadiotapVHTBw {
		case "20":
			baseRate = 6.5 // MCS0, 20MHz, NSS1, Long GI
		case "40":
			baseRate = 13.5 // MCS0, 40MHz, NSS1, Long GI
		case "80":
			baseRate = 29.3 // MCS0, 80MHz, NSS1, Long GI (approx)
		case "160":
			baseRate = 58.5 // MCS0, 160MHz, NSS1, Long GI (approx)
		default:
			// Fall through if BW string is not recognized or empty
		}

		if baseRate > 0 {
			// Adjust for actual MCS (very simplified scaling)
			// Real VHT rates depend on coding, modulation, etc.
			rate := baseRate * (mcs + 1) / 1.0 * nss // Simplified: (MCS index + 1) * base for MCS0
			if info.RadiotapVHTShortGI {
				rate *= 1.11 // Approx 10-11% increase for Short GI
			}
			return rate
		}
	}

	// HT
	if info.RadiotapMCSIndex <= 31 { // HT MCS indices 0-31
		mcs := float64(info.RadiotapMCSIndex)
		var baseRate float64 = 6.5   // MCS0, 20MHz, NSS1, Long GI
		if info.RadiotapMCSBw == 1 { // 40MHz
			baseRate = 13.5 // MCS0, 40MHz, NSS1, Long GI
		}
		// Assuming NSS=1 for simplicity if not explicitly available for HT from tshark
		// HT MCS rates are complex (e.g. MCS0-7 for NSS1, MCS8-15 for NSS2 etc.)
		// This is a very rough estimate.
		rate := baseRate * (mcs/8 + 1) // Very rough scaling by NSS group
		if info.RadiotapMCSGI {        // Short GI
			rate *= 1.11
		}
		return rate
	}

	// Legacy
	if info.RadiotapDataRate > 0 {
		return info.RadiotapDataRate // radiotap.datarate is already in Mbps
	}

	// Fallback
	if info.FrameType != "" && strings.HasPrefix(info.FrameType, "Mgmt") {
		return 6.0 // Common base rate for management frames
	}
	return 1.0 // Absolute fallback
}

// CalculateFrameAirtime estimates the airtime of a given 802.11 frame.
// This is a simplified model.
func CalculateFrameAirtime(frameLengthBytes int, phyRateMbps float64, isShortPreamble bool, isShortGI bool) time.Duration {
	if phyRateMbps <= 0 {
		return 0
	}
	dataTxTimeMicroseconds := float64(frameLengthBytes*8) / phyRateMbps
	preamblePlcpTimeMicroseconds := 192.0 // Long Preamble
	if isShortPreamble {
		preamblePlcpTimeMicroseconds = 96.0
	}
	sifsMicroseconds := 10.0 // OFDM
	giFactor := 1.0
	if isShortGI {
		giFactor = 0.9 // Approximation
	}
	totalMicroseconds := (preamblePlcpTimeMicroseconds + dataTxTimeMicroseconds*giFactor) + sifsMicroseconds
	return time.Duration(totalMicroseconds * float64(time.Microsecond))
}
