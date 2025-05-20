type Dot11
added in v1.1.4
type Dot11 struct {
	BaseLayer
	Type           Dot11Type
	Proto          uint8
	Flags          Dot11Flags
	DurationID     uint16
	Address1       net.HardwareAddr
	Address2       net.HardwareAddr
	Address3       net.HardwareAddr
	Address4       net.HardwareAddr
	SequenceNumber uint16
	FragmentNumber uint16
	Checksum       uint32
	QOS            *Dot11QOS
	HTControl      *Dot11HTControl
	DataLayer      gopacket.Layer
}
Dot11 provides an IEEE 802.11 base packet header. See http://standards.ieee.org/findstds/standard/802.11-2012.html for excruciating detail.

func (*Dot11) CanDecode ¶
added in v1.1.4
func (m *Dot11) CanDecode() gopacket.LayerClass
func (*Dot11) ChecksumValid ¶
added in v1.1.4
func (m *Dot11) ChecksumValid() bool
func (*Dot11) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11) LayerType ¶
added in v1.1.4
func (m *Dot11) LayerType() gopacket.LayerType
func (*Dot11) NextLayerType ¶
added in v1.1.4
func (m *Dot11) NextLayerType() gopacket.LayerType
func (Dot11) SerializeTo ¶
added in v1.1.12
func (m Dot11) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11ASEL ¶
added in v1.1.15
type Dot11ASEL struct {
	Command uint8
	Data    uint8
}
type Dot11AckPolicy ¶
added in v1.1.4
type Dot11AckPolicy uint8
const (
	Dot11AckPolicyNormal     Dot11AckPolicy = 0
	Dot11AckPolicyNone       Dot11AckPolicy = 1
	Dot11AckPolicyNoExplicit Dot11AckPolicy = 2
	Dot11AckPolicyBlock      Dot11AckPolicy = 3
)
func (Dot11AckPolicy) String ¶
added in v1.1.4
func (a Dot11AckPolicy) String() string
String provides a human readable string for Dot11AckPolicy. This string is possibly subject to change over time; if you're storing this persistently, you should probably store the Dot11AckPolicy value, not its string.

type Dot11Algorithm ¶
added in v1.1.4
type Dot11Algorithm uint16
const (
	Dot11AlgorithmOpen      Dot11Algorithm = 0
	Dot11AlgorithmSharedKey Dot11Algorithm = 1
)
func (Dot11Algorithm) String ¶
added in v1.1.4
func (a Dot11Algorithm) String() string
String provides a human readable string for Dot11Algorithm. This string is possibly subject to change over time; if you're storing this persistently, you should probably store the Dot11Algorithm value, not its string.

type Dot11CodingType ¶
added in v1.1.15
type Dot11CodingType uint8
func (Dot11CodingType) String ¶
added in v1.1.15
func (a Dot11CodingType) String() string
type Dot11Ctrl ¶
added in v1.1.4
type Dot11Ctrl struct {
	BaseLayer
}
Dot11Ctrl is a base for all IEEE 802.11 control layers.

func (*Dot11Ctrl) CanDecode ¶
added in v1.1.4
func (m *Dot11Ctrl) CanDecode() gopacket.LayerClass
func (*Dot11Ctrl) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11Ctrl) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11Ctrl) LayerType ¶
added in v1.1.4
func (m *Dot11Ctrl) LayerType() gopacket.LayerType
func (*Dot11Ctrl) NextLayerType ¶
added in v1.1.4
func (m *Dot11Ctrl) NextLayerType() gopacket.LayerType
type Dot11CtrlAck ¶
added in v1.1.4
type Dot11CtrlAck struct {
	Dot11Ctrl
}
func (*Dot11CtrlAck) CanDecode ¶
added in v1.1.4
func (m *Dot11CtrlAck) CanDecode() gopacket.LayerClass
func (*Dot11CtrlAck) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11CtrlAck) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11CtrlAck) LayerType ¶
added in v1.1.4
func (m *Dot11CtrlAck) LayerType() gopacket.LayerType
type Dot11CtrlBlockAck ¶
added in v1.1.4
type Dot11CtrlBlockAck struct {
	Dot11Ctrl
}
func (*Dot11CtrlBlockAck) CanDecode ¶
added in v1.1.4
func (m *Dot11CtrlBlockAck) CanDecode() gopacket.LayerClass
func (*Dot11CtrlBlockAck) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11CtrlBlockAck) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11CtrlBlockAck) LayerType ¶
added in v1.1.4
func (m *Dot11CtrlBlockAck) LayerType() gopacket.LayerType
type Dot11CtrlBlockAckReq ¶
added in v1.1.4
type Dot11CtrlBlockAckReq struct {
	Dot11Ctrl
}
func (*Dot11CtrlBlockAckReq) CanDecode ¶
added in v1.1.4
func (m *Dot11CtrlBlockAckReq) CanDecode() gopacket.LayerClass
func (*Dot11CtrlBlockAckReq) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11CtrlBlockAckReq) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11CtrlBlockAckReq) LayerType ¶
added in v1.1.4
func (m *Dot11CtrlBlockAckReq) LayerType() gopacket.LayerType
type Dot11CtrlCFEnd ¶
added in v1.1.4
type Dot11CtrlCFEnd struct {
	Dot11Ctrl
}
func (*Dot11CtrlCFEnd) CanDecode ¶
added in v1.1.4
func (m *Dot11CtrlCFEnd) CanDecode() gopacket.LayerClass
func (*Dot11CtrlCFEnd) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11CtrlCFEnd) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11CtrlCFEnd) LayerType ¶
added in v1.1.4
func (m *Dot11CtrlCFEnd) LayerType() gopacket.LayerType
type Dot11CtrlCFEndAck ¶
added in v1.1.4
type Dot11CtrlCFEndAck struct {
	Dot11Ctrl
}
func (*Dot11CtrlCFEndAck) CanDecode ¶
added in v1.1.4
func (m *Dot11CtrlCFEndAck) CanDecode() gopacket.LayerClass
func (*Dot11CtrlCFEndAck) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11CtrlCFEndAck) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11CtrlCFEndAck) LayerType ¶
added in v1.1.4
func (m *Dot11CtrlCFEndAck) LayerType() gopacket.LayerType
type Dot11CtrlCTS ¶
added in v1.1.4
type Dot11CtrlCTS struct {
	Dot11Ctrl
}
func (*Dot11CtrlCTS) CanDecode ¶
added in v1.1.4
func (m *Dot11CtrlCTS) CanDecode() gopacket.LayerClass
func (*Dot11CtrlCTS) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11CtrlCTS) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11CtrlCTS) LayerType ¶
added in v1.1.4
func (m *Dot11CtrlCTS) LayerType() gopacket.LayerType
type Dot11CtrlPowersavePoll ¶
added in v1.1.4
type Dot11CtrlPowersavePoll struct {
	Dot11Ctrl
}
func (*Dot11CtrlPowersavePoll) CanDecode ¶
added in v1.1.4
func (m *Dot11CtrlPowersavePoll) CanDecode() gopacket.LayerClass
func (*Dot11CtrlPowersavePoll) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11CtrlPowersavePoll) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11CtrlPowersavePoll) LayerType ¶
added in v1.1.4
func (m *Dot11CtrlPowersavePoll) LayerType() gopacket.LayerType
type Dot11CtrlRTS ¶
added in v1.1.4
type Dot11CtrlRTS struct {
	Dot11Ctrl
}
func (*Dot11CtrlRTS) CanDecode ¶
added in v1.1.4
func (m *Dot11CtrlRTS) CanDecode() gopacket.LayerClass
func (*Dot11CtrlRTS) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11CtrlRTS) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11CtrlRTS) LayerType ¶
added in v1.1.4
func (m *Dot11CtrlRTS) LayerType() gopacket.LayerType
type Dot11Data ¶
added in v1.1.4
type Dot11Data struct {
	BaseLayer
}
Dot11Data is a base for all IEEE 802.11 data layers.

