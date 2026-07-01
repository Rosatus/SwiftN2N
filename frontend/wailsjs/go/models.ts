export namespace edge {
	
	export class Config {
	    edgePath: string;
	    allowCustomEdgePath: boolean;
	    supernodes: string[];
	    community: string;
	    address: string;
	    key: string;
	    cipher: string;
	    headerEncryption: boolean;
	    mac: string;
	    deviceName: string;
	    mtu: number;
	    localPort: string;
	    managementPort: number;
	    verbose: number;
	    authUsername: string;
	    authPassword: string;
	    federationPublicKey: string;
	    supernodeOnly: string;
	    compression: string;
	    acceptMulticast: boolean;
	    enableRouting: boolean;
	    routes: string[];
	    trafficRules: string[];
	    windowsMetric: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.edgePath = source["edgePath"];
	        this.allowCustomEdgePath = source["allowCustomEdgePath"];
	        this.supernodes = source["supernodes"];
	        this.community = source["community"];
	        this.address = source["address"];
	        this.key = source["key"];
	        this.cipher = source["cipher"];
	        this.headerEncryption = source["headerEncryption"];
	        this.mac = source["mac"];
	        this.deviceName = source["deviceName"];
	        this.mtu = source["mtu"];
	        this.localPort = source["localPort"];
	        this.managementPort = source["managementPort"];
	        this.verbose = source["verbose"];
	        this.authUsername = source["authUsername"];
	        this.authPassword = source["authPassword"];
	        this.federationPublicKey = source["federationPublicKey"];
	        this.supernodeOnly = source["supernodeOnly"];
	        this.compression = source["compression"];
	        this.acceptMulticast = source["acceptMulticast"];
	        this.enableRouting = source["enableRouting"];
	        this.routes = source["routes"];
	        this.trafficRules = source["trafficRules"];
	        this.windowsMetric = source["windowsMetric"];
	    }
	}
	export class EnvironmentStatus {
	    os: string;
	    arch: string;
	    elevated: boolean;
	    helperAvailable: boolean;
	    edgeFound: boolean;
	    edgePath: string;
	    edgeVersion: string;
	    missing: string[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new EnvironmentStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.elevated = source["elevated"];
	        this.helperAvailable = source["helperAvailable"];
	        this.edgeFound = source["edgeFound"];
	        this.edgePath = source["edgePath"];
	        this.edgeVersion = source["edgeVersion"];
	        this.missing = source["missing"];
	        this.message = source["message"];
	    }
	}
	export class Status {
	    state: string;
	    message: string;
	    pid: number;
	    edgePath: string;
	    startedAt: string;
	    updatedAt: string;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.message = source["message"];
	        this.pid = source["pid"];
	        this.edgePath = source["edgePath"];
	        this.startedAt = source["startedAt"];
	        this.updatedAt = source["updatedAt"];
	        this.lastError = source["lastError"];
	    }
	}

}
