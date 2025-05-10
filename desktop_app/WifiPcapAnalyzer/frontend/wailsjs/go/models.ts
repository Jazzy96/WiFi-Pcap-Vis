export namespace config {
	
	export class AppConfig {
	    grpc_server_address: string;
	    websocket_address: string;
	    log_file: string;
	    log_level: string;
	    min_bss_creation_rssi: number;
	    tshark_path: string;
	
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
	
	    static createFrom(source: any = {}) {
	        return new HECapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported_mcs_set = source["supported_mcs_set"];
	    }
	}
	export class VHTCapabilities {
	    supported_mcs_set: Record<string, number[]>;
	    short_gi_80mhz: boolean;
	    short_gi_160mhz: boolean;
	    channel_width_80mhz: boolean;
	    channel_width_160mhz: boolean;
	    channel_width_80plus80mhz: boolean;
	
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
	    }
	}
	export class HTCapabilities {
	    supported_mcs_set: number[];
	    short_gi_20mhz: boolean;
	    short_gi_40mhz: boolean;
	    channel_width_40mhz: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HTCapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported_mcs_set = source["supported_mcs_set"];
	        this.short_gi_20mhz = source["short_gi_20mhz"];
	        this.short_gi_40mhz = source["short_gi_40mhz"];
	        this.channel_width_40mhz = source["channel_width_40mhz"];
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