func (*Dot11Data) CanDecode ¶
added in v1.1.4
func (m *Dot11Data) CanDecode() gopacket.LayerClass
func (*Dot11Data) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11Data) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11Data) LayerType ¶
added in v1.1.4
func (m *Dot11Data) LayerType() gopacket.LayerType
func (*Dot11Data) NextLayerType ¶
added in v1.1.4
func (m *Dot11Data) NextLayerType() gopacket.LayerType
type Dot11DataCFAck ¶
added in v1.1.4
type Dot11DataCFAck struct {
	Dot11Data
}
func (*Dot11DataCFAck) CanDecode ¶
added in v1.1.4
func (m *Dot11DataCFAck) CanDecode() gopacket.LayerClass
func (*Dot11DataCFAck) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11DataCFAck) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11DataCFAck) LayerType ¶
added in v1.1.4
func (m *Dot11DataCFAck) LayerType() gopacket.LayerType
type Dot11DataCFAckNoData ¶
added in v1.1.4
type Dot11DataCFAckNoData struct {
	Dot11Data
}
func (*Dot11DataCFAckNoData) CanDecode ¶
added in v1.1.4
func (m *Dot11DataCFAckNoData) CanDecode() gopacket.LayerClass
func (*Dot11DataCFAckNoData) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11DataCFAckNoData) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11DataCFAckNoData) LayerType ¶
added in v1.1.4
func (m *Dot11DataCFAckNoData) LayerType() gopacket.LayerType
type Dot11DataCFAckPoll ¶
added in v1.1.4
type Dot11DataCFAckPoll struct {
	Dot11Data
}
func (*Dot11DataCFAckPoll) CanDecode ¶
added in v1.1.4
func (m *Dot11DataCFAckPoll) CanDecode() gopacket.LayerClass
func (*Dot11DataCFAckPoll) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11DataCFAckPoll) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11DataCFAckPoll) LayerType ¶
added in v1.1.4
func (m *Dot11DataCFAckPoll) LayerType() gopacket.LayerType
type Dot11DataCFAckPollNoData ¶
added in v1.1.4
type Dot11DataCFAckPollNoData struct {
	Dot11Data
}
func (*Dot11DataCFAckPollNoData) CanDecode ¶
added in v1.1.4
func (m *Dot11DataCFAckPollNoData) CanDecode() gopacket.LayerClass
func (*Dot11DataCFAckPollNoData) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11DataCFAckPollNoData) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11DataCFAckPollNoData) LayerType ¶
added in v1.1.4
func (m *Dot11DataCFAckPollNoData) LayerType() gopacket.LayerType
type Dot11DataCFPoll ¶
added in v1.1.4
type Dot11DataCFPoll struct {
	Dot11Data
}
func (*Dot11DataCFPoll) CanDecode ¶
added in v1.1.4
func (m *Dot11DataCFPoll) CanDecode() gopacket.LayerClass
func (*Dot11DataCFPoll) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11DataCFPoll) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11DataCFPoll) LayerType ¶
added in v1.1.4
func (m *Dot11DataCFPoll) LayerType() gopacket.LayerType
type Dot11DataCFPollNoData ¶
added in v1.1.4
type Dot11DataCFPollNoData struct {
	Dot11Data
}
func (*Dot11DataCFPollNoData) CanDecode ¶
added in v1.1.4
func (m *Dot11DataCFPollNoData) CanDecode() gopacket.LayerClass
func (*Dot11DataCFPollNoData) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11DataCFPollNoData) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11DataCFPollNoData) LayerType ¶
added in v1.1.4
func (m *Dot11DataCFPollNoData) LayerType() gopacket.LayerType
type Dot11DataNull ¶
added in v1.1.4
type Dot11DataNull struct {
	Dot11Data
}
func (*Dot11DataNull) CanDecode ¶
added in v1.1.4
func (m *Dot11DataNull) CanDecode() gopacket.LayerClass
func (*Dot11DataNull) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11DataNull) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11DataNull) LayerType ¶
added in v1.1.4
func (m *Dot11DataNull) LayerType() gopacket.LayerType
type Dot11DataQOS ¶
added in v1.1.4
type Dot11DataQOS struct {
	Dot11Ctrl
}
func (*Dot11DataQOS) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11DataQOS) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
type Dot11DataQOSCFAckPollNoData ¶
added in v1.1.4
type Dot11DataQOSCFAckPollNoData struct {
	Dot11DataQOS
}
func (*Dot11DataQOSCFAckPollNoData) CanDecode ¶
added in v1.1.4
func (m *Dot11DataQOSCFAckPollNoData) CanDecode() gopacket.LayerClass
func (*Dot11DataQOSCFAckPollNoData) LayerType ¶
added in v1.1.4
func (m *Dot11DataQOSCFAckPollNoData) LayerType() gopacket.LayerType
func (*Dot11DataQOSCFAckPollNoData) NextLayerType ¶
added in v1.1.4
func (m *Dot11DataQOSCFAckPollNoData) NextLayerType() gopacket.LayerType
type Dot11DataQOSCFPollNoData ¶
added in v1.1.4
type Dot11DataQOSCFPollNoData struct {
	Dot11DataQOS
}
func (*Dot11DataQOSCFPollNoData) CanDecode ¶
added in v1.1.4
func (m *Dot11DataQOSCFPollNoData) CanDecode() gopacket.LayerClass
func (*Dot11DataQOSCFPollNoData) LayerType ¶
added in v1.1.4
func (m *Dot11DataQOSCFPollNoData) LayerType() gopacket.LayerType
func (*Dot11DataQOSCFPollNoData) NextLayerType ¶
added in v1.1.4
func (m *Dot11DataQOSCFPollNoData) NextLayerType() gopacket.LayerType
type Dot11DataQOSData ¶
added in v1.1.4
type Dot11DataQOSData struct {
	Dot11DataQOS
}
func (*Dot11DataQOSData) CanDecode ¶
added in v1.1.4
func (m *Dot11DataQOSData) CanDecode() gopacket.LayerClass
func (*Dot11DataQOSData) LayerType ¶
added in v1.1.4
func (m *Dot11DataQOSData) LayerType() gopacket.LayerType
func (*Dot11DataQOSData) NextLayerType ¶
added in v1.1.4
func (m *Dot11DataQOSData) NextLayerType() gopacket.LayerType
type Dot11DataQOSDataCFAck ¶
added in v1.1.4
type Dot11DataQOSDataCFAck struct {
	Dot11DataQOS
}
func (*Dot11DataQOSDataCFAck) CanDecode ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFAck) CanDecode() gopacket.LayerClass
func (*Dot11DataQOSDataCFAck) LayerType ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFAck) LayerType() gopacket.LayerType
func (*Dot11DataQOSDataCFAck) NextLayerType ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFAck) NextLayerType() gopacket.LayerType
type Dot11DataQOSDataCFAckPoll ¶
added in v1.1.4
type Dot11DataQOSDataCFAckPoll struct {
	Dot11DataQOS
}
func (*Dot11DataQOSDataCFAckPoll) CanDecode ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFAckPoll) CanDecode() gopacket.LayerClass
func (*Dot11DataQOSDataCFAckPoll) LayerType ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFAckPoll) LayerType() gopacket.LayerType
func (*Dot11DataQOSDataCFAckPoll) NextLayerType ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFAckPoll) NextLayerType() gopacket.LayerType
type Dot11DataQOSDataCFPoll ¶
added in v1.1.4
type Dot11DataQOSDataCFPoll struct {
	Dot11DataQOS
}
func (*Dot11DataQOSDataCFPoll) CanDecode ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFPoll) CanDecode() gopacket.LayerClass
func (*Dot11DataQOSDataCFPoll) LayerType ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFPoll) LayerType() gopacket.LayerType
func (*Dot11DataQOSDataCFPoll) NextLayerType ¶
added in v1.1.4
func (m *Dot11DataQOSDataCFPoll) NextLayerType() gopacket.LayerType
type Dot11DataQOSNull ¶
added in v1.1.4
type Dot11DataQOSNull struct {
	Dot11DataQOS
}
func (*Dot11DataQOSNull) CanDecode ¶
added in v1.1.4
func (m *Dot11DataQOSNull) CanDecode() gopacket.LayerClass
func (*Dot11DataQOSNull) LayerType ¶
added in v1.1.4
func (m *Dot11DataQOSNull) LayerType() gopacket.LayerType
func (*Dot11DataQOSNull) NextLayerType ¶
added in v1.1.4
func (m *Dot11DataQOSNull) NextLayerType() gopacket.LayerType
type Dot11Flags ¶
added in v1.1.4
type Dot11Flags uint8
Dot11Flags contains the set of 8 flags in the IEEE 802.11 frame control header, all in one place.

