export namespace main {
	
	export class ConfigDTO {
	    api_url: string;
	    api_secret: string;
	    check_interval: number;
	    target_groups: string[];
	    dedicated_test_group: string;
	    test_urls: string[];
	    test_timeout: number;
	    tolerance_ms: number;
	    cleanup_days: number;
	    max_concurrent: number;
	    web_port: number;
	    clash_proxy_url: string;
	    max_backoff_cycles: number;
	    enable_browser_test: boolean;
	    browser_test_urls: string[];
	
	    static createFrom(source: any = {}) {
	        return new ConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_url = source["api_url"];
	        this.api_secret = source["api_secret"];
	        this.check_interval = source["check_interval"];
	        this.target_groups = source["target_groups"];
	        this.dedicated_test_group = source["dedicated_test_group"];
	        this.test_urls = source["test_urls"];
	        this.test_timeout = source["test_timeout"];
	        this.tolerance_ms = source["tolerance_ms"];
	        this.cleanup_days = source["cleanup_days"];
	        this.max_concurrent = source["max_concurrent"];
	        this.web_port = source["web_port"];
	        this.clash_proxy_url = source["clash_proxy_url"];
	        this.max_backoff_cycles = source["max_backoff_cycles"];
	        this.enable_browser_test = source["enable_browser_test"];
	        this.browser_test_urls = source["browser_test_urls"];
	    }
	}
	export class GroupFilter {
	    keyword_regex: string;
	
	    static createFrom(source: any = {}) {
	        return new GroupFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyword_regex = source["keyword_regex"];
	    }
	}
	export class GroupStatus {
	    name: string;
	    now: string;
	    provider: string;
	    all_count: number;
	    all_nodes: string[];
	    locked: boolean;
	    filter: GroupFilter;
	
	    static createFrom(source: any = {}) {
	        return new GroupStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.now = source["now"];
	        this.provider = source["provider"];
	        this.all_count = source["all_count"];
	        this.all_nodes = source["all_nodes"];
	        this.locked = source["locked"];
	        this.filter = this.convertValues(source["filter"], GroupFilter);
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
	export class StatNode {
	    Name: string;
	    AvgDelay: number;
	    Jitter: number;
	    Score: number;
	    provider: string;
	    highest_in_groups: string[];
	    backoff_remaining: number;
	    browser_backoff_remaining: Record<string, number>;
	    is_dead: boolean;
	
	    static createFrom(source: any = {}) {
	        return new StatNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.AvgDelay = source["AvgDelay"];
	        this.Jitter = source["Jitter"];
	        this.Score = source["Score"];
	        this.provider = source["provider"];
	        this.highest_in_groups = source["highest_in_groups"];
	        this.backoff_remaining = source["backoff_remaining"];
	        this.browser_backoff_remaining = source["browser_backoff_remaining"];
	        this.is_dead = source["is_dead"];
	    }
	}
	export class WebLogEntry {
	    level: string;
	    message: string;
	    time: string;
	
	    static createFrom(source: any = {}) {
	        return new WebLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.time = source["time"];
	    }
	}

}

