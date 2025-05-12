export namespace config {
	
	export class LoggingConfig {
	    level: string;
	    file?: string;
	    console?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LoggingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.file = source["file"];
	        this.console = source["console"];
	    }
	}
	export class AppConfig {
	    grpc_server_address: string;
	    websocket_address: string;
	    log_file: string;
	    log_level: string;
	    min_bss_creation_rssi: number;
	    tshark_path: string;
	    logging?: LoggingConfig;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.grpc_server_address = source["grpc_server_address"];
	        this.websocket_address = source["websocket_address"];
	        this.log_file = source["log_file"];
	        this.log_level = source["log_level"];
	        this.min_bss_creation_rssi = source["min_bss_creation_rssi"];
	        this.tshark_path = source["tshark_path"];
	        this.logging = this.convertValues(source["logging"], LoggingConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace state_manager {
	
	export class STAInfo {
	    mac_address: string;
	    associated_bssid?: string;
	    signal_strength: number;
	    last_seen: number;
	    ht_capabilities?: HTCapabilities;
	    vht_capabilities?: VHTCapabilities;
	    he_capabilities?: HECapabilities;
	    channel_utilization: number;
	    uplink_throughput: number;
	    downlink_throughput: number;
	    historical_channel_utilization: number[];
	    historical_uplink_throughput: number[];
	    historical_downlink_throughput: number[];
	    rx_bytes: number;
	    tx_bytes: number;
	    rx_packets: number;
	    tx_packets: number;
	    rx_retries: number;
	    tx_retries: number;
	    AccumulatedNavMicroseconds: number;
	    util: number;
	    thrpt: number;
	    bitrate: number;
	
	    static createFrom(source: any = {}) {
	        return new STAInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mac_address = source["mac_address"];
	        this.associated_bssid = source["associated_bssid"];
	        this.signal_strength = source["signal_strength"];
	        this.last_seen = source["last_seen"];
	        this.ht_capabilities = this.convertValues(source["ht_capabilities"], HTCapabilities);
	        this.vht_capabilities = this.convertValues(source["vht_capabilities"], VHTCapabilities);
	        this.he_capabilities = this.convertValues(source["he_capabilities"], HECapabilities);
	        this.channel_utilization = source["channel_utilization"];
	        this.uplink_throughput = source["uplink_throughput"];
	        this.downlink_throughput = source["downlink_throughput"];
	        this.historical_channel_utilization = source["historical_channel_utilization"];
	        this.historical_uplink_throughput = source["historical_uplink_throughput"];
	        this.historical_downlink_throughput = source["historical_downlink_throughput"];
	        this.rx_bytes = source["rx_bytes"];
	        this.tx_bytes = source["tx_bytes"];
	        this.rx_packets = source["rx_packets"];
	        this.tx_packets = source["tx_packets"];
	        this.rx_retries = source["rx_retries"];
	        this.tx_retries = source["tx_retries"];
	        this.AccumulatedNavMicroseconds = source["AccumulatedNavMicroseconds"];
	        this.util = source["util"];
	        this.thrpt = source["thrpt"];
	        this.bitrate = source["bitrate"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HECapabilities {
	    supported_mcs_set: Record<string, number[]>;
	    bss_color: string;
	    htc_he_support: boolean;
	    twt_requester_support: boolean;
	    twt_responder_support: boolean;
	    su_beamformer: boolean;
	    su_beamformee: boolean;
	    channel_width_160mhz: boolean;
	    channel_width_80plus80mhz: boolean;
	    channel_width_40_80mhz_in_5g: boolean;
	    max_mcs_for_1_ss: number;
	    max_mcs_for_2_ss: number;
	    max_mcs_for_3_ss: number;
	    max_mcs_for_4_ss: number;
	    rx_he_mcs_map: number;
	    tx_he_mcs_map: number;
	
	    static createFrom(source: any = {}) {
	        return new HECapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported_mcs_set = source["supported_mcs_set"];
	        this.bss_color = source["bss_color"];
	        this.htc_he_support = source["htc_he_support"];
	        this.twt_requester_support = source["twt_requester_support"];
	        this.twt_responder_support = source["twt_responder_support"];
	        this.su_beamformer = source["su_beamformer"];
	        this.su_beamformee = source["su_beamformee"];
	        this.channel_width_160mhz = source["channel_width_160mhz"];
	        this.channel_width_80plus80mhz = source["channel_width_80plus80mhz"];
	        this.channel_width_40_80mhz_in_5g = source["channel_width_40_80mhz_in_5g"];
	        this.max_mcs_for_1_ss = source["max_mcs_for_1_ss"];
	        this.max_mcs_for_2_ss = source["max_mcs_for_2_ss"];
	        this.max_mcs_for_3_ss = source["max_mcs_for_3_ss"];
	        this.max_mcs_for_4_ss = source["max_mcs_for_4_ss"];
	        this.rx_he_mcs_map = source["rx_he_mcs_map"];
	        this.tx_he_mcs_map = source["tx_he_mcs_map"];
	    }
	}
	export class VHTCapabilities {
	    supported_mcs_set: Record<string, number[]>;
	    short_gi_80mhz: boolean;
	    short_gi_160mhz: boolean;
	    channel_width_80mhz: boolean;
	    channel_width_160mhz: boolean;
	    channel_width_80plus80mhz: boolean;
	    max_mpdu_length: number;
	    rx_ldpc: boolean;
	    tx_stbc: boolean;
	    rx_stbc: number;
	    su_beamformer_capable: boolean;
	    su_beamformee_capable: boolean;
	    mu_beamformer_capable: boolean;
	    mu_beamformee_capable: boolean;
	    beamformee_sts: number;
	    sounding_dimensions: number;
	    max_ampdu_length_exp: number;
	    rx_pattern_consistency: boolean;
	    tx_pattern_consistency: boolean;
	    rx_mcs_map: number;
	    tx_mcs_map: number;
	    rx_highest_long_gi_rate: number;
	    tx_highest_long_gi_rate: number;
	    vht_htc_capability: boolean;
	    vht_txop_ps_capability: boolean;
	    channel_center_0: number;
	    channel_center_1: number;
	    supported_channel_width_set: number;
	
	    static createFrom(source: any = {}) {
	        return new VHTCapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported_mcs_set = source["supported_mcs_set"];
	        this.short_gi_80mhz = source["short_gi_80mhz"];
	        this.short_gi_160mhz = source["short_gi_160mhz"];
	        this.channel_width_80mhz = source["channel_width_80mhz"];
	        this.channel_width_160mhz = source["channel_width_160mhz"];
	        this.channel_width_80plus80mhz = source["channel_width_80plus80mhz"];
	        this.max_mpdu_length = source["max_mpdu_length"];
	        this.rx_ldpc = source["rx_ldpc"];
	        this.tx_stbc = source["tx_stbc"];
	        this.rx_stbc = source["rx_stbc"];
	        this.su_beamformer_capable = source["su_beamformer_capable"];
	        this.su_beamformee_capable = source["su_beamformee_capable"];
	        this.mu_beamformer_capable = source["mu_beamformer_capable"];
	        this.mu_beamformee_capable = source["mu_beamformee_capable"];
	        this.beamformee_sts = source["beamformee_sts"];
	        this.sounding_dimensions = source["sounding_dimensions"];
	        this.max_ampdu_length_exp = source["max_ampdu_length_exp"];
	        this.rx_pattern_consistency = source["rx_pattern_consistency"];
	        this.tx_pattern_consistency = source["tx_pattern_consistency"];
	        this.rx_mcs_map = source["rx_mcs_map"];
	        this.tx_mcs_map = source["tx_mcs_map"];
	        this.rx_highest_long_gi_rate = source["rx_highest_long_gi_rate"];
	        this.tx_highest_long_gi_rate = source["tx_highest_long_gi_rate"];
	        this.vht_htc_capability = source["vht_htc_capability"];
	        this.vht_txop_ps_capability = source["vht_txop_ps_capability"];
	        this.channel_center_0 = source["channel_center_0"];
	        this.channel_center_1 = source["channel_center_1"];
	        this.supported_channel_width_set = source["supported_channel_width_set"];
	    }
	}
	export class HTCapabilities {
	    supported_mcs_set: number[];
	    short_gi_20mhz: boolean;
	    short_gi_40mhz: boolean;
	    channel_width_40mhz: boolean;
	    ldpc_coding: boolean;
	    "40mhz_intolerant": boolean;
	    tx_stbc: boolean;
	    rx_stbc: number;
	    max_amsdu_length: number;
	    dsss_cck_mode_40mhz: boolean;
	    delayed_block_ack: boolean;
	    max_ampdu_length: number;
	    primary_channel: number;
	
	    static createFrom(source: any = {}) {
	        return new HTCapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported_mcs_set = source["supported_mcs_set"];
	        this.short_gi_20mhz = source["short_gi_20mhz"];
	        this.short_gi_40mhz = source["short_gi_40mhz"];
	        this.channel_width_40mhz = source["channel_width_40mhz"];
	        this.ldpc_coding = source["ldpc_coding"];
	        this["40mhz_intolerant"] = source["40mhz_intolerant"];
	        this.tx_stbc = source["tx_stbc"];
	        this.rx_stbc = source["rx_stbc"];
	        this.max_amsdu_length = source["max_amsdu_length"];
	        this.dsss_cck_mode_40mhz = source["dsss_cck_mode_40mhz"];
	        this.delayed_block_ack = source["delayed_block_ack"];
	        this.max_ampdu_length = source["max_ampdu_length"];
	        this.primary_channel = source["primary_channel"];
	    }
	}
	export class BSSInfo {
	    bssid: string;
	    ssid: string;
	    channel: number;
	    bandwidth: string;
	    security: string;
	    signal_strength: number;
	    last_seen: number;
	    ht_capabilities?: HTCapabilities;
	    vht_capabilities?: VHTCapabilities;
	    he_capabilities?: HECapabilities;
	    associated_stas: Record<string, STAInfo>;
	    channel_utilization: number;
	    throughput: number;
	    historical_channel_utilization: number[];
	    historical_throughput: number[];
	    AccumulatedNavMicroseconds: number;
	    util: number;
	    thrpt: number;
	
	    static createFrom(source: any = {}) {
	        return new BSSInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bssid = source["bssid"];
	        this.ssid = source["ssid"];
	        this.channel = source["channel"];
	        this.bandwidth = source["bandwidth"];
	        this.security = source["security"];
	        this.signal_strength = source["signal_strength"];
	        this.last_seen = source["last_seen"];
	        this.ht_capabilities = this.convertValues(source["ht_capabilities"], HTCapabilities);
	        this.vht_capabilities = this.convertValues(source["vht_capabilities"], VHTCapabilities);
	        this.he_capabilities = this.convertValues(source["he_capabilities"], HECapabilities);
	        this.associated_stas = this.convertValues(source["associated_stas"], STAInfo, true);
	        this.channel_utilization = source["channel_utilization"];
	        this.throughput = source["throughput"];
	        this.historical_channel_utilization = source["historical_channel_utilization"];
	        this.historical_throughput = source["historical_throughput"];
	        this.AccumulatedNavMicroseconds = source["AccumulatedNavMicroseconds"];
	        this.util = source["util"];
	        this.thrpt = source["thrpt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class Snapshot {
	    bsss: BSSInfo[];
	    stas: STAInfo[];
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bsss = this.convertValues(source["bsss"], BSSInfo);
	        this.stas = this.convertValues(source["stas"], STAInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