const (
	Dot11FlagsToDS Dot11Flags = 1 << iota
	Dot11FlagsFromDS
	Dot11FlagsMF
	Dot11FlagsRetry
	Dot11FlagsPowerManagement
	Dot11FlagsMD
	Dot11FlagsWEP
	Dot11FlagsOrder
)
func (Dot11Flags) FromDS ¶
added in v1.1.4
func (d Dot11Flags) FromDS() bool
func (Dot11Flags) MD ¶
added in v1.1.4
func (d Dot11Flags) MD() bool
func (Dot11Flags) MF ¶
added in v1.1.4
func (d Dot11Flags) MF() bool
func (Dot11Flags) Order ¶
added in v1.1.4
func (d Dot11Flags) Order() bool
func (Dot11Flags) PowerManagement ¶
added in v1.1.4
func (d Dot11Flags) PowerManagement() bool
func (Dot11Flags) Retry ¶
added in v1.1.4
func (d Dot11Flags) Retry() bool
func (Dot11Flags) String ¶
added in v1.1.4
func (a Dot11Flags) String() string
String provides a human readable string for Dot11Flags. This string is possibly subject to change over time; if you're storing this persistently, you should probably store the Dot11Flags value, not its string.

func (Dot11Flags) ToDS ¶
added in v1.1.4
func (d Dot11Flags) ToDS() bool
func (Dot11Flags) WEP ¶
added in v1.1.4
func (d Dot11Flags) WEP() bool
type Dot11HTControl ¶
added in v1.1.15
type Dot11HTControl struct {
	ACConstraint bool
	RDGMorePPDU  bool

	VHT *Dot11HTControlVHT
	HT  *Dot11HTControlHT
}
type Dot11HTControlHT ¶
added in v1.1.15
type Dot11HTControlHT struct {
	LinkAdapationControl *Dot11LinkAdapationControl
	CalibrationPosition  uint8
	CalibrationSequence  uint8
	CSISteering          uint8
	NDPAnnouncement      bool
	DEI                  bool
}
type Dot11HTControlMFB ¶
added in v1.1.15
type Dot11HTControlMFB struct {
	NumSTS uint8
	VHTMCS uint8
	BW     uint8
	SNR    int8
}
func (*Dot11HTControlMFB) NoFeedBackPresent ¶
added in v1.1.15
func (m *Dot11HTControlMFB) NoFeedBackPresent() bool
type Dot11HTControlVHT ¶
added in v1.1.15
type Dot11HTControlVHT struct {
	MRQ            bool
	UnsolicitedMFB bool
	MSI            *uint8
	MFB            Dot11HTControlMFB
	CompressedMSI  *uint8
	STBCIndication bool
	MFSI           *uint8
	GID            *uint8
	CodingType     *Dot11CodingType
	FbTXBeamformed bool
}
type Dot11InformationElement ¶
added in v1.1.4
type Dot11InformationElement struct {
	BaseLayer
	ID     Dot11InformationElementID
	Length uint8
	OUI    []byte
	Info   []byte
}
func (*Dot11InformationElement) CanDecode ¶
added in v1.1.4
func (m *Dot11InformationElement) CanDecode() gopacket.LayerClass
func (*Dot11InformationElement) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11InformationElement) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11InformationElement) LayerType ¶
added in v1.1.4
func (m *Dot11InformationElement) LayerType() gopacket.LayerType
func (*Dot11InformationElement) NextLayerType ¶
added in v1.1.4
func (m *Dot11InformationElement) NextLayerType() gopacket.LayerType
func (Dot11InformationElement) SerializeTo ¶
added in v1.1.11
func (m Dot11InformationElement) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
func (*Dot11InformationElement) String ¶
added in v1.1.4
func (d *Dot11InformationElement) String() string
type Dot11InformationElementID ¶
added in v1.1.4
type Dot11InformationElementID uint8
const (
	Dot11InformationElementIDSSID                      Dot11InformationElementID = 0
	Dot11InformationElementIDRates                     Dot11InformationElementID = 1
	Dot11InformationElementIDFHSet                     Dot11InformationElementID = 2
	Dot11InformationElementIDDSSet                     Dot11InformationElementID = 3
	Dot11InformationElementIDCFSet                     Dot11InformationElementID = 4
	Dot11InformationElementIDTIM                       Dot11InformationElementID = 5
	Dot11InformationElementIDIBSSSet                   Dot11InformationElementID = 6
	Dot11InformationElementIDCountryInfo               Dot11InformationElementID = 7
	Dot11InformationElementIDHoppingPatternParam       Dot11InformationElementID = 8
	Dot11InformationElementIDHoppingPatternTable       Dot11InformationElementID = 9
	Dot11InformationElementIDRequest                   Dot11InformationElementID = 10
	Dot11InformationElementIDQBSSLoadElem              Dot11InformationElementID = 11
	Dot11InformationElementIDEDCAParamSet              Dot11InformationElementID = 12
	Dot11InformationElementIDTrafficSpec               Dot11InformationElementID = 13
	Dot11InformationElementIDTrafficClass              Dot11InformationElementID = 14
	Dot11InformationElementIDSchedule                  Dot11InformationElementID = 15
	Dot11InformationElementIDChallenge                 Dot11InformationElementID = 16
	Dot11InformationElementIDPowerConst                Dot11InformationElementID = 32
	Dot11InformationElementIDPowerCapability           Dot11InformationElementID = 33
	Dot11InformationElementIDTPCRequest                Dot11InformationElementID = 34
	Dot11InformationElementIDTPCReport                 Dot11InformationElementID = 35
	Dot11InformationElementIDSupportedChannels         Dot11InformationElementID = 36
	Dot11InformationElementIDSwitchChannelAnnounce     Dot11InformationElementID = 37
	Dot11InformationElementIDMeasureRequest            Dot11InformationElementID = 38
	Dot11InformationElementIDMeasureReport             Dot11InformationElementID = 39
	Dot11InformationElementIDQuiet                     Dot11InformationElementID = 40
	Dot11InformationElementIDIBSSDFS                   Dot11InformationElementID = 41
	Dot11InformationElementIDERPInfo                   Dot11InformationElementID = 42
	Dot11InformationElementIDTSDelay                   Dot11InformationElementID = 43
	Dot11InformationElementIDTCLASProcessing           Dot11InformationElementID = 44
	Dot11InformationElementIDHTCapabilities            Dot11InformationElementID = 45
	Dot11InformationElementIDQOSCapability             Dot11InformationElementID = 46
	Dot11InformationElementIDERPInfo2                  Dot11InformationElementID = 47
	Dot11InformationElementIDRSNInfo                   Dot11InformationElementID = 48
	Dot11InformationElementIDESRates                   Dot11InformationElementID = 50
	Dot11InformationElementIDAPChannelReport           Dot11InformationElementID = 51
	Dot11InformationElementIDNeighborReport            Dot11InformationElementID = 52
	Dot11InformationElementIDRCPI                      Dot11InformationElementID = 53
	Dot11InformationElementIDMobilityDomain            Dot11InformationElementID = 54
	Dot11InformationElementIDFastBSSTrans              Dot11InformationElementID = 55
	Dot11InformationElementIDTimeoutInt                Dot11InformationElementID = 56
	Dot11InformationElementIDRICData                   Dot11InformationElementID = 57
	Dot11InformationElementIDDSERegisteredLoc          Dot11InformationElementID = 58
	Dot11InformationElementIDSuppOperatingClass        Dot11InformationElementID = 59
	Dot11InformationElementIDExtChanSwitchAnnounce     Dot11InformationElementID = 60
	Dot11InformationElementIDHTInfo                    Dot11InformationElementID = 61
	Dot11InformationElementIDSecChanOffset             Dot11InformationElementID = 62
	Dot11InformationElementIDBSSAverageAccessDelay     Dot11InformationElementID = 63
	Dot11InformationElementIDAntenna                   Dot11InformationElementID = 64
	Dot11InformationElementIDRSNI                      Dot11InformationElementID = 65
	Dot11InformationElementIDMeasurePilotTrans         Dot11InformationElementID = 66
	Dot11InformationElementIDBSSAvailAdmCapacity       Dot11InformationElementID = 67
	Dot11InformationElementIDBSSACAccDelayWAPIParam    Dot11InformationElementID = 68
	Dot11InformationElementIDTimeAdvertisement         Dot11InformationElementID = 69
	Dot11InformationElementIDRMEnabledCapabilities     Dot11InformationElementID = 70
	Dot11InformationElementIDMultipleBSSID             Dot11InformationElementID = 71
	Dot11InformationElementID2040BSSCoExist            Dot11InformationElementID = 72
	Dot11InformationElementID2040BSSIntChanReport      Dot11InformationElementID = 73
	Dot11InformationElementIDOverlapBSSScanParam       Dot11InformationElementID = 74
	Dot11InformationElementIDRICDescriptor             Dot11InformationElementID = 75
	Dot11InformationElementIDManagementMIC             Dot11InformationElementID = 76
	Dot11InformationElementIDEventRequest              Dot11InformationElementID = 78
	Dot11InformationElementIDEventReport               Dot11InformationElementID = 79
	Dot11InformationElementIDDiagnosticRequest         Dot11InformationElementID = 80
	Dot11InformationElementIDDiagnosticReport          Dot11InformationElementID = 81
	Dot11InformationElementIDLocationParam             Dot11InformationElementID = 82
	Dot11InformationElementIDNonTransBSSIDCapability   Dot11InformationElementID = 83
	Dot11InformationElementIDSSIDList                  Dot11InformationElementID = 84
	Dot11InformationElementIDMultipleBSSIDIndex        Dot11InformationElementID = 85
	Dot11InformationElementIDFMSDescriptor             Dot11InformationElementID = 86
	Dot11InformationElementIDFMSRequest                Dot11InformationElementID = 87
	Dot11InformationElementIDFMSResponse               Dot11InformationElementID = 88
	Dot11InformationElementIDQOSTrafficCapability      Dot11InformationElementID = 89
	Dot11InformationElementIDBSSMaxIdlePeriod          Dot11InformationElementID = 90
	Dot11InformationElementIDTFSRequest                Dot11InformationElementID = 91
	Dot11InformationElementIDTFSResponse               Dot11InformationElementID = 92
	Dot11InformationElementIDWNMSleepMode              Dot11InformationElementID = 93
	Dot11InformationElementIDTIMBroadcastRequest       Dot11InformationElementID = 94
	Dot11InformationElementIDTIMBroadcastResponse      Dot11InformationElementID = 95
	Dot11InformationElementIDCollInterferenceReport    Dot11InformationElementID = 96
	Dot11InformationElementIDChannelUsage              Dot11InformationElementID = 97
	Dot11InformationElementIDTimeZone                  Dot11InformationElementID = 98
	Dot11InformationElementIDDMSRequest                Dot11InformationElementID = 99
	Dot11InformationElementIDDMSResponse               Dot11InformationElementID = 100
	Dot11InformationElementIDLinkIdentifier            Dot11InformationElementID = 101
	Dot11InformationElementIDWakeupSchedule            Dot11InformationElementID = 102
	Dot11InformationElementIDChannelSwitchTiming       Dot11InformationElementID = 104
	Dot11InformationElementIDPTIControl                Dot11InformationElementID = 105
	Dot11InformationElementIDPUBufferStatus            Dot11InformationElementID = 106
	Dot11InformationElementIDInterworking              Dot11InformationElementID = 107
	Dot11InformationElementIDAdvertisementProtocol     Dot11InformationElementID = 108
	Dot11InformationElementIDExpBWRequest              Dot11InformationElementID = 109
	Dot11InformationElementIDQOSMapSet                 Dot11InformationElementID = 110
	Dot11InformationElementIDRoamingConsortium         Dot11InformationElementID = 111
	Dot11InformationElementIDEmergencyAlertIdentifier  Dot11InformationElementID = 112
	Dot11InformationElementIDMeshConfiguration         Dot11InformationElementID = 113
	Dot11InformationElementIDMeshID                    Dot11InformationElementID = 114
	Dot11InformationElementIDMeshLinkMetricReport      Dot11InformationElementID = 115
	Dot11InformationElementIDCongestionNotification    Dot11InformationElementID = 116
	Dot11InformationElementIDMeshPeeringManagement     Dot11InformationElementID = 117
	Dot11InformationElementIDMeshChannelSwitchParam    Dot11InformationElementID = 118
	Dot11InformationElementIDMeshAwakeWindows          Dot11InformationElementID = 119
	Dot11InformationElementIDBeaconTiming              Dot11InformationElementID = 120
	Dot11InformationElementIDMCCAOPSetupRequest        Dot11InformationElementID = 121
	Dot11InformationElementIDMCCAOPSetupReply          Dot11InformationElementID = 122
	Dot11InformationElementIDMCCAOPAdvertisement       Dot11InformationElementID = 123
	Dot11InformationElementIDMCCAOPTeardown            Dot11InformationElementID = 124
	Dot11InformationElementIDGateAnnouncement          Dot11InformationElementID = 125
	Dot11InformationElementIDRootAnnouncement          Dot11InformationElementID = 126
	Dot11InformationElementIDExtCapability             Dot11InformationElementID = 127
	Dot11InformationElementIDAgereProprietary          Dot11InformationElementID = 128
	Dot11InformationElementIDPathRequest               Dot11InformationElementID = 130
	Dot11InformationElementIDPathReply                 Dot11InformationElementID = 131
	Dot11InformationElementIDPathError                 Dot11InformationElementID = 132
	Dot11InformationElementIDCiscoCCX1CKIPDeviceName   Dot11InformationElementID = 133
	Dot11InformationElementIDCiscoCCX2                 Dot11InformationElementID = 136
	Dot11InformationElementIDProxyUpdate               Dot11InformationElementID = 137
	Dot11InformationElementIDProxyUpdateConfirmation   Dot11InformationElementID = 138
	Dot11InformationElementIDAuthMeshPerringExch       Dot11InformationElementID = 139
	Dot11InformationElementIDMIC                       Dot11InformationElementID = 140
	Dot11InformationElementIDDestinationURI            Dot11InformationElementID = 141
	Dot11InformationElementIDUAPSDCoexistence          Dot11InformationElementID = 142
	Dot11InformationElementIDWakeupSchedule80211ad     Dot11InformationElementID = 143
	Dot11InformationElementIDExtendedSchedule          Dot11InformationElementID = 144
	Dot11InformationElementIDSTAAvailability           Dot11InformationElementID = 145
	Dot11InformationElementIDDMGTSPEC                  Dot11InformationElementID = 146
	Dot11InformationElementIDNextDMGATI                Dot11InformationElementID = 147
	Dot11InformationElementIDDMSCapabilities           Dot11InformationElementID = 148
	Dot11InformationElementIDCiscoUnknown95            Dot11InformationElementID = 149
	Dot11InformationElementIDVendor2                   Dot11InformationElementID = 150
	Dot11InformationElementIDDMGOperating              Dot11InformationElementID = 151
	Dot11InformationElementIDDMGBSSParamChange         Dot11InformationElementID = 152
	Dot11InformationElementIDDMGBeamRefinement         Dot11InformationElementID = 153
	Dot11InformationElementIDChannelMeasFeedback       Dot11InformationElementID = 154
	Dot11InformationElementIDAwakeWindow               Dot11InformationElementID = 157
	Dot11InformationElementIDMultiBand                 Dot11InformationElementID = 158
	Dot11InformationElementIDADDBAExtension            Dot11InformationElementID = 159
	Dot11InformationElementIDNEXTPCPList               Dot11InformationElementID = 160
	Dot11InformationElementIDPCPHandover               Dot11InformationElementID = 161
	Dot11InformationElementIDDMGLinkMargin             Dot11InformationElementID = 162
	Dot11InformationElementIDSwitchingStream           Dot11InformationElementID = 163
	Dot11InformationElementIDSessionTransmission       Dot11InformationElementID = 164
	Dot11InformationElementIDDynamicTonePairReport     Dot11InformationElementID = 165
	Dot11InformationElementIDClusterReport             Dot11InformationElementID = 166
	Dot11InformationElementIDRelayCapabilities         Dot11InformationElementID = 167
	Dot11InformationElementIDRelayTransferParameter    Dot11InformationElementID = 168
	Dot11InformationElementIDBeamlinkMaintenance       Dot11InformationElementID = 169
	Dot11InformationElementIDMultipleMacSublayers      Dot11InformationElementID = 170
	Dot11InformationElementIDUPID                      Dot11InformationElementID = 171
	Dot11InformationElementIDDMGLinkAdaptionAck        Dot11InformationElementID = 172
	Dot11InformationElementIDSymbolProprietary         Dot11InformationElementID = 173
	Dot11InformationElementIDMCCAOPAdvertOverview      Dot11InformationElementID = 174
	Dot11InformationElementIDQuietPeriodRequest        Dot11InformationElementID = 175
	Dot11InformationElementIDQuietPeriodResponse       Dot11InformationElementID = 177
	Dot11InformationElementIDECPACPolicy               Dot11InformationElementID = 182
	Dot11InformationElementIDClusterTimeOffset         Dot11InformationElementID = 183
	Dot11InformationElementIDAntennaSectorID           Dot11InformationElementID = 190
	Dot11InformationElementIDVHTCapabilities           Dot11InformationElementID = 191
	Dot11InformationElementIDVHTOperation              Dot11InformationElementID = 192
	Dot11InformationElementIDExtendedBSSLoad           Dot11InformationElementID = 193
	Dot11InformationElementIDWideBWChannelSwitch       Dot11InformationElementID = 194
	Dot11InformationElementIDVHTTxPowerEnvelope        Dot11InformationElementID = 195
	Dot11InformationElementIDChannelSwitchWrapper      Dot11InformationElementID = 196
	Dot11InformationElementIDOperatingModeNotification Dot11InformationElementID = 199
	Dot11InformationElementIDUPSIM                     Dot11InformationElementID = 200
	Dot11InformationElementIDReducedNeighborReport     Dot11InformationElementID = 201
	Dot11InformationElementIDTVHTOperation             Dot11InformationElementID = 202
	Dot11InformationElementIDDeviceLocation            Dot11InformationElementID = 204
	Dot11InformationElementIDWhiteSpaceMap             Dot11InformationElementID = 205
	Dot11InformationElementIDFineTuningMeasureParams   Dot11InformationElementID = 206
	Dot11InformationElementIDVendor                    Dot11InformationElementID = 221
)
func (Dot11InformationElementID) String ¶
added in v1.1.4
func (a Dot11InformationElementID) String() string
String provides a human readable string for Dot11InformationElementID. This string is possibly subject to change over time; if you're storing this persistently, you should probably store the Dot11InformationElementID value, not its string.

type Dot11LinkAdapationControl ¶
added in v1.1.15
type Dot11LinkAdapationControl struct {
	TRQ  bool
	MRQ  bool
	MSI  uint8
	MFSI uint8
	ASEL *Dot11ASEL
	MFB  *uint8
}
type Dot11Mgmt ¶
added in v1.1.4
type Dot11Mgmt struct {
	BaseLayer
}
Dot11Mgmt is a base for all IEEE 802.11 management layers.

func (*Dot11Mgmt) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11Mgmt) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11Mgmt) NextLayerType ¶
added in v1.1.4
func (m *Dot11Mgmt) NextLayerType() gopacket.LayerType
type Dot11MgmtATIM ¶
added in v1.1.4
type Dot11MgmtATIM struct {
	Dot11Mgmt
}
func (*Dot11MgmtATIM) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtATIM) CanDecode() gopacket.LayerClass
func (*Dot11MgmtATIM) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtATIM) LayerType() gopacket.LayerType
type Dot11MgmtAction ¶
added in v1.1.4
type Dot11MgmtAction struct {
	Dot11Mgmt
}
func (*Dot11MgmtAction) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtAction) CanDecode() gopacket.LayerClass
func (*Dot11MgmtAction) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtAction) LayerType() gopacket.LayerType
type Dot11MgmtActionNoAck ¶
added in v1.1.4
type Dot11MgmtActionNoAck struct {
	Dot11Mgmt
}
func (*Dot11MgmtActionNoAck) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtActionNoAck) CanDecode() gopacket.LayerClass
func (*Dot11MgmtActionNoAck) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtActionNoAck) LayerType() gopacket.LayerType
type Dot11MgmtArubaWLAN ¶
added in v1.1.4
type Dot11MgmtArubaWLAN struct {
	Dot11Mgmt
}
func (*Dot11MgmtArubaWLAN) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtArubaWLAN) CanDecode() gopacket.LayerClass
func (*Dot11MgmtArubaWLAN) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtArubaWLAN) LayerType() gopacket.LayerType
type Dot11MgmtAssociationReq ¶
added in v1.1.4
type Dot11MgmtAssociationReq struct {
	Dot11Mgmt
	CapabilityInfo uint16
	ListenInterval uint16
}
func (*Dot11MgmtAssociationReq) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtAssociationReq) CanDecode() gopacket.LayerClass
func (*Dot11MgmtAssociationReq) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11MgmtAssociationReq) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11MgmtAssociationReq) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtAssociationReq) LayerType() gopacket.LayerType
func (*Dot11MgmtAssociationReq) NextLayerType ¶
added in v1.1.4
func (m *Dot11MgmtAssociationReq) NextLayerType() gopacket.LayerType
func (Dot11MgmtAssociationReq) SerializeTo ¶
added in v1.1.12
func (m Dot11MgmtAssociationReq) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11MgmtAssociationResp ¶
added in v1.1.4
type Dot11MgmtAssociationResp struct {
	Dot11Mgmt
	CapabilityInfo uint16
	Status         Dot11Status
	AID            uint16
}
func (*Dot11MgmtAssociationResp) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtAssociationResp) CanDecode() gopacket.LayerClass
func (*Dot11MgmtAssociationResp) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11MgmtAssociationResp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11MgmtAssociationResp) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtAssociationResp) LayerType() gopacket.LayerType
func (*Dot11MgmtAssociationResp) NextLayerType ¶
added in v1.1.4
func (m *Dot11MgmtAssociationResp) NextLayerType() gopacket.LayerType
func (Dot11MgmtAssociationResp) SerializeTo ¶
added in v1.1.12
func (m Dot11MgmtAssociationResp) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11MgmtAuthentication ¶
added in v1.1.4
type Dot11MgmtAuthentication struct {
	Dot11Mgmt
	Algorithm Dot11Algorithm
	Sequence  uint16
	Status    Dot11Status
}
func (*Dot11MgmtAuthentication) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtAuthentication) CanDecode() gopacket.LayerClass
func (*Dot11MgmtAuthentication) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11MgmtAuthentication) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11MgmtAuthentication) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtAuthentication) LayerType() gopacket.LayerType
func (*Dot11MgmtAuthentication) NextLayerType ¶
added in v1.1.4
func (m *Dot11MgmtAuthentication) NextLayerType() gopacket.LayerType
func (Dot11MgmtAuthentication) SerializeTo ¶
added in v1.1.12
func (m Dot11MgmtAuthentication) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11MgmtBeacon ¶
added in v1.1.4
type Dot11MgmtBeacon struct {
	Dot11Mgmt
	Timestamp uint64
	Interval  uint16
	Flags     uint16
}
func (*Dot11MgmtBeacon) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtBeacon) CanDecode() gopacket.LayerClass
func (*Dot11MgmtBeacon) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11MgmtBeacon) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11MgmtBeacon) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtBeacon) LayerType() gopacket.LayerType
func (*Dot11MgmtBeacon) NextLayerType ¶
added in v1.1.4
func (m *Dot11MgmtBeacon) NextLayerType() gopacket.LayerType
func (Dot11MgmtBeacon) SerializeTo ¶
added in v1.1.12
func (m Dot11MgmtBeacon) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11MgmtDeauthentication ¶
added in v1.1.4
type Dot11MgmtDeauthentication struct {
	Dot11Mgmt
	Reason Dot11Reason
}
func (*Dot11MgmtDeauthentication) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtDeauthentication) CanDecode() gopacket.LayerClass
func (*Dot11MgmtDeauthentication) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11MgmtDeauthentication) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11MgmtDeauthentication) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtDeauthentication) LayerType() gopacket.LayerType
func (Dot11MgmtDeauthentication) SerializeTo ¶
added in v1.1.12
func (m Dot11MgmtDeauthentication) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11MgmtDisassociation ¶
added in v1.1.4
type Dot11MgmtDisassociation struct {
	Dot11Mgmt
	Reason Dot11Reason
}
func (*Dot11MgmtDisassociation) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtDisassociation) CanDecode() gopacket.LayerClass
func (*Dot11MgmtDisassociation) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11MgmtDisassociation) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11MgmtDisassociation) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtDisassociation) LayerType() gopacket.LayerType
func (Dot11MgmtDisassociation) SerializeTo ¶
added in v1.1.12
func (m Dot11MgmtDisassociation) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11MgmtMeasurementPilot ¶
added in v1.1.4
type Dot11MgmtMeasurementPilot struct {
	Dot11Mgmt
}
func (*Dot11MgmtMeasurementPilot) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtMeasurementPilot) CanDecode() gopacket.LayerClass
func (*Dot11MgmtMeasurementPilot) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtMeasurementPilot) LayerType() gopacket.LayerType
type Dot11MgmtProbeReq ¶
added in v1.1.4
type Dot11MgmtProbeReq struct {
	Dot11Mgmt
}
func (*Dot11MgmtProbeReq) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtProbeReq) CanDecode() gopacket.LayerClass
func (*Dot11MgmtProbeReq) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtProbeReq) LayerType() gopacket.LayerType
func (*Dot11MgmtProbeReq) NextLayerType ¶
added in v1.1.4
func (m *Dot11MgmtProbeReq) NextLayerType() gopacket.LayerType
type Dot11MgmtProbeResp ¶
added in v1.1.4
type Dot11MgmtProbeResp struct {
	Dot11Mgmt
	Timestamp uint64
	Interval  uint16
	Flags     uint16
}
func (*Dot11MgmtProbeResp) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtProbeResp) CanDecode() gopacket.LayerClass
func (*Dot11MgmtProbeResp) DecodeFromBytes ¶
added in v1.1.12
func (m *Dot11MgmtProbeResp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11MgmtProbeResp) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtProbeResp) LayerType() gopacket.LayerType
func (*Dot11MgmtProbeResp) NextLayerType ¶
added in v1.1.4
func (m *Dot11MgmtProbeResp) NextLayerType() gopacket.LayerType
func (Dot11MgmtProbeResp) SerializeTo ¶
added in v1.1.12
func (m Dot11MgmtProbeResp) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11MgmtReassociationReq ¶
added in v1.1.4
type Dot11MgmtReassociationReq struct {
	Dot11Mgmt
	CapabilityInfo   uint16
	ListenInterval   uint16
	CurrentApAddress net.HardwareAddr
}
func (*Dot11MgmtReassociationReq) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtReassociationReq) CanDecode() gopacket.LayerClass
func (*Dot11MgmtReassociationReq) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11MgmtReassociationReq) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11MgmtReassociationReq) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtReassociationReq) LayerType() gopacket.LayerType
func (*Dot11MgmtReassociationReq) NextLayerType ¶
added in v1.1.4
func (m *Dot11MgmtReassociationReq) NextLayerType() gopacket.LayerType
func (Dot11MgmtReassociationReq) SerializeTo ¶
added in v1.1.12
func (m Dot11MgmtReassociationReq) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error
type Dot11MgmtReassociationResp ¶
added in v1.1.4
type Dot11MgmtReassociationResp struct {
	Dot11Mgmt
}
func (*Dot11MgmtReassociationResp) CanDecode ¶
added in v1.1.4
func (m *Dot11MgmtReassociationResp) CanDecode() gopacket.LayerClass
func (*Dot11MgmtReassociationResp) LayerType ¶
added in v1.1.4
func (m *Dot11MgmtReassociationResp) LayerType() gopacket.LayerType
func (*Dot11MgmtReassociationResp) NextLayerType ¶
added in v1.1.4
func (m *Dot11MgmtReassociationResp) NextLayerType() gopacket.LayerType
type Dot11QOS ¶
added in v1.1.15
type Dot11QOS struct {
	TID       uint8 /* Traffic IDentifier */
	EOSP      bool  /* End of service period */
	AckPolicy Dot11AckPolicy
	TXOP      uint8
}
type Dot11Reason ¶
added in v1.1.4
type Dot11Reason uint16
const (
	Dot11ReasonReserved          Dot11Reason = 1
	Dot11ReasonUnspecified       Dot11Reason = 2
	Dot11ReasonAuthExpired       Dot11Reason = 3
	Dot11ReasonDeauthStLeaving   Dot11Reason = 4
	Dot11ReasonInactivity        Dot11Reason = 5
	Dot11ReasonApFull            Dot11Reason = 6
	Dot11ReasonClass2FromNonAuth Dot11Reason = 7
	Dot11ReasonClass3FromNonAss  Dot11Reason = 8
	Dot11ReasonDisasStLeaving    Dot11Reason = 9
	Dot11ReasonStNotAuth         Dot11Reason = 10
)
func (Dot11Reason) String ¶
added in v1.1.4
func (a Dot11Reason) String() string
String provides a human readable string for Dot11Reason. This string is possibly subject to change over time; if you're storing this persistently, you should probably store the Dot11Reason value, not its string.

type Dot11Status ¶
added in v1.1.4
type Dot11Status uint16
const (
	Dot11StatusSuccess                      Dot11Status = 0
	Dot11StatusFailure                      Dot11Status = 1  // Unspecified failure
	Dot11StatusCannotSupportAllCapabilities Dot11Status = 10 // Cannot support all requested capabilities in the Capability Information field
	Dot11StatusInabilityExistsAssociation   Dot11Status = 11 // Reassociation denied due to inability to confirm that association exists
	Dot11StatusAssociationDenied            Dot11Status = 12 // Association denied due to reason outside the scope of this standard
	Dot11StatusAlgorithmUnsupported         Dot11Status = 13 // Responding station does not support the specified authentication algorithm
	Dot11StatusOufOfExpectedSequence        Dot11Status = 14 // Received an Authentication frame with authentication transaction sequence number out of expected sequence
	Dot11StatusChallengeFailure             Dot11Status = 15 // Authentication rejected because of challenge failure
	Dot11StatusTimeout                      Dot11Status = 16 // Authentication rejected due to timeout waiting for next frame in sequence
	Dot11StatusAPUnableToHandle             Dot11Status = 17 // Association denied because AP is unable to handle additional associated stations
	Dot11StatusRateUnsupported              Dot11Status = 18 // Association denied due to requesting station not supporting all of the data rates in the BSSBasicRateSet parameter
)
func (Dot11Status) String ¶
added in v1.1.4
func (a Dot11Status) String() string
String provides a human readable string for Dot11Status. This string is possibly subject to change over time; if you're storing this persistently, you should probably store the Dot11Status value, not its string.

type Dot11Type ¶
added in v1.1.4
type Dot11Type uint8
Dot11Type is a combination of IEEE 802.11 frame's Type and Subtype fields. By combining these two fields together into a single type, we're able to provide a String function that correctly displays the subtype given the top-level type.

If you just care about the top-level type, use the MainType function.

const (
	Dot11TypeMgmt     Dot11Type = 0x00
	Dot11TypeCtrl     Dot11Type = 0x01
	Dot11TypeData     Dot11Type = 0x02
	Dot11TypeReserved Dot11Type = 0x03

	// Management
	Dot11TypeMgmtAssociationReq    Dot11Type = 0x00
	Dot11TypeMgmtAssociationResp   Dot11Type = 0x04
	Dot11TypeMgmtReassociationReq  Dot11Type = 0x08
	Dot11TypeMgmtReassociationResp Dot11Type = 0x0c
	Dot11TypeMgmtProbeReq          Dot11Type = 0x10
	Dot11TypeMgmtProbeResp         Dot11Type = 0x14
	Dot11TypeMgmtMeasurementPilot  Dot11Type = 0x18
	Dot11TypeMgmtBeacon            Dot11Type = 0x20
	Dot11TypeMgmtATIM              Dot11Type = 0x24
	Dot11TypeMgmtDisassociation    Dot11Type = 0x28
	Dot11TypeMgmtAuthentication    Dot11Type = 0x2c
	Dot11TypeMgmtDeauthentication  Dot11Type = 0x30
	Dot11TypeMgmtAction            Dot11Type = 0x34
	Dot11TypeMgmtActionNoAck       Dot11Type = 0x38

	// Control
	Dot11TypeCtrlWrapper       Dot11Type = 0x1d
	Dot11TypeCtrlBlockAckReq   Dot11Type = 0x21
	Dot11TypeCtrlBlockAck      Dot11Type = 0x25
	Dot11TypeCtrlPowersavePoll Dot11Type = 0x29
	Dot11TypeCtrlRTS           Dot11Type = 0x2d
	Dot11TypeCtrlCTS           Dot11Type = 0x31
	Dot11TypeCtrlAck           Dot11Type = 0x35
	Dot11TypeCtrlCFEnd         Dot11Type = 0x39
	Dot11TypeCtrlCFEndAck      Dot11Type = 0x3d

	// Data
	Dot11TypeDataCFAck              Dot11Type = 0x06
	Dot11TypeDataCFPoll             Dot11Type = 0x0a
	Dot11TypeDataCFAckPoll          Dot11Type = 0x0e
	Dot11TypeDataNull               Dot11Type = 0x12
	Dot11TypeDataCFAckNoData        Dot11Type = 0x16
	Dot11TypeDataCFPollNoData       Dot11Type = 0x1a
	Dot11TypeDataCFAckPollNoData    Dot11Type = 0x1e
	Dot11TypeDataQOSData            Dot11Type = 0x22
	Dot11TypeDataQOSDataCFAck       Dot11Type = 0x26
	Dot11TypeDataQOSDataCFPoll      Dot11Type = 0x2a
	Dot11TypeDataQOSDataCFAckPoll   Dot11Type = 0x2e
	Dot11TypeDataQOSNull            Dot11Type = 0x32
	Dot11TypeDataQOSCFPollNoData    Dot11Type = 0x3a
	Dot11TypeDataQOSCFAckPollNoData Dot11Type = 0x3e
)
func (Dot11Type) Decode ¶
added in v1.1.4
func (a Dot11Type) Decode(data []byte, p gopacket.PacketBuilder) error
Decoder calls Dot11TypeMetadata.DecodeWith's decoder.

func (Dot11Type) LayerType ¶
added in v1.1.4
func (a Dot11Type) LayerType() gopacket.LayerType
LayerType returns Dot11TypeMetadata.LayerType.

func (Dot11Type) MainType ¶
added in v1.1.4
func (d Dot11Type) MainType() Dot11Type
MainType strips the subtype information from the given type, returning just the overarching type (Mgmt, Ctrl, Data, Reserved).

func (Dot11Type) QOS ¶
added in v1.1.15
func (d Dot11Type) QOS() bool
func (Dot11Type) String ¶
added in v1.1.4
func (a Dot11Type) String() string
String returns Dot11TypeMetadata.Name.

type Dot11WEP ¶
added in v1.1.4
type Dot11WEP struct {
	BaseLayer
}
Dot11WEP contains WEP encrpted IEEE 802.11 data.

func (*Dot11WEP) CanDecode ¶
added in v1.1.4
func (m *Dot11WEP) CanDecode() gopacket.LayerClass
func (*Dot11WEP) DecodeFromBytes ¶
added in v1.1.4
func (m *Dot11WEP) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error
func (*Dot11WEP) LayerType ¶
added in v1.1.4
func (m *Dot11WEP) LayerType() gopacket.LayerType
func (*Dot11WEP) NextLayerType ¶
added in v1.1.4
func (m *Dot11WEP) NextLayerType() gopacket.LayerType